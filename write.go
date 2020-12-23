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

	"go.bug.st/serial"
)

const keyLeft = 0b10000000
const keyUp = 0b01000000
const keyDown = 0b00100000
const keySelect = 0b00010000
const keyStart = 0b00001000
const keyRight = 0b00000100
const keyOpt = 0b00000010
const keyEdit = 0b00000001

func sendInput(port serial.Port, input <-chan byte) {
	for b := range input {
		sendKey(port, b)
	}
}

func sendKey(port serial.Port, b byte) {
	bytes := []byte{'C', b}
	n, err := port.Write(bytes)
	if err != nil {
		log.Fatal(err)
	}

	if n != len(bytes) {
		log.Fatalf("failed to write input: %016b\n", b)
	}
}

func enableAndResetDisplay(port serial.Port) {
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

func disconnect(port serial.Port) {
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

