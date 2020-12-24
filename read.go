package main

import (
	"log"
	"os"
)

func read(port *os.File, commands chan<- Command) {

	log.Printf("Reading ...\n")

	buf := make([]byte, 1024)
	readStartIndex := 0

	for {
		// Read raw data from serial port

		n, err := port.Read(buf[readStartIndex:])
		if err != nil {
			log.Fatal(err)
		}

		// Read the raw data as a SLIP packets

		packets, remaining, err := decodeSLIP(buf[:readStartIndex+n])
		if err != nil {
			log.Fatal(err)
		}

		// Decode the packets as commands

		for _, packet := range packets {

			command, err := decodeCommand(packet)
			if err != nil {
				log.Fatal(err)
			}

			commands <- command
		}

		// There might be an incomplete packet,
		// copy it to the start

		readStartIndex = len(remaining)
		copy(buf[:], remaining)
	}
}
