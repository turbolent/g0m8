package main

import (
	"log"
	"os"
)

func newReader(port *os.File) func(handle func(packet []byte)) {

	buf := make([]byte, 4*1024)
	readStartIndex := 0

	return func(handle func(packet []byte)) {

		// Read raw data from serial port

		n, err := port.Read(buf[readStartIndex:])
		if err != nil {
			log.Fatal(err)
		}

		// Read the raw data as a SLIP packets

		data := buf[:readStartIndex+n]

		remaining, err := decodeSLIP(data, handle)
		if err != nil {
			log.Fatal(err)
		}

		// There might be an incomplete packet,
		// copy it to the start

		readStartIndex = len(remaining)
		copy(buf[:], remaining)
	}
}
