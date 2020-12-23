package main

import (
	"container/list"
	"log"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var rectangles = list.New()

const characterCommandCount = 41 * 48
var characters = [characterCommandCount]DrawCharacterCommand{}

func render(commands <-chan Command, inputs chan<- byte) {

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	const windowWidth = 640
	const windowHeight = 480

	window, err := sdl.CreateWindow(
		"M8",
		sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED,
		windowWidth, windowHeight,
		sdl.WINDOW_SHOWN | sdl.WINDOW_ALLOW_HIGHDPI,
	)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	drawableWidth, drawableHeight := window.GLGetDrawableSize()

	scaleX := drawableWidth / windowWidth
	scaleY := drawableHeight / windowHeight

	err = ttf.Init()
	if err != nil {
		panic(err)
	}
	font, err := ttf.OpenFont("stealth57.ttf", 16)
	if err != nil {
		panic(err)
	}
	defer func() {
		font.Close()

		for _, glyph := range glyphCache {
			glyph.Free()
		}
	}()

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}

	_ = surface.FillRect(nil, 0)
	_ = window.UpdateSurface()

	frameCount := 0
	time := sdl.GetTicks()

	var input uint8

	for {
		event := sdl.PollEvent()
		switch event := event.(type) {
		case *sdl.QuitEvent:
			return

		case *sdl.KeyboardEvent:
			var key uint8

			switch event.Keysym.Sym {
			case sdl.K_RIGHT:
				key = keyRight
			case sdl.K_LEFT:
				key = keyLeft
			case sdl.K_UP:
				key = keyUp
			case sdl.K_DOWN:
				key = keyDown
			case sdl.K_x:
				key = keyEdit
			case sdl.K_z:
				key = keyOpt
			case sdl.K_SPACE:
				key = keyStart
			case sdl.K_LSHIFT:
				key = keySelect
			}

			if key == 0 {
				break
			}

			if event.State == sdl.PRESSED {
				input |= key
			} else {
				// Go does not have a bitwise negation operator
				input &= 0b11111111 ^ key
			}

			inputs <- input
		}

		for i := 0; i < 16; i++ {
			command, ok := <- commands
			if !ok {
				break
			}

			queue(command)
		}

		_ = surface.FillRect(nil, 0)

		// Draw rectangles

		for e := rectangles.Front(); e != nil; e = e.Next() {
			command := e.Value.(DrawRectangleCommand)
			drawRectangle(command, surface, scaleX, scaleY)
		}

		// Draw characters

		for _, command := range characters {
			if command.c == 0 {
				continue
			}
			drawCharacter(command, surface, font, scaleX, scaleY)
		}

		window.UpdateSurface()

		// Determine when one second has passed
		if sdl.GetTicks() - time > 1000 {
			log.Printf("FPS: %d", frameCount);
			frameCount = 0
			time = sdl.GetTicks()
		} else {
			frameCount++
		}
	}
}

func queue(command Command) {
	switch command := command.(type) {
	case DrawRectangleCommand:
		rectangles.PushBack(command)
		if rectangles.Len() >= 128 {
			rectangles.Remove(rectangles.Front())
		}

	case DrawCharacterCommand:
		x := command.pos.x / 8
		y := command.pos.y / 10
		index := y * 41 + x

		characters[index] = command
	}
}

type glyphCacheKey struct{
	c byte
	color Color
}

var glyphCache = map[glyphCacheKey]*sdl.Surface{}

var renderRect = &sdl.Rect{}

func drawCharacter(command DrawCharacterCommand, surface *sdl.Surface, font *ttf.Font, scaleX int32, scaleY int32) {
	var err error

	cacheKey := glyphCacheKey{
		c:     command.c,
		color: command.foreground,
	}
	glyph := glyphCache[cacheKey]
	if glyph == nil {
		glyph, err = font.RenderUTF8Solid(
			string([]byte{command.c}),
			sdl.Color{
				R: command.foreground.r,
				G: command.foreground.g,
				B: command.foreground.b,
			},
		)
		if err != nil {
			panic(err)
		}
		glyphCache[cacheKey] = glyph
	}

	renderRect.X = int32(command.pos.x) * scaleX
	renderRect.Y = int32(command.pos.y) * scaleY
	renderRect.W = glyph.H * scaleX
	renderRect.H = glyph.W * scaleY
	_ = glyph.Blit(nil, surface, renderRect)
}

func drawRectangle(command DrawRectangleCommand, surface *sdl.Surface, scaleX int32, scaleY int32) {
	color := sdl.MapRGB(
		surface.Format,
		command.color.r,
		command.color.g,
		command.color.b,
	)

	renderRect.X = int32(command.pos.x) * scaleX
	renderRect.Y = int32(command.pos.y) * scaleY - 6
	renderRect.W = int32(command.size.width) * scaleX
	renderRect.H = int32(command.size.height) * scaleY

	_ = surface.FillRect(renderRect, color)
}