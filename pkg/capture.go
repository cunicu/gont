// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package gont

import (
	"bytes"
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
	Node      string
	Network   string
	PID       int
}

func (t filenameTemplate) execute(filename string) (string, error) {
	b := &bytes.Buffer{}

	tpl, err := template.New("filename").Parse(filename)
	if err != nil {
		return "", fmt.Errorf("invalid filename template: %w", err)
	}

	if err := tpl.Execute(b, t); err != nil {
		return "", fmt.Errorf("failed to execute filename template: %w", err)
	}

	return b.String(), nil
}

type CaptureOption interface {
	ApplyCapture(n *Capture)
}

type Capture struct {
	// Options
	SnapshotLength int
	Promiscuous    bool
	Comment        string
	Timeout        time.Duration
	LogKeys        bool
	FlushEach      uint64

	// Filter options
	FilterInterface    CaptureFilterInterfaceFunc
	FilterPackets      CaptureFilterPacketFunc
	FilterExpression   string
	FilterInstructions []bpf.Instruction

	// Output options
	Files       []*os.File
	Filenames   []string
	Channels    []chan CapturePacket
	Callbacks   []CaptureCallbackFunc
	Pipenames   []string
	ListenAddrs []string

	writer     *pcapgo.NgWriter
	stop       chan any
	queue      *prque.PriorityQueue
	count      atomic.Uint64
	interfaces []*captureInterface
	closables  []io.Closer
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

func (c *Capture) ApplyTracer(n *Tracer) {
	n.Captures = append(n.Captures, c)
}

func NewCapture(opts ...CaptureOption) *Capture {
	c := &Capture{
		// Default options
		SnapshotLength: 1600,

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

		if err := c.writePacket(p); err != nil {
			return err
		}
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

	for _, closable := range c.closables {
		if err := closable.Close(); err != nil {
			return fmt.Errorf("failed to close: %w", err)
		}
	}

	return nil
}

func (c *Capture) newPacket(cp CapturePacket) {
	if c.FilterPackets == nil || c.FilterPackets(&cp) {
		c.queue.Push(cp)
	}
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
	var err error
	wrs := []io.Writer{}

	// File handlers
	for _, file := range c.Files {
		wrs = append(wrs, file)
	}

	// Filenames
	for _, filename := range c.Filenames {
		file, err := c.openFile(filename, i)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}

		wrs = append(wrs, file)
		c.closables = append(c.closables, file)
	}

	// Pipenames
	for _, pipename := range c.Pipenames {
		pipe, err := c.createAndOpenPipe(pipename)
		if err != nil {
			return nil, fmt.Errorf("failed to create pipe: %w", err)
		}

		wrs = append(wrs, pipe)
		c.closables = append(c.closables, pipe)
	}

	// Listeners
	for _, lAddr := range c.ListenAddrs {
		listener, err := c.createListener(lAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create listener: %w", err)
		}

		wrs = append(wrs, listener)
		c.closables = append(c.closables, listener)
	}

	wr := io.MultiWriter(wrs...)

	comment := c.Comment
	if comment == "" {
		if i.Interface == nil {
			comment = "Captured with Gont, the Go network testing toolkit (https://github.com/stv0g/gont)"
		} else {
			comment = fmt.Sprintf("Captured network '%s' with Gont, the Go network testing toolkit (https://github.com/stv0g/gont)", i.node.Network().Name)
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

func (c *Capture) startInterface(i *Interface) (*captureInterface, error) {
	var err error
	var hdl PacketSource

	if err := i.node.RunFunc(func() error {
		hdl, err = c.createPCAPHandle(i.Name)
		return err
	}); err != nil {
		return nil, fmt.Errorf("failed to get PCAP handle: %w", err)
	}

	ci := &captureInterface{
		Interface: i,
		source:    hdl,
		pcapInterface: pcapgo.NgInterface{
			Name:        fmt.Sprintf("%s/%s", i.node.Name(), i.Name),
			Filter:      c.FilterExpression,
			LinkType:    hdl.LinkType(),
			SnapLength:  uint32(c.SnapshotLength),
			OS:          "Linux",
			Description: "Linux veth pair",
			Comment:     fmt.Sprintf("Gont Network: '%s'", i.node.Network().Name),
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

	go ci.readPackets(c)

	return ci, nil
}

func (c *Capture) startTrace() (*captureInterface, *traceEventPacketSource, error) {
	var err error

	tps := newTracepointPacketSource()

	ci := &captureInterface{
		pcapInterface: pcapgo.NgInterface{
			Name:        "tracer",
			LinkType:    LinkTypeTrace,
			SnapLength:  uint32(c.SnapshotLength),
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

	go ci.readPackets(c)

	return ci, tps, nil
}

func (c *Capture) createListener(lAddr string) (*captureListener, error) {
	listener, err := newCaptureListener(lAddr)
	if err != nil {
		return nil, err
	}

	// Wait for first connection before proceeding
	c.logger.Info("Opened listener. Waiting for a reader...", zap.String("addr", lAddr))
	<-listener.Conns

	return listener, nil
}

func (c *Capture) createAndOpenPipe(pipename string) (*os.File, error) {
	logger := c.logger.With(zap.String("path", pipename))

	if stat, err := os.Stat(pipename); err != nil { //nolint:nestif
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to stat pipe %s: %w", pipename, err)
		}

		logger.Debug("Pipe does not exist yet. Creating..")
		if err := syscall.Mkfifo(pipename, 0o644); err != nil {
			return nil, fmt.Errorf("failed to create fifo: %w", err)
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

	pipe, err := os.OpenFile(pipename, os.O_WRONLY, 0o300)
	if err != nil {
		return nil, fmt.Errorf("failed to open fifo: %w", err)
	}

	logger.Info("Reader opened remote site of the fifo. Continuing execution")

	return pipe, nil
}

func (c *Capture) openFile(filename string, i *captureInterface) (*os.File, error) {
	var err error

	if i.Interface != nil {
		tpl := filenameTemplate{
			Network:   i.node.Network().Name,
			Node:      i.node.Name(),
			Interface: i.Name,
		}

		if filename, err = tpl.execute(filename); err != nil {
			return nil, err
		}
	}

	return os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
}
