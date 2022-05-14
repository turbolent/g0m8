package main

import (
	"encoding/hex"
	"flag"
	"io/ioutil"
	"log"
)

var deviceFlag = flag.String("device", "", "connect to given device")
var debugFlag = flag.Bool("debug", true, "enable debug logging")
var softwareFlag = flag.Bool("software", true, "use software rendering")
var widthFlag = flag.Int("width", 640, "width of the window")
var heightFlag = flag.Int("height", 460, "height of the window")

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

	var input input
	screen := newScreen()

	windowWidth := int32(*widthFlag)
	windowHeight := int32(*heightFlag)

	renderer := newSDLRenderer(windowWidth, windowHeight, *softwareFlag)
	defer renderer.quit()

	log.Printf("Opening serial port ...")

	port := openSerialPort(device)
	defer port.Close()
	defer disconnect(port)

	enableAndResetDisplay(port)

	read := newReader(port)

	sendController := func(controller byte) {
		sendController(port, controller)
	}

	for {
		if !input.handle(renderer.toggleFullscreen, sendController) {
			log.Println("Quit")
			return
		}

		var render bool
		read(func(packet []byte) {
			command, err := decodeCommand(packet)
			if err != nil {
				log.Printf(
					"failed to decode packet: %s. packet: %s",
					err.Error(),
					hex.Dump(packet),
				)
				if _, ok := err.(unknownCommandError); !ok {
					return
				}
			}

			if screen.update(command) {
				render = true
			}
		})

		if render {
			renderer.render(screen)
		}
	}
}
