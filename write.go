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

const (
	keyLeft   = 1 << 7
	keyUp     = 1 << 6
	keyDown   = 1 << 5
	keySelect = 1 << 4
	keyStart  = 1 << 3
	keyRight  = 1 << 2
	keyOpt    = 1 << 1
	keyEdit   = 1
)

var sendControllerCommand = []byte{'C', 0}

func sendController(port *os.File, controller byte) {
	sendControllerCommand[1] = controller

	n, err := port.Write(sendControllerCommand)
	if err != nil {
		log.Fatal(err)
	}

	if n != len(sendControllerCommand) {
		log.Fatalf("failed to send controller: %016b", controller)
	}
}

var enableAndResetDisplayCommand = []byte{'E', 'R'}

func enableAndResetDisplay(port *os.File) {
	log.Println("Enabling and resetting display ...")

	n, err := port.Write(enableAndResetDisplayCommand)
	if err != nil {
		log.Fatal(err)
	}

	if n != len(enableAndResetDisplayCommand) {
		log.Fatalln("failed to enable and reset display")
	}
}

var disconnectCommand = []byte{'D'}

func disconnect(port *os.File) {
	log.Println("Disconnecting ...")

	n, err := port.Write(disconnectCommand)
	if err != nil {
		log.Fatal(err)
	}

	if n != len(disconnectCommand) {
		log.Fatalln("failed to disconnect")
	}
}
