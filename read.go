package main

import (
	"log"

	"go.bug.st/serial"
)

func read(port serial.Port, commands chan<- Command) {

	log.Printf("Reading ...\n")

	buf := make([]byte, 2048)
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

func openSerialPort(device string) serial.Port {
	port, err := serial.Open(
		device,
		&serial.Mode{
			BaudRate: 115200,
			DataBits: 8,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	return port
}
