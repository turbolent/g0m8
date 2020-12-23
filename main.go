package main

import (
	"flag"
	"io/ioutil"
	"log"

	"go.bug.st/serial"
)

var listFlag = flag.Bool("list", false, "list available devices")
var deviceFlag = flag.String("device", "", "connect to given device")
var debugFlag = flag.Bool("debug", true, "enable debug logging")

func main() {
	flag.Parse()

	if !*debugFlag {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	switch {
	case *listFlag:
		listDevices()
	case *deviceFlag != "":
		run(*deviceFlag)
	default:
		flag.Usage()
	}
}

func run(device string) {
	port := openSerialPort(device)

	enableAndResetDisplay(port)

	commands := make(chan Command, 256)
	input := make(chan byte)
	go read(port, commands)
	go sendInput(port, input)
	render(commands, input)

	disconnect(port)
}

func listDevices() {
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	}
	log.Printf("Found devices:\n")
	for _, port := range ports {
		log.Printf("- %s\n", port)
	}
}
