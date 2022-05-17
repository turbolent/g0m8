package main

import (
	"math"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const logicalWidth = 320
const logicalHeight = 240

type sdlGlyphCacheKey struct {
	c  byte
	fg Color
	bg Color
}

type sdlGlyph struct {
	texture *sdl.Texture
	width   int32
	height  int32
}

type sdlRenderer struct {
	backgroundColor     Color
	fullscreen          bool
	window              *sdl.Window
	renderer            *sdl.Renderer
	font                *ttf.Font
	glyphCache          map[sdlGlyphCacheKey]sdlGlyph
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

	err = ttf.Init()
	if err != nil {
		panic(err)
	}
	r.font, err = ttf.OpenFont("m8stealth57.ttf", 8)
	if err != nil {
		panic(err)
	}

	r.glyphCache = make(map[sdlGlyphCacheKey]sdlGlyph)

	return r
}

func (r *sdlRenderer) quit() {
	_ = r.renderer.Destroy()
	_ = r.window.Destroy()
	r.font.Close()

	for _, glyph := range r.glyphCache {
		_ = glyph.texture.Destroy()
	}

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

	cacheKey := sdlGlyphCacheKey{
		c:  command.c,
		fg: command.foreground,
		bg: command.background,
	}
	g := r.glyphCache[cacheKey]
	if g.texture == nil {
		var glyphSurface *sdl.Surface
		var err error

		if command.foreground == command.background {
			glyphSurface, err = r.font.RenderGlyphSolid(
				rune(command.c),
				sdl.Color{
					R: command.foreground.r,
					G: command.foreground.g,
					B: command.foreground.b,
					A: math.MaxUint8,
				},
			)
		} else {
			glyphSurface, err = r.font.RenderUTF8Shaded(
				string(command.c),
				sdl.Color{
					R: command.foreground.r,
					G: command.foreground.g,
					B: command.foreground.b,
					A: math.MaxUint8,
				},
				sdl.Color{
					R: command.background.r,
					G: command.background.g,
					B: command.background.b,
					A: math.MaxUint8,
				},
			)
		}

		if err != nil {
			panic(err)
		}
		texture, _ := renderer.CreateTextureFromSurface(glyphSurface)
		g.texture = texture
		g.width = glyphSurface.W
		g.height = glyphSurface.H
		r.glyphCache[cacheKey] = g
		glyphSurface.Free()
	}

	var renderRect = sdl.Rect{
		X: int32(command.pos.x),
		Y: int32(command.pos.y) + 3,
		W: g.width,
		H: g.height,
	}

	_ = renderer.Copy(g.texture, nil, &renderRect)
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
