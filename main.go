package main

import (
	"flag"
	"io/ioutil"
	"log"
)

var deviceFlag = flag.String("device", "", "connect to given device")
var debugFlag = flag.Bool("debug", true, "enable debug logging")

func main() {
	flag.Parse()

	if !*debugFlag {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	device := *deviceFlag
	if device == "" {
		flag.Usage()
		return
	}

	run(device)
}

func run(device string) {
	port := openSerialPort(device)
	defer port.Close()

	commands := make(chan Command, 256)
	input := make(chan byte)
	go read(port, commands)

	enableAndResetDisplay(port)

	go sendInput(port, input)
	render(commands, input)

	disconnect(port)
}