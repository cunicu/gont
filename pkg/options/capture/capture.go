// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package capture contains the options to configure to packet capturing
package capture

import (
	"os"

	g "github.com/stv0g/gont/pkg"
	"golang.org/x/net/bpf"
)

// SnapshotLength defines the PCAP snapshot length.
//
// If, when capturing, you capture the entire contents of the packet, that
// requires more CPU time to copy the packet to your application, more disk
// and possibly network bandwidth to write the packet data to a file, and more
// disk space to save the packet. If you don't need the entire contents of the
// packet - for example, if you are only interested in the TCP headers of
// packets - you can set the "snapshot length" for the capture to an
// appropriate value. If the snapshot length is set to snaplen, and snaplen is
// less than the size of a packet that is captured, only the first snaplen bytes
// of that packet will be captured and provided as packet data.
// A snapshot length of 65535 should be sufficient, on most if not all networks,
// to capture all the data available from the packet.
//
// See: https://www.tcpdump.org/manpages/pcap.3pcap.html
type SnapshotLength int

func (sl SnapshotLength) ApplyCapture(c *g.Capture) {
	c.SnapshotLength = int(sl)
}

// Promiscuous enables capturing of all packets.
//
// On broadcast LANs such as Ethernet, if the network isn't switched, or if the
// adapter is connected to a "mirror port" on a switch to which all packets
// passing through the switch are sent, a network adapter receives all packets
// on the LAN, including unicast or multicast packets not sent to a network
// address that the network adapter isn't configured to recognize.
// Normally, the adapter will discard those packets; however, many network
// adapters support "promiscuous mode", which is a mode in which all packets,
// even if they are not sent to an address that the adapter recognizes, are
// provided to the host. This is useful for passively capturing traffic between
// two or more other hosts for analysis.
// Note that even if an application does not set promiscuous mode, the adapter
// could well be in promiscuous mode for some other reason.
//
// See: https://www.tcpdump.org/manpages/pcap.3pcap.html
type Promiscuous bool

func (p Promiscuous) ApplyCapture(c *g.Capture) {
	c.Promiscuous = bool(p)
}

// FilterInterface is a filter callback to limit the interfaces which will be
// recorded.
//
// This option is useful if you attach the capturer to a whole network or
// node and want to filter which of the interfaces should be captured.
type FilterInterfaces g.CaptureFilterInterfaceFunc

func (f FilterInterfaces) ApplyCapture(c *g.Capture) {
	c.FilterInterface = g.CaptureFilterInterfaceFunc(f)
}

// FilterPackets is a callback to filter packets within the Go application
// rather via BPF in the kernel.
//
// Passed packets are not decoded. Have a look at gopacket to decode the
// packet layers.
type FilterPackets g.CaptureFilterPacketFunc

func (f FilterPackets) ApplyCapture(c *g.Capture) {
	c.FilterPackets = g.CaptureFilterPacketFunc(f)
}

// FilterExpression is a libpcap filter expression
//
// The filter expression consists of one or more primitives.
// Primitives usually consist of an id (name or number) preceded by one or more qualifiers
//
// See: https://www.tcpdump.org/manpages/pcap-filter.7.html
type FilterExpression string

func (bpf FilterExpression) ApplyCapture(c *g.Capture) {
	c.FilterExpression = string(bpf)
}

// FilterInstructions allows filtering the captured packets
// by providing a compiled BPF filter program.
//
// See: https://docs.kernel.org/bpf/instruction-set.html
type FilterInstructions []bpf.Instruction

func (fi FilterInstructions) ApplyCapture(c *g.Capture) {
	c.FilterInstructions = fi
}

// Comment can be used to add a custom comment to the PCAPng file.
//
// See: https://datatracker.ietf.org/doc/html/draft-ietf-opsawg-pcapng
type Comment string

func (d Comment) ApplyCapture(c *g.Capture) {
	c.Comment = string(d)
}

// File writes all captured packets in PCAPng format to the provided file handle.
type File struct {
	*os.File
}

func (f File) ApplyCapture(c *g.Capture) {
	c.Files = append(c.Files, f.File)
}

func ToFile(f *os.File) File { return File{f} }

// Filename writes all captured packets in PCAPng format to a new or existing file
// with the provided filename.
// Any existing files will be truncated
type Filename string

func (fn Filename) ApplyCapture(c *g.Capture) {
	c.Filenames = append(c.Filenames, string(fn))
}

func ToFilename(fn string) Filename { return Filename(fn) }

// Pipename writes all captured packets in PCAPng format to a newly created
// named pipe.
//
// You can use WireShark to open this named pipe to stream captures packets
// in real-time to a local machine.
//
// See: https://wiki.wireshark.org/CaptureSetup/Pipes.md#named-pipes
// See: https://man7.org/linux/man-pages/man7/fifo.7.html
type Pipename string

func (pn Pipename) ApplyCapture(c *g.Capture) {
	c.Pipenames = append(c.Pipenames, string(pn))

	// Flush to pipe after each packet
	c.FlushEach = 1
}

func ToPipename(pn string) Pipename { return Pipename(pn) }

// ListenAddr opens a UNIX, UDP or TCP socket which serves a PCAPng trace.
//
// You can use WireShark to connect to this socket to stream captured packets
// in real-time to a local/remote machine.
//
// See: https://wiki.wireshark.org/CaptureSetup/Pipes.md#tcp-socket
type ListenAddr string

func (s ListenAddr) ApplyCapture(c *g.Capture) {
	c.ListenAddrs = append(c.ListenAddrs, string(s))

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
