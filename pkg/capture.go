// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"syscall"
	"text/template"
	"time"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/pcapgo"
	"github.com/stv0g/gont/internal/prque"
	"go.uber.org/zap"
	"golang.org/x/net/bpf"
)

type (
	CaptureFilterInterfaceFunc func(i *Interface) bool
	CaptureFilterPacketFunc    func(p *CapturePacket) bool
	CaptureCallbackFunc        func(pkt CapturePacket)
)

type PacketSource interface {
	ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error)
	Stats() (captureStats, error)
	LinkType() layers.LinkType
}

type filenameTemplate struct {
	Interface string
	Host      string
	Network   string
}

type captureInterface struct {
	*Interface

	pcapInterfaceIndex int
	pcapInterface      pcapgo.NgInterface

	StartTime time.Time

	source PacketSource
	logger *zap.Logger
}

type captureStats struct {
	PacketsReceived uint64
	PacketsDropped  uint64
}

type CaptureOption interface {
	ApplyCapture(n *Capture)
}

type Capture struct {
	// Options
	CaptureLength int
	Promiscuous   bool
	Comment       string
	Timeout       time.Duration
	LogKeys       bool
	FlushEach     uint64

	// Filter options
	FilterInterface    CaptureFilterInterfaceFunc
	FilterPackets      CaptureFilterPacketFunc
	FilterExpression   string
	FilterInstructions []bpf.Instruction

	// Output options
	Files     []*os.File
	Filenames []string
	Channels  []chan CapturePacket
	Callbacks []CaptureCallbackFunc
	Pipenames []string
	Listeners []string

	writer     *pcapgo.NgWriter
	stop       chan any
	queue      *prque.PriorityQueue
	count      atomic.Uint64
	interfaces []*captureInterface
	logger     *zap.Logger
}

func (c *Capture) ApplyInterface(i *Interface) {
	i.Captures = append(i.Captures, c)
}

func (c *Capture) ApplyBaseNode(n *BaseNode) {
	n.Captures = append(n.Captures, c)
}

func (c *Capture) ApplyNetwork(n *Network) {
	n.Captures = append(n.Captures, c)
}

func NewCapture(opts ...CaptureOption) *Capture {
	c := &Capture{
		// Default options
		CaptureLength: 1600,

		stop:   make(chan any),
		queue:  prque.New(),
		logger: zap.L().Named("capture"),
	}

	for _, opt := range opts {
		opt.ApplyCapture(c)
	}

	return c
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

func (c *Capture) writeDecryptionSecret(typ uint32, payload []byte) error {
	return c.writer.WriteDecryptionSecretsBlock(typ, payload)
}

func (c *Capture) writePacket(p CapturePacket) error {
	ci := p.CaptureInfo
	ci.InterfaceIndex = p.Interface.pcapInterfaceIndex

	if err := c.writer.WritePacket(ci, p.Data); err != nil {
		return fmt.Errorf("failed to write packet: %w", err)
	}

	count := c.count.Add(1)
	if c.FlushEach > 0 && count%c.FlushEach == 0 {
		if err := c.writer.Flush(); err != nil {
			return fmt.Errorf("failed to flush: %w", err)
		}
	}

	for _, ch := range c.Channels {
		ch <- p
	}

	for _, cb := range c.Callbacks {
		cb(p)
	}

	return nil
}

func (c *Capture) writeStats(ci *captureInterface) error {
	counters, err := ci.source.Stats()
	if err != nil {
		ci.logger.Error("Failed to get interface statistics", zap.Error(err))
	}

	return c.writer.WriteInterfaceStats(ci.pcapInterfaceIndex, pcapgo.NgInterfaceStatistics{
		StartTime:       ci.StartTime,
		LastUpdate:      time.Now(),
		PacketsReceived: counters.PacketsReceived,
		PacketsDropped:  counters.PacketsDropped,
	})
}

func (c *Capture) writePackets() {
	tickerPackets := time.NewTicker(1 * time.Second)
	tickerStats := time.NewTicker(10 * time.Second)

out:
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
					c.logger.Error("Failed to handle packet. Stop capturing...", zap.Error(err))
					break out
				}
			}

		case <-c.stop:
			return
		}
	}
}

