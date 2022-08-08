package gont

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	"github.com/stv0g/gont/internal/pprque"
	"go.uber.org/zap"
	"golang.org/x/net/bpf"
)

type captureInterface struct {
	pcapgo.NgInterface
	Index  int
	Handle *pcapgo.EthernetHandle

	logger *zap.Logger
}

type filenameTemplate struct {
	Interface string
	Host      string
	Network   string
}

type CaptureFilterInterfaceFunc func(i *Interface) bool

type CaptureOption interface {
	Apply(c *Capture)
}

type Capture struct {
	CaptureLength int
	Promisc       bool
	Filter        CaptureFilterInterfaceFunc
	BPFilter      string
	Comment       string

	File     *os.File
	Filename string

	writer       *pcapgo.NgWriter
	filterInstrs []bpf.RawInstruction

	stop chan any

	queue *pprque.PacketPriorityQueue

	interfaces []*captureInterface

	logger *zap.Logger
}

func (c *Capture) Apply(i *Interface) {
	i.Captures = append(i.Captures, c)
}

func NewCapture() *Capture {
	return &Capture{
		// Default options
		CaptureLength: 1600,
		Promisc:       false,
		Filename:      "capture.pcapng",

		stop:   make(chan any),
		queue:  pprque.New(),
		logger: zap.L().Named("pcap"),
	}
}

func (c *Capture) Start(i *Interface) error {
	var err error
	var hdl *pcapgo.EthernetHandle

	if err := i.Node.RunFunc(func() error {
		hdl, err = c.createHandle(i.Name)
		return err
	}); err != nil {
		return fmt.Errorf("failed to get PCAP handle: %w", err)
	}

	ci := &captureInterface{
		NgInterface: pcapgo.NgInterface{
			Name:        fmt.Sprintf("%s/%s", i.Node.Name(), i.Name),
			Filter:      c.BPFilter,
			LinkType:    layers.LinkTypeEthernet,
			SnapLength:  uint32(c.CaptureLength),
			OS:          "Linux",
			Description: "Linux veth pair",
			Comment:     fmt.Sprintf("Gont Network: '%s'", i.Node.Network().Name),
		},
		Handle: hdl,
		logger: c.logger.With(zap.String("intf", i.Name)),
	}

	if c.writer == nil {
		f := c.File
		if f == nil {
			tpl, err := template.New("filename").Parse(c.Filename)
			if err != nil {
				return fmt.Errorf("invalid filename template: %w", err)
			}

			var fn bytes.Buffer
			if err := tpl.Execute(&fn, filenameTemplate{
				Network:   i.Node.Network().Name,
				Host:      i.Node.Name(),
				Interface: i.Name,
			}); err != nil {
				return fmt.Errorf("failed to execute filename template: %w", err)
			}

			if f, err = os.OpenFile(fn.String(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755); err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
		}

		comment := c.Comment
		if comment == "" {
			comment = fmt.Sprintf("Captured network '%s' with Gont, the Go network testing toolkit (https://github.com/stv0g/gont)", i.Node.Network().Name)
		}

		opts := pcapgo.NgWriterOptions{
			SectionInfo: pcapgo.NgSectionInfo{
				OS:          "Linux",
				Application: "Gont",
				Comment:     comment,
			},
		}

		// The first interface has always id 0
		ci.Index = 0

		if c.writer, err = pcapgo.NewNgWriterInterface(f, ci.NgInterface, opts); err != nil {
			return fmt.Errorf("failed to create PCAPng writer: %w", err)
		}

		go c.writePackets()
	} else {
		if ci.Index, err = c.writer.AddInterface(ci.NgInterface); err != nil {
			return fmt.Errorf("failed to add interface: %w", err)
		}
	}

	ci.Statistics.StartTime = time.Now()

	c.interfaces = append(c.interfaces, ci)

	go ci.readPackets(c.queue)

	return nil
}

func (c *Capture) Flush() error {
	for c.queue.Len() > 0 {
		pkt := c.queue.Pop()
		if err := c.writer.WritePacket(pkt.CaptureInfo, pkt.Data); err != nil {
			return fmt.Errorf("failed to write packet: %w", err)
		}
	}

	for _, ci := range c.interfaces {
		if err := ci.writeStats(c.writer); err != nil {
			return fmt.Errorf("failed to write stats %w", err)
		}
	}

	return c.writer.Flush()
}

func (c *Capture) Close() error {
	close(c.stop)

	if err := c.Flush(); err != nil {
		return fmt.Errorf("failed to flush: %w", err)
	}

	return nil
}

func (c *Capture) Reader() (*pcapgo.NgReader, error) {
	if err := c.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush capture: %w", err)
	}

	if _, err := c.File.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to rewind file: %w", err)
	}

	rd, err := pcapgo.NewNgReader(c.File, pcapgo.DefaultNgReaderOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to read PCAPng file: %w", err)
	}

	return rd, nil
}

