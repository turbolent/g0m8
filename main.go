package main

import (
	"encoding/hex"
	"flag"
	"io/ioutil"
	"log"

	"github.com/veandco/go-sdl2/sdl"
)

var deviceFlag = flag.String("device", "", "connect to given device")
var debugFlag = flag.Bool("debug", true, "enable debug logging")
var softwareFlag = flag.Bool("software", true, "use software rendering")
var widthFlag = flag.Int("width", 640, "width of the window")
var heightFlag = flag.Int("height", 480, "height of the window")
var fpsFlag = flag.Int("fps", 30, "target FPS")

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

	fps := *fpsFlag

	var lastRender uint64
	var skippedRender bool

	for {
		if !input.handle(func() {
			renderer.toggleFullscreen()
			enableAndResetDisplay(port)
		}, sendController) {
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

			if renderer.draw(command) {
				render = true
			}
		})

		if skippedRender || render {
			skippedRender = false

			now := sdl.GetPerformanceCounter()

			diff := float64(now-lastRender) / float64(sdl.GetPerformanceFrequency())

			if diff < (1.0 / float64(fps)) {
				skippedRender = true
			} else {
				renderer.render()

				lastRender = now
			}
		}
	}
}
