package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

type input struct {
	input uint8
	run   bool
}

func (i *input) handle(
	toggleFullscreen func(),
	sendController func(uint8),
) bool {
	event := sdl.PollEvent()
	switch event := event.(type) {
	case *sdl.QuitEvent:
		return false

	case *sdl.KeyboardEvent:
		if event.Type == sdl.KEYUP {
			switch event.Keysym.Sym {
			case sdl.K_RETURN:
				if event.Keysym.Mod&sdl.KMOD_ALT > 0 {
					toggleFullscreen()
				}

			case sdl.K_q:
				return false
			}
		}

		var key uint8

		switch event.Keysym.Sym {
		case sdl.K_RIGHT, sdl.K_KP_6:
			key = keyRight
		case sdl.K_LEFT, sdl.K_KP_4:
			key = keyLeft
		case sdl.K_UP, sdl.K_KP_8:
			key = keyUp
		case sdl.K_DOWN, sdl.K_KP_2:
			key = keyDown
		case sdl.K_x, sdl.K_m:
			key = keyEdit
		case sdl.K_z, sdl.K_n:
			key = keyOpt
		case sdl.K_SPACE:
			key = keyStart
		case sdl.K_LSHIFT, sdl.K_RSHIFT:
			key = keySelect
		}

		if key == 0 {
			break
		}

		if event.State == sdl.PRESSED {
			i.input |= key
		} else {
			// Go does not have a bitwise negation operator
			i.input &= 255 ^ key
		}

		sendController(i.input)
	}

	return true
}
