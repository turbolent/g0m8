package main

import (
	"container/list"
	"flag"
	"log"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var widthFlag = flag.Int("width", 640, "width of the window")
var heightFlag = flag.Int("height", 460, "height of the window")

var rectangles = list.New()

const charactersPerRow = 41
const characterCommandCount = charactersPerRow * 48

var characters = [characterCommandCount]DrawCharacterCommand{}

var waveform DrawOscilloscopeWaveformCommand

var scaleX int32
var scaleY int32

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

	windowWidth := int32(*widthFlag)
	windowHeight := int32(*heightFlag)

	scaleX = windowWidth / 320
	scaleY = windowHeight / 230

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

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()
	_ = renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

	background, err := renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_TARGET, windowWidth, windowHeight)
	if err != nil {
		panic(err)
	}
	defer background.Destroy()

	overlay, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_TARGET, windowWidth, windowHeight)
	if err != nil {
		panic(err)
	}
	defer overlay.Destroy()
	_ = overlay.SetBlendMode(sdl.BLENDMODE_BLEND)

	err = ttf.Init()
	if err != nil {
		panic(err)
	}
	font, err := ttf.OpenFont("stealth57.ttf", 8)
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

	fullScreen := false

	for {
		event := sdl.PollEvent()
		switch event := event.(type) {
		case *sdl.QuitEvent:
			return

		case *sdl.KeyboardEvent:
			if event.Type == sdl.KEYUP {
				if event.Keysym.Sym == sdl.K_RETURN &&
					event.Keysym.Mod&sdl.KMOD_ALT > 0 {

					var flags uint32
					if !fullScreen {
						flags = sdl.WINDOW_FULLSCREEN_DESKTOP
					}
					_ = window.SetFullscreen(flags)
					fullScreen = !fullScreen

					break
				}

				if event.Keysym.Sym == sdl.K_q {
					return
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
			default:
				log.Printf("unknown key: %v", event.Keysym.Sym)
				break
			}

			if event.State == sdl.PRESSED {
				input |= key
			} else {
				// Go does not have a bitwise negation operator
				input &= 255 ^ key
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

		// Draw overlay: waveform and characters

		_ = renderer.SetRenderTarget(overlay)
		_ = renderer.SetDrawColor(0, 0, 0, 0)
		_ = renderer.Clear()

		drawWaveform(waveform, renderer)

		// Draw characters

		for _, command := range characters {
			if command.c == 0 {
				continue
			}
			drawCharacter(command, renderer, font)
		}

		_ = renderer.SetRenderTarget(nil)
		_ = renderer.Copy(overlay, nil, nil)

		renderer.Present()

		// Determine when one second has passed
		if sdl.GetTicks()-ticks > 1000 {
			log.Printf("FPS: %d", frameCount)
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
		index := y*charactersPerRow + x

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

var glyphCache = map[glyphCacheKey]glyph{}

var renderRect = sdl.Rect{}

func drawCharacter(command DrawCharacterCommand, renderer *sdl.Renderer, font *ttf.Font) {
	cacheKey := glyphCacheKey{
		c:     command.c,
		color: command.foreground,
	}
	g := glyphCache[cacheKey]
	if g.texture == nil {
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
		g.texture = texture
		g.width = glyphSurface.W
		g.height = glyphSurface.H
		glyphCache[cacheKey] = g
		glyphSurface.Free()
	}

	renderRect.X = int32(command.pos.x) * scaleX
	renderRect.Y = int32(command.pos.y) * scaleY
	renderRect.W = g.width * scaleX
	renderRect.H = g.height * scaleY
	_ = renderer.Copy(g.texture, nil, &renderRect)
}

func drawRectangle(command DrawRectangleCommand, renderer *sdl.Renderer) {

	_ = renderer.SetDrawColor(
		command.color.r,
		command.color.g,
		command.color.b,
		0xff,
	)

	renderRect.X = int32(command.pos.x) * scaleX
	renderRect.Y = int32(command.pos.y)*scaleY - scaleY*3
	renderRect.W = int32(command.size.width) * scaleX
	renderRect.H = int32(command.size.height) * scaleY

	_ = renderer.FillRect(&renderRect)
}

func drawWaveform(command DrawOscilloscopeWaveformCommand, renderer *sdl.Renderer) {

	_ = renderer.SetDrawColor(
		command.color.r,
		command.color.g,
		command.color.b,
		0xff,
	)

	for x, y := range waveform.waveform {

		renderRect.X = int32(x) * scaleX
		renderRect.Y = int32(y) * scaleY
		renderRect.W = scaleX
		renderRect.H = scaleY

		_ = renderer.FillRect(&renderRect)
	}
}