func (c *Capture) createWriter(i *captureInterface) (*pcapgo.NgWriter, error) {
	wrs := []io.Writer{}

	// File handlers
	for _, file := range c.Files {
		wrs = append(wrs, file)
	}

	// Filenames
	for _, filename := range c.Filenames {
		if i.Interface != nil {
			b := &bytes.Buffer{}

			tpl, err := template.New("filename").Parse(filename)
			if err != nil {
				return nil, fmt.Errorf("invalid filename template: %w", err)
			}

			if err := tpl.Execute(b, filenameTemplate{
				Network:   i.Node.Network().Name,
				Host:      i.Node.Name(),
				Interface: i.Name,
			}); err != nil {
				return nil, fmt.Errorf("failed to execute filename template: %w", err)
			}

			filename = b.String()
		}

		var file *os.File
		var err error
		if file, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o755); err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}

		wrs = append(wrs, file)
	}

	// Pipenames
	for _, pipename := range c.Pipenames {
		logger := c.logger.With(zap.String("path", pipename))

		if stat, err := os.Stat(pipename); err != nil {
			if os.IsNotExist(err) {
				logger.Debug("Pipe does not exist yet. Creating..")
				if err := syscall.Mkfifo(pipename, 0o644); err != nil {
					return nil, fmt.Errorf("failed to create fifo: %w", err)
				}
			} else {
				return nil, fmt.Errorf("failed to stat pipe %s: %w", pipename, err)
			}
		} else if stat.Mode()&os.ModeNamedPipe == 0 {
			logger.Debug("Non-pipe exists. Removing before recreating")
			if err := os.RemoveAll(pipename); err != nil {
				return nil, fmt.Errorf("failed to delete: %w", err)
			}
			if err := syscall.Mkfifo(pipename, 0o644); err != nil {
				return nil, fmt.Errorf("failed to create fifo: %w", err)
			}
		}

		logger.Info("Opening named pipe. Waiting for a reader...")

		fifo, err := os.OpenFile("/var/run/pcap", os.O_WRONLY, 0o300)
		if err != nil {
			return nil, fmt.Errorf("failed to open fifo: %w", err)
		}

		logger.Info("Reader opened remote site of the fifo. Continuing execution")

		wrs = append(wrs, fifo)
	}

	// Listeners
	for _, lAddr := range c.Listeners {
		listener, err := newCaptureListener(lAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create listener: %w", err)
		}

		// Wait for first connection before proceeding
		c.logger.Info("Opened listener. Waiting for a reader...", zap.String("addr", lAddr))
		<-listener.Conns

		wrs = append(wrs, listener)
	}

	wr := io.MultiWriter(wrs...)

	comment := c.Comment
	if comment == "" {
		if i.Interface == nil {
			comment = fmt.Sprintf("Captured with Gont, the Go network testing toolkit (https://github.com/stv0g/gont)")
		} else {
			comment = fmt.Sprintf("Captured network '%s' with Gont, the Go network testing toolkit (https://github.com/stv0g/gont)", i.Node.Network().Name)
		}
	}

	opts := pcapgo.NgWriterOptions{
		SectionInfo: pcapgo.NgSectionInfo{
			OS:          "Linux",
			Application: "Gont",
			Comment:     comment,
		},
	}

	// The first interface has always id 0
	i.pcapInterfaceIndex = 0

	writer, err := pcapgo.NewNgWriterInterface(wr, i.pcapInterface, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create PCAPng writer: %w", err)
	}

	return writer, err
}

func (c *Capture) readPackets(ci *captureInterface) {
	var err error

	for {
		var cp CapturePacket
		cp.Data, cp.CaptureInfo, err = ci.source.ReadPacketData()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				c.logger.Error("Failed to read packet data", zap.Error(err))
				continue
			}
		}

		cp.Interface = ci

		if c.FilterPackets == nil || c.FilterPackets(&cp) {
			c.queue.Push(cp)
		}
	}
}

func (c *Capture) startInterface(i *Interface) (*captureInterface, error) {
	var err error
	var hdl PacketSource

	if err := i.Node.RunFunc(func() error {
		hdl, err = c.createPCAPHandle(i.Name)
		return err
	}); err != nil {
		return nil, fmt.Errorf("failed to get PCAP handle: %w", err)
	}

	ci := &captureInterface{
		Interface: i,
		source:    hdl,
		pcapInterface: pcapgo.NgInterface{
			Name:        fmt.Sprintf("%s/%s", i.Node.Name(), i.Name),
			Filter:      c.FilterExpression,
			LinkType:    hdl.LinkType(),
			SnapLength:  uint32(c.CaptureLength),
			OS:          "Linux",
			Description: "Linux veth pair",
			Comment:     fmt.Sprintf("Gont Network: '%s'", i.Node.Network().Name),
		},
		logger: c.logger.With(zap.String("intf", i.Name)),
	}

	if c.writer == nil {
		if c.writer, err = c.createWriter(ci); err != nil {
			return nil, err
		}

		go c.writePackets()
	} else {
		if ci.pcapInterfaceIndex, err = c.writer.AddInterface(ci.pcapInterface); err != nil {
			return nil, fmt.Errorf("failed to add interface: %w", err)
		}
	}

	ci.StartTime = time.Now()

	c.interfaces = append(c.interfaces, ci)

	go c.readPackets(ci)

	return ci, nil
}

func (c *Capture) startTrace() (*captureInterface, *traceEventPacketSource, error) {
	var err error

	tps := newTracepointPacketSource()

	ci := &captureInterface{
		pcapInterface: pcapgo.NgInterface{
			Name:        "tracer",
			LinkType:    LinkTypeTrace,
			SnapLength:  uint32(c.CaptureLength),
			OS:          "Debug",
			Description: "Trace output",
		},
		source: tps,
		logger: c.logger.With(zap.String("intf", "tracer")),
	}

	if c.writer == nil {
		if c.writer, err = c.createWriter(ci); err != nil {
			return nil, nil, err
		}

		go c.writePackets()
	} else {
		if ci.pcapInterfaceIndex, err = c.writer.AddInterface(ci.pcapInterface); err != nil {
			return nil, nil, fmt.Errorf("failed to add interface: %w", err)
		}
	}

	ci.StartTime = time.Now()

	c.interfaces = append(c.interfaces, ci)

	go c.readPackets(ci)

	return ci, tps, nil
}
