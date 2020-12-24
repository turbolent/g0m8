package main

import (
	"encoding/hex"
	"log"
	"os"
)

func read(port *os.File, commands chan<- Command) {

	log.Printf("Reading ...\n")

	buf := make([]byte, 4*1024)
	readStartIndex := 0

	for {
		// Read raw data from serial port

		n, err := port.Read(buf[readStartIndex:])
		if err != nil {
			log.Fatal(err)
		}

		// Read the raw data as a SLIP packets

		data := buf[:readStartIndex+n]

		packets, remaining, err := decodeSLIP(data)
		if err != nil {
			log.Fatal(err)
		}

		// Decode the packets as commands

		for _, packet := range packets {

			command, err := decodeCommand(packet)
			if err != nil {
				log.Println(err)
				log.Println("data:", hex.Dump(data))
				log.Println("packet:", hex.Dump(packet))
				if _, ok := err.(unknownCommandError); ok {
					continue
				}
				os.Exit(1)
			}

			commands <- command
		}

		// There might be an incomplete packet,
		// copy it to the start

		readStartIndex = len(remaining)
		copy(buf[:], remaining)
	}
}
