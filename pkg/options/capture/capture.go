// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package capture contains the options to configure to packet capturing
package capture

import (
	"os"

	g "github.com/stv0g/gont/pkg"
	"golang.org/x/net/bpf"
)

type Length int

func (sl Length) ApplyCapture(c *g.Capture) {
	c.CaptureLength = int(sl)
}

type Promiscuous bool

func (p Promiscuous) ApplyCapture(c *g.Capture) {
	c.Promiscuous = bool(p)
}

// FilterInterface is a filter callback to limit the interfaces which will be recorded
type FilterInterfaces g.CaptureFilterInterfaceFunc

func (f FilterInterfaces) ApplyCapture(c *g.Capture) {
	c.FilterInterface = g.CaptureFilterInterfaceFunc(f)
}

// FilterPackets is a callback to filter packets within the Go application rather via BPF in the kernel
type FilterPackets g.CaptureFilterPacketFunc

func (f FilterPackets) ApplyCapture(c *g.Capture) {
	c.FilterPackets = g.CaptureFilterPacketFunc(f)
}

// FilterExpression is a libpcap filter expression
// See: https://www.tcpdump.org/manpages/pcap-filter.7.html
type FilterExpression string

func (bpf FilterExpression) ApplyCapture(c *g.Capture) {
	c.FilterExpression = string(bpf)
}

// FilterInstructions allows filtering the captured packets by providing a compiled BPF filter program.
type FilterInstructions []bpf.Instruction

func (fi FilterInstructions) ApplyCapture(c *g.Capture) {
	c.FilterInstructions = fi
}

// Comment can be used to add a custom comment to the PCAPng file
type Comment string

func (d Comment) ApplyCapture(c *g.Capture) {
	c.Comment = string(d)
}

// File writes all captured packets to a PCAPng file handle
type File struct {
	*os.File
}

func (f File) ApplyCapture(c *g.Capture) {
	c.Files = append(c.Files, f.File)
}

func ToFile(f *os.File) File { return File{f} }

// Filename writes all captured packets to a PCAPng file
type Filename string

func (fn Filename) ApplyCapture(c *g.Capture) {
	c.Filenames = append(c.Filenames, string(fn))
}

func ToFilename(fn string) Filename { return Filename(fn) }

// Pipename writes all captured packets to a PCAPng file
type Pipename string

func (pn Pipename) ApplyCapture(c *g.Capture) {
	c.Pipenames = append(c.Pipenames, string(pn))

	// Flush to pipe after each packet
	c.FlushEach = 1
}

func ToNewPipe(pn string) Pipename { return Pipename(pn) }

// Listener opens a UNIX, UDP or TCP socket to which you can connect with WireShark
// See: https://wiki.wireshark.org/CaptureSetup/Pipes.md#tcp-socket
type Listener string

func (s Listener) ApplyCapture(c *g.Capture) {
	c.Listeners = append(c.Listeners, string(s))

	// Flush to pipe after each packet
	c.FlushEach = 1
}

// Channel sends all captured packets to the provided channel.
type Channel chan g.CapturePacket

func (d Channel) ApplyCapture(c *g.Capture) {
	c.Channels = append(c.Channels, d)
}

func ToChannel(ch chan g.CapturePacket) Channel { return Channel(ch) }

// Callback provides a custom callback function which is called for each captured packet
type Callback g.CaptureCallbackFunc

func (cb Callback) ApplyCapture(c *g.Capture) {
	c.Callbacks = append(c.Callbacks, g.CaptureCallbackFunc(cb))
}

// LogKeys captures encryption keys from applications started via Gont and embeds them into PCAPng files
//
// This is achieved by passing the SSLKEYLOGFILE environment variable to each process started via Run().
// The environment variable points to a pipe from which Gont reads session secrets and embeds them into
// PCAPng files.
//
// Aside from SSLKEYLOGFILE, also WG_KEYLOGFILE is supported for capturing session secrets from
// wireguard-go
type LogKeys bool

func (lk LogKeys) ApplyCapture(c *g.Capture) {
	c.LogKeys = bool(lk)
}
