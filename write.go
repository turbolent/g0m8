package main

// # M8 SLIP Serial Receive command list
// 'S' - Theme Color command: 4 bytes. First byte is index (0 to 12), following 3 bytes is R, G, and B
// 'C' - Joypad/Controller command: 1 byte. Represents all 8 keys in hardware pin order: LEFT|UP|DOWN|SELECT|START|RIGHT|OPT|EDIT
// 'K' - Keyjazz note command: 1 or 2 bytes. First byte is note, second is velocity, if note is zero stops note and does not expect a second byte.
// 'D' - Disable command. Send this command when disconnecting from M8. No extra bytes following
// 'E' - Enable display command: No extra bytes following
// 'R' - Reset display command: No extra bytes following

import (
	"log"
	"os"
)

const keyLeft = 1 << 7
const keyUp = 1 << 6
const keyDown = 1 << 5
const keySelect = 1 << 4
const keyStart = 1 << 3
const keyRight = 1 << 2
const keyOpt = 1 << 1
const keyEdit = 1

func sendInput(port *os.File, input <-chan byte) {
	for b := range input {
		sendKey(port, b)
	}
}

func sendKey(port *os.File, b byte) {
	bytes := []byte{'C', b}
	n, err := port.Write(bytes)
	if err != nil {
		log.Fatal(err)
	}

	if n != len(bytes) {
		log.Fatalf("failed to write input: %016b\n", b)
	}
}

func enableAndResetDisplay(port *os.File) {
	log.Printf("Enabling and resetting display ...\n")

	bytes := []byte{'E', 'R'}
	n, err := port.Write(bytes)
	if err != nil {
		log.Fatal(err)
	}

	if n != len(bytes) {
		log.Fatalf("failed to enable and reset display")
	}
}

func disconnect(port *os.File) {
	log.Printf("Disconnecting ...\n")

	bytes := []byte{'D'}
	n, err := port.Write(bytes)
	if err != nil {
		log.Fatal(err)
	}

	if n != len(bytes) {
		log.Fatalf("failed to disconnect")
	}
}
