package main

import (
	"container/list"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var rectangles = list.New()

const characterCommandCount = 41 * 48

var characters = [characterCommandCount]DrawCharacterCommand{}

var waveform DrawOscilloscopeWaveformCommand

func render(commands <-chan Command, inputs chan<- byte) {

	go func() {
		for {
			queue(<-commands)
		}
	}()

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	const windowWidth = 640
	const windowHeight = 460

	window, err := sdl.CreateWindow(
		"M8",
		sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED,
		windowWidth, windowHeight,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED | sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()

	format, err := window.GetPixelFormat()
	if err != nil {
		panic(err)
	}

	background, err := renderer.CreateTexture(format, sdl.TEXTUREACCESS_TARGET, windowWidth, windowHeight)
	if err != nil {
		panic(err)
	}
	defer background.Destroy()

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
			_ = glyph.texture.Destroy()
		}
	}()

	frameCount := 0
	ticks := sdl.GetTicks()

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
				input |= key
			} else {
				// Go does not have a bitwise negation operator
				input &= 0b11111111 ^ key
			}

			inputs <- input
		}

		_ = renderer.SetDrawColor(0, 0, 0, 255)
		_ = renderer.Clear()

		// Draw rectangles

		_ = renderer.SetRenderTarget(background)

		for e := rectangles.Front(); e != nil; e = e.Next() {
			command := e.Value.(DrawRectangleCommand)
			drawRectangle(command, renderer)
		}

		_ = renderer.SetRenderTarget(nil)
		_ = renderer.Copy(background, nil, nil)

		// Draw waveform
		drawWaveform(waveform, renderer)

		// Draw characters

		for _, command := range characters {
			if command.c == 0 {
				continue
			}
			drawCharacter(command, renderer, font)
		}

		renderer.Present()

		// Determine when one second has passed
		if sdl.GetTicks()-ticks > 1000 {
			//log.Printf("FPS: %d", frameCount)
			frameCount = 0
			ticks = sdl.GetTicks()
		} else {
			frameCount++
		}
	}
}

func queue(command Command) {
	switch command := command.(type) {
	case DrawRectangleCommand:
		rectangles.PushBack(command)
		if rectangles.Len() >= 1024 {
			rectangles.Remove(rectangles.Front())
		}

	case DrawCharacterCommand:
		x := command.pos.x / 8
		y := command.pos.y / 10
		index := y*41 + x

		characters[index] = command

	case DrawOscilloscopeWaveformCommand:
		waveform = command
	}
}

type glyphCacheKey struct {
	c     byte
	color Color
}

type glyph struct {
	texture *sdl.Texture
	width   int32
	height  int32
}

var glyphCache = map[glyphCacheKey]*glyph{}

var renderRect = &sdl.Rect{}

func drawCharacter(command DrawCharacterCommand, renderer *sdl.Renderer, font *ttf.Font) {
	cacheKey := glyphCacheKey{
		c:     command.c,
		color: command.foreground,
	}
	g := glyphCache[cacheKey]
	if g == nil {
		glyphSurface, err := font.RenderUTF8Solid(
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
		texture, _ := renderer.CreateTextureFromSurface(glyphSurface)
		g = &glyph{
			texture: texture,
			width: glyphSurface.W,
			height: glyphSurface.H,
		}
		glyphCache[cacheKey] = g
		glyphSurface.Free()
	}

	renderRect.X = int32(command.pos.x) * 2
	renderRect.Y = int32(command.pos.y) * 2
	renderRect.W = g.width
	renderRect.H = g.height
	_ = renderer.Copy(g.texture, nil, renderRect)
}

func drawRectangle(command DrawRectangleCommand, renderer *sdl.Renderer) {

	_ = renderer.SetDrawColor(
		command.color.r,
		command.color.g,
		command.color.b,
		255,
	)

	renderRect.X = int32(command.pos.x) * 2
	renderRect.Y = int32(command.pos.y) * 2 - 6
	renderRect.W = int32(command.size.width) * 2
	renderRect.H = int32(command.size.height) * 2

	_ = renderer.FillRect(renderRect)
}

func drawWaveform(command DrawOscilloscopeWaveformCommand, renderer *sdl.Renderer) {

	_ = renderer.SetDrawColor(
		command.color.r,
		command.color.g,
		command.color.b,
		255,
	)

	for x, y := range waveform.waveform {

		renderRect.X = int32(x) * 2
		renderRect.Y = int32(y) * 2
		renderRect.W = 2
		renderRect.H = 2

		_ = renderer.FillRect(renderRect)
	}
}