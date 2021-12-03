package gont

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-ping/ping"
	log "github.com/sirupsen/logrus"
)

func (h *Host) Ping(o *Host) (*ping.Statistics, error) {
	return h.PingWithOptions(o, "ip", 1, 2*time.Second, time.Second, true)
}

func (h *Host) PingWithNetwork(o *Host, net string) (*ping.Statistics, error) {
	return h.PingWithOptions(o, net, 1, 2*time.Second, time.Second, true)
}

func (h *Host) PingWithOptions(o *Host, net string, count int, timeout time.Duration, intv time.Duration, output bool) (*ping.Statistics, error) {
	var err error

	p := ping.New(o.Name())

	p.Count = count
	p.RecordRtts = true
	p.Timeout = timeout
	p.Interval = intv

	if h.network != o.network {
		return nil, fmt.Errorf("hosts must be on same network")
	}

	// Find first IP address of first interface
	ip := o.LookupAddress(net)
	if ip == nil {
		return nil, errors.New("failed to find address")
	}

	logger := log.WithFields(log.Fields{
		"logger": "ping",
		"node":   h,
	})

	p.SetIPAddr(ip)
	p.SetLogger(logger)
	p.SetPrivileged(true)
	p.SetNetwork(net)

	if output {
		p.OnRecv = func(p *ping.Packet) {
			logger.Printf("%d bytes from %s (%s): icmp_seq=%d ttl=%d time=%v\n",
				p.Nbytes,
				p.Addr,
				p.IPAddr.String(),
				p.Seq,
				p.Ttl,
				p.Rtt,
			)
		}

		p.OnFinish = func(s *ping.Statistics) {
			logger.Printf("-- %s (%s) ping statistics ---", o.Name(), s.IPAddr)
			logger.Printf("%d packets transmitted, %d received, %d duplicates, %.2f%% packet loss\n", s.PacketsSent, s.PacketsRecv, s.PacketsRecvDuplicates, s.PacketLoss)
			logger.Printf("rtt min/avg/max/mdev = %s/%s/%s/%s\n", s.MinRtt, s.AvgRtt, s.MaxRtt, s.StdDevRtt)
		}
	}

	if err = h.RunFunc(func() error {
		if output {
			logger.Printf("PING %s(%s) %d data bytes\n",
				o.Name(),
				p.Addr(),
				p.Size,
			)
		}

		return p.Run()
	}); err != nil {
		return nil, err
	}

	lost := p.PacketsSent - p.PacketsRecv
	if lost > 0 {
		err = fmt.Errorf("lost %d packets", lost)
	}

	return p.Statistics(), err
}
