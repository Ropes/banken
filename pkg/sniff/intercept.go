package sniff

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

// Adapted from original Source: https://github.com/google/gopacket/blob/master/examples/httpassembly/main.go
// Copyright 2012 Google, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

// This binary provides sample code for using the gopacket TCP assembler and TCP
// stream reader.  It reads packets off the wire and reconstructs HTTP requests
// it sees, logging them.

// httpStreamFactory implements tcpassembly.StreamFactory
type httpStreamFactory struct {
	logger *log.Logger
}

func (h *httpStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	hstream := &httpXStream{
		net:       net,
		transport: transport,
		r:         tcpreader.NewReaderStream(),
		logger:    h.logger,
	}
	go hstream.run() // Important... we must guarantee that data from the reader stream is read.

	// ReaderStream implements tcpassembly.Stream, so we can return a pointer to it.
	return &hstream.r
}

// httpStream will handle the actual decoding of http requests.
type httpXStream struct {
	net       gopacket.Flow
	transport gopacket.Flow
	r         tcpreader.ReaderStream
	logger    *log.Logger
}

func (h *httpXStream) run() {
	buf := bufio.NewReader(&h.r)
	for {
		// TODO: Detect the packet type before reading
		req, err := http.ReadRequest(buf)
		if err == io.EOF {
			h.logger.Trace("EOF signaled")
			// We must read until we see an EOF... very important!
			return
		} else if err != nil {
			// Common error case from TCP SYN, ACK communication.
			var errStr string
			if len(err.Error()) > 30 {
				errStr = err.Error()[:30]
			} else {
				errStr = err.Error()
			}
			h.logger.WithFields(log.Fields{"net": h.net, "transport": h.transport, "err": errStr}).
				Tracef("http.ReadRequest error reading packet")
		} else if req != nil {
			// HTTP data was read into request
			h.logger.WithFields(log.Fields{"host": req.Host, "path": req.URL.Path, "method": req.Method,
				"transport": h.transport, "net": h.net, "time": time.Now()}).Info("httpX packet read")
			// TODO: Create HTTPXPacket to return to processors.
		} else {
			h.logger.Trace("http packet read failed")
		}
	}
}

// HTTPXPacket provides information to categorize HTTP requests.
type HTTPXPacket struct {
	TS        time.Time
	Protocol  string
	Host      string
	Path      string
	Method    string
	Transport string
	Net       string
}

// InterfaceListener establishes a libpcap listener and BPF matching
// for capturing and reconstructing packets.
func InterfaceListener(ctx context.Context, iface, bpfFilter string, snaplen int, logger *log.Logger) {
	// TODO: Return reconstructed packet data via channel.
	var handle *pcap.Handle
	var err error

	// Set up pcap packet capture
	logger.Infof("Starting capture on interface %q", iface)
	handle, err = pcap.OpenLive(iface, int32(snaplen), true, pcap.BlockForever)
	if err != nil {
		logger.Fatal(err)
	}

	if err := handle.SetBPFFilter(bpfFilter); err != nil {
		logger.Fatal(err)
	}

	// Set up assembly
	streamFactory := &httpStreamFactory{
		logger: logger,
	}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)

	logger.Debugf("reading in packets from %s", iface)
	// Read in packets, pass to assembler.
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packets := packetSource.Packets()
	ticker := time.Tick(time.Minute)
	for {
		select {
		case <-ctx.Done():
			return
		case packet := <-packets:
			if packet.NetworkLayer() == nil || packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
				logger.Tracef("Unreadable packet: %#v", packet.String())
				continue
			}
			tcp := packet.TransportLayer().(*layers.TCP)
			assembler.AssembleWithTimestamp(packet.NetworkLayer().NetworkFlow(), tcp, packet.Metadata().Timestamp)
			//logger.Infof("%v", packet.String())

		case <-ticker:
			// Every minute, flush connections that haven't seen activity in the past 2 minutes.
			assembler.FlushOlderThan(time.Now().Add(time.Minute * -2))
		}
	}
}
