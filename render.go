package main

import (
	"math"

	"github.com/veandco/go-sdl2/sdl"
)

const logicalWidth = 320
const logicalHeight = 240

type sdlRenderer struct {
	backgroundColor     Color
	fullscreen          bool
	window              *sdl.Window
	renderer            *sdl.Renderer
	font                *sdl.Texture
	lastWaveformCommand DrawOscilloscopeWaveformCommand
}

func newSDLRenderer(width, height int32, software bool) *sdlRenderer {

	r := &sdlRenderer{}

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	_, _ = sdl.ShowCursor(sdl.DISABLE)

	var err error
	r.window, err = sdl.CreateWindow(
		"M8",
		sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED,
		width, height,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		panic(err)
	}

	var flags uint32 = sdl.RENDERER_ACCELERATED
	if software {
		flags = sdl.RENDERER_SOFTWARE
	}
	r.renderer, err = sdl.CreateRenderer(r.window, -1, flags)
	if err != nil {
		panic(err)
	}

	err = r.renderer.SetLogicalSize(logicalWidth, logicalHeight)
	if err != nil {
		panic(err)
	}

	r.initFont()

	return r
}

func (r *sdlRenderer) initFont() {
	surface, err := sdl.CreateRGBSurfaceWithFormat(0, fontWidth, fontHeight, 32, sdl.PIXELFORMAT_ARGB8888)
	defer surface.Free()

	pixels := surface.Pixels()

	l := int(surface.W*surface.H) / 8
	for i := 0; i < l; i++ {
		p := fontData[i]
		for j := 0; j < 8; j++ {
			var c byte
			if p&(1<<j) == 0 {
				c = math.MaxUint8
			}

			// Set all 4 color components (ARGB)
			for k := 0; k < 4; k++ {
				pixels[(i*8+j)*4+k] = c
			}
		}
	}

	r.font, err = r.renderer.CreateTextureFromSurface(surface)
	if err != nil {
		panic(err)
	}

	err = r.font.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		panic(err)
	}
}

func (r *sdlRenderer) quit() {
	_ = r.font.Destroy()
	_ = r.renderer.Destroy()
	_ = r.window.Destroy()
	sdl.Quit()
}

func (r *sdlRenderer) draw(command Command) bool {
	switch command := command.(type) {
	case DrawRectangleCommand:
		r.drawRectangle(command)

	case DrawCharacterCommand:
		r.drawCharacter(command)

	case DrawOscilloscopeWaveformCommand:
		if command.Equals(r.lastWaveformCommand) {
			return false
		}
		r.drawWaveform(command)
		r.lastWaveformCommand = command
	}

	return true
}

func (r *sdlRenderer) toggleFullscreen() {
	var flags uint32
	if !r.fullscreen {
		flags = sdl.WINDOW_FULLSCREEN
	}
	_ = r.window.SetFullscreen(flags)
	r.fullscreen = !r.fullscreen
}

func (r *sdlRenderer) render() {
	r.renderer.Present()
}

func (r *sdlRenderer) drawCharacter(command DrawCharacterCommand) {
	renderer := r.renderer

	x := int32(command.pos.x)
	y := int32(command.pos.y)

	if command.background != command.foreground {
		_ = renderer.SetDrawColor(
			command.background.r,
			command.background.g,
			command.background.b,
			math.MaxUint8,
		)

		var renderRect = sdl.Rect{
			X: x - 1,
			Y: y + 2,
			W: fontCharWidth - 1,
			H: fontCharHeight + 1,
		}
		_ = renderer.FillRect(&renderRect)
	}

	_ = r.font.SetColorMod(
		command.foreground.r,
		command.foreground.g,
		command.foreground.b,
	)

	row := command.c / fontCharsByRow
	column := command.c % fontCharsByRow

	var sourceRect = sdl.Rect{
		X: int32(column * 8),
		Y: int32(row * 8),
		W: fontCharWidth,
		H: fontCharHeight,
	}

	var renderRect = sdl.Rect{
		X: x,
		Y: y + 3,
		W: fontCharWidth,
		H: fontCharHeight,
	}

	_ = renderer.Copy(r.font, &sourceRect, &renderRect)
}

func (r *sdlRenderer) drawRectangle(command DrawRectangleCommand) {
	renderer := r.renderer

	if command.pos.x == 0 &&
		command.pos.y == 0 &&
		command.size.width == logicalWidth &&
		command.size.height == logicalHeight {

		r.backgroundColor.r = command.color.r
		r.backgroundColor.g = command.color.g
		r.backgroundColor.b = command.color.b
	}

	_ = renderer.SetDrawColor(
		command.color.r,
		command.color.g,
		command.color.b,
		0xff,
	)

	var renderRect = sdl.Rect{
		X: int32(command.pos.x),
		Y: int32(command.pos.y),
		W: int32(command.size.width),
		H: int32(command.size.height),
	}

	_ = renderer.FillRect(&renderRect)
}

func (r *sdlRenderer) drawWaveform(command DrawOscilloscopeWaveformCommand) {
	renderer := r.renderer

	renderRect := sdl.Rect{
		X: 0,
		Y: 0,
		W: logicalWidth,
		H: logicalHeight / 10,
	}

	_ = renderer.SetDrawColor(
		r.backgroundColor.r,
		r.backgroundColor.g,
		r.backgroundColor.b,
		math.MaxUint8,
	)

	_ = renderer.FillRect(&renderRect)

	_ = renderer.SetDrawColor(
		command.color.r,
		command.color.g,
		command.color.b,
		math.MaxUint8,
	)

	for x, y := range command.waveform {

		renderRect.X = int32(x)
		renderRect.Y = int32(y)
		renderRect.W = 1
		renderRect.H = 1

		_ = renderer.FillRect(&renderRect)
	}
}
