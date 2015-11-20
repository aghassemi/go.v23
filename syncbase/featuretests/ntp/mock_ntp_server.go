// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

const (
	udp  = "udp"
	port = "10001"
)

// TODO(sadovsky): Add the test logic to provide next server timestamp here.
func nextTimestamp() time.Time {
	return time.Now()
}

func main() {
	saddr, err := net.ResolveUDPAddr(udp, ":"+port)
	exitOnError(err)

	// Now listen at selected port
	conn, err := net.ListenUDP(udp, saddr)
	exitOnError(err)
	defer conn.Close()
	fmt.Printf("Listening to UDP packets on port %v...\n", port)

	buf := make([]byte, 48)
	for {
		_, caddr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Fatalf("Error while reading: ", err)
		}
		// Client will send its transmit timestamp starting at the 40th byte
		clientTransmitTs := extractTime(buf[40:48])
		fmt.Printf("Received timestamp %v\n", clientTransmitTs)
		serverRecvTs := nextTimestamp()
		serverTransmitTs := serverRecvTs
		resp := createResponse(clientTransmitTs, serverRecvTs, serverTransmitTs)
		_, err = conn.WriteTo(resp, caddr)
		if err != nil {
			log.Fatalf("Error while writing: ", err)
		}
	}
}

func createResponse(clientTransmitTs, serverRecvTs, serverTransmitTs time.Time) []byte {
	resp := make([]byte, 48)
	// the client's transmit timestamp is sent back starting at byte 24
	writeTime(resp[24:32], clientTransmitTs)
	// server's receive timestamp is set starting at byte 32
	writeTime(resp[32:40], serverRecvTs)
	// server's transmit timestamp is set starting at byte 40
	writeTime(resp[40:48], serverTransmitTs)
	return resp
}

// ExtractTime takes a byte array which contains encoded timestamp from NTP
// server starting at the 0th byte and is 8 bytes long. The encoded timestamp is
// in seconds since 1900. The first 4 bytes contain the integer part of of the
// seconds while the last 4 bytes contain the fractional part of the seconds
// where (FFFFFFFF + 1) represents 1 second while 00000001 represents 2^(-32) of
// a second.
func extractTime(data []byte) time.Time {
	var sec, frac uint64
	sec = uint64(data[3]) | uint64(data[2])<<8 | uint64(data[1])<<16 | uint64(data[0])<<24
	frac = uint64(data[7]) | uint64(data[6])<<8 | uint64(data[5])<<16 | uint64(data[4])<<24

	// multiply the integral second part with 1Billion to convert to nanoseconds
	nsec := sec * 1e9
	// multiply frac part with 2^(-32) to get the correct value in seconds and
	// then multiply with 1Billion to convert to nanoseconds. The multiply by
	// Billion is done first to make sure that we dont loose precision.
	nsec += (frac * 1e9) >> 32

	t := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(nsec)).Local()

	return t
}

// writeTime writes the given time as the first 8 bytes of the byte array.
func writeTime(data []byte, tnow time.Time) {
	// For NTP the prime epoch, or base date of era 0, is 0 h 1 January 1900 UTC
	t0 := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	d := tnow.Sub(t0)
	nsec := d.Nanoseconds()

	// The encoding of timestamp below is an exact opposite of the decoding
	// being done in extractTime(). Refer extractTime() for more explaination.
	sec := nsec / 1e9                  // Integer part of seconds since epoch
	frac := ((nsec % 1e9) << 32) / 1e9 // fractional part of seconds since epoch

	// write the timestamp to Transmit Timestamp section of request.
	data[3] = byte(sec)
	data[2] = byte(sec >> 8)
	data[1] = byte(sec >> 16)
	data[0] = byte(sec >> 24)

	data[7] = byte(frac)
	data[6] = byte(frac >> 8)
	data[5] = byte(frac >> 16)
	data[4] = byte(frac >> 24)
}

func exitOnError(err error) {
	if err != nil {
		log.Fatalf("Error: ", err)
	}
}
