package gont

import (
	"bytes"
	"fmt"
	"os"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"github.com/stv0g/gont/internal/prque"
	"go.uber.org/zap"
	"golang.org/x/net/bpf"
)

type handle interface {
	ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error)
	Stats() (CaptureStats, error)
	LinkType() layers.LinkType
}

type filenameTemplate struct {
	Interface string
	Host      string
	Network   string
}

type CaptureFilterInterfaceFunc func(i *Interface) bool
type CaptureFilterPacketFunc func(p *CapturePacket) bool
type CaptureCallbackFunc func(pkt CapturePacket)

type CaptureInterface struct {
	*Interface

	PCAPInterfaceIndex int
	Handle             handle

	StartTime time.Time

	logger *zap.Logger
}

type CapturePacket struct {
	gopacket.Packet

	Interface *CaptureInterface
}

func (p CapturePacket) Time() time.Time {
	return p.Metadata().Timestamp
}

type CaptureStats struct {
	PacketsReceived int
	PacketsDropped  int
}

type CaptureOption interface {
	Apply(n *Capture)
}

type Capture struct {
	// Options
	CaptureLength int
	Promiscuous   bool
	Comment       string
	Timeout       time.Duration
	LogKeys       bool

	// Filter options
	FilterInterface    CaptureFilterInterfaceFunc
	FilterPackets      CaptureFilterPacketFunc
	FilterExpression   string
	FilterInstructions []bpf.Instruction

	// Output options
	File     *os.File
	Filename string
	Channel  chan CapturePacket
	Callback CaptureCallbackFunc

	writer *pcapgo.NgWriter

	stop chan any

	queue *prque.PriorityQueue
	count atomic.Uint64

	interfaces []*CaptureInterface

	logger *zap.Logger
}

func (c *Capture) Apply(i *Interface) {
	i.Captures = append(i.Captures, c)
}

func NewCapture() *Capture {
	return &Capture{
		// Default options
		CaptureLength: 1600,

		stop:   make(chan any),
		queue:  prque.New(),
		logger: zap.L().Named("pcap"),
	}
}

func (c *Capture) Start(i *Interface) error {
	var err error
	var hdl handle

	if err := i.Node.RunFunc(func() error {
		hdl, err = c.createHandle(i.Name)
		return err
	}); err != nil {
		return fmt.Errorf("failed to get PCAP handle: %w", err)
	}

	ci := &CaptureInterface{
		Interface: i,
		Handle:    hdl,
		logger:    c.logger.With(zap.String("intf", i.Name)),
	}

	intf := pcapgo.NgInterface{
		Name:        fmt.Sprintf("%s/%s", i.Node.Name(), i.Name),
		Filter:      c.FilterExpression,
		LinkType:    hdl.LinkType(),
		SnapLength:  uint32(c.CaptureLength),
		OS:          "Linux",
		Description: "Linux veth pair",
		Comment:     fmt.Sprintf("Gont Network: '%s'", i.Node.Network().Name),
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
		ci.PCAPInterfaceIndex = 0

		if c.writer, err = pcapgo.NewNgWriterInterface(f, intf, opts); err != nil {
			return fmt.Errorf("failed to create PCAPng writer: %w", err)
		}

		go c.writePackets()
	} else {
		if ci.PCAPInterfaceIndex, err = c.writer.AddInterface(intf); err != nil {
			return fmt.Errorf("failed to add interface: %w", err)
		}
	}

	ci.StartTime = time.Now()

	c.interfaces = append(c.interfaces, ci)

	go ci.readPackets(c.queue, c.FilterPackets)

	return nil
}

// Count returns the total number of captured packets
func (c *Capture) Count() uint64 {
	return uint64(c.count.Load())
}

func (c *Capture) Flush() error {
	for c.queue.Len() > 0 {
		p := c.queue.Pop().(CapturePacket)

		c.writePacket(p)
	}

	for _, ci := range c.interfaces {
		if err := c.writeStats(ci); err != nil {
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

func (c *Capture) writePacket(p CapturePacket) error {
	ci := p.Metadata().CaptureInfo
	ci.InterfaceIndex = p.Interface.PCAPInterfaceIndex

	c.count.Add(1)

	if err := c.writer.WritePacket(ci, p.Data()); err != nil {
		c.logger.Error("Failed to write packet", zap.Error(err))
	}

	if c.Channel != nil {
		c.Channel <- p
	}

	if c.Callback != nil {
		c.Callback(p)
	}

	return nil
}

func (c *Capture) writeStats(ci *CaptureInterface) error {
	counters, err := ci.Handle.Stats()
	if err != nil {
		ci.logger.Error("Failed to get interface statistics", zap.Error(err))
	}

	return c.writer.WriteInterfaceStats(ci.PCAPInterfaceIndex, pcapgo.NgInterfaceStatistics{
		StartTime:       ci.StartTime,
		LastUpdate:      time.Now(),
		PacketsReceived: uint64(counters.PacketsReceived),
		PacketsDropped:  uint64(counters.PacketsDropped),
	})
}

func (c *Capture) writePackets() {
	tickerPackets := time.NewTicker(1 * time.Second)
	tickerStats := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-tickerStats.C:
			for _, ci := range c.interfaces {
				if err := c.writeStats(ci); err != nil {
					c.logger.Error("Failed to write stats:", zap.Error(err))
				}
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

				p := c.queue.Pop().(CapturePacket)

				if err := c.writePacket(p); err != nil {
					c.logger.Error("Failed to handle packet", zap.Error(err))
				}
			}

		case <-c.stop:
			return
		}
	}
}

func (c *Capture) WriteDecryptionSecret(typ uint32, payload []byte) error {
	return c.writer.WriteDecryptionSecretsBlock(typ, payload)
}

func (ci *CaptureInterface) readPackets(queue *prque.PriorityQueue, filter CaptureFilterPacketFunc) {
	src := gopacket.NewPacketSource(ci.Handle, ci.Handle.LinkType())

	for {
		p, err := src.NextPacket()
		if err != nil {
			ci.logger.Error("Failed to decode next packet", zap.Error(err))
			continue
		}

		cp := CapturePacket{
			Packet:    p,
			Interface: ci,
		}

		if filter == nil || filter(&cp) {
			queue.Push(cp)
		}
	}
}