func (c *Capture) writePackets() {
	tickerPackets := time.NewTicker(1 * time.Second)
	tickerStats := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-tickerStats.C:
			for _, ci := range c.interfaces {
				if err := ci.writeStats(c.writer); err != nil {
					c.logger.Error("Failed to write stats:", zap.Error(err))
				}
				c.logger.Info("stats")
			}

		case now := <-tickerPackets.C:
			for {
				if c.queue.Len() < 1 {
					break
				}

				oldest := c.queue.Oldest()
				oldestAge := now.Sub(oldest)
				if oldestAge < 1*time.Second {
					break
				}

				pkt := c.queue.Pop()
				if err := c.writer.WritePacket(pkt.CaptureInfo, pkt.Data); err != nil {
					c.logger.Error("Failed to write packet", zap.Error(err))
				}
				c.logger.Info("packet")

			}

		case <-c.stop:
			return
		}
	}
}

func (c *Capture) createHandle(name string) (*pcapgo.EthernetHandle, error) {
	hdl, err := pcapgo.NewEthernetHandle(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open PCAP handle: %w", err)
	}

	if err := hdl.SetCaptureLength(int(c.CaptureLength)); err != nil {
		return nil, fmt.Errorf("failed to set capture length: %w", err)
	}

	if err := hdl.SetPromiscuous(c.Promisc); err != nil {
		return nil, fmt.Errorf("failed to set promiscuous mode: %w", err)
	}

	if c.BPFilter != "" {
		if c.filterInstrs == nil {
			instr, err := pcap.CompileBPFFilter(layers.LinkTypeEthernet, c.CaptureLength, c.BPFilter)
			if err != nil {
				return nil, fmt.Errorf("failed to compile BPF filter expression: %w", err)
			}

			for _, ins := range instr {
				c.filterInstrs = append(c.filterInstrs, bpf.RawInstruction{
					Op: ins.Code,
					Jt: ins.Jt,
					Jf: ins.Jf,
					K:  ins.K,
				})
			}
		}

		if err := hdl.SetBPF(c.filterInstrs); err != nil {
			return nil, fmt.Errorf("failed to set BPF filter: %w", err)
		}
	}

	return hdl, nil
}

func (ci *captureInterface) readPackets(queue *pprque.PacketPriorityQueue) {
	for {
		data, cinfo, err := ci.Handle.ReadPacketData()
		if err != nil {
			ci.logger.Error("Failed to read packet", zap.Error(err))
			break
		}

		cinfo.InterfaceIndex = ci.Index

		queue.Push(pprque.Packet{
			CaptureInfo: cinfo,
			Data:        data,
		})
	}
}

func (ci *captureInterface) writeStats(wg *pcapgo.NgWriter) error {
	counters, err := ci.Handle.Stats()
	if err != nil {
		ci.logger.Error("Failed to get interface statistics", zap.Error(err))
	}

	ci.Statistics.LastUpdate = time.Now()
	ci.Statistics.PacketsReceived = uint64(counters.Packets)
	ci.Statistics.PacketsDropped = uint64(counters.Drops)

	return wg.WriteInterfaceStats(ci.Index, ci.Statistics)
}
