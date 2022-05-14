package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const fontScale = 2

type glyphCacheKey struct {
	c     byte
	color Color
}

type glyph struct {
	texture *sdl.Texture
	width   int32
	height  int32
}

func createTexture(renderer *sdl.Renderer, width int32, height int32) (*sdl.Texture, error) {
	return renderer.CreateTexture(
		sdl.PIXELFORMAT_ARGB8888,
		sdl.TEXTUREACCESS_TARGET,
		width,
		height,
	)
}

type sdlRenderer struct {
	scaleX     int32
	scaleY     int32
	fullscreen bool
	window     *sdl.Window
	renderer   *sdl.Renderer
	background *sdl.Texture
	overlay    *sdl.Texture
	font       *ttf.Font
	glyphCache map[glyphCacheKey]glyph
}

func newSDLRenderer(width, height int32, software bool) *sdlRenderer {

	r := &sdlRenderer{}

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	_, _ = sdl.ShowCursor(sdl.DISABLE)

	r.scaleX = width / 320
	r.scaleY = height / 230

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
	_ = r.renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

	r.background, err = createTexture(r.renderer, width, height)
	if err != nil {
		panic(err)
	}

	r.overlay, err = createTexture(r.renderer, width, height)
	if err != nil {
		panic(err)
	}
	_ = r.overlay.SetBlendMode(sdl.BLENDMODE_BLEND)

	err = ttf.Init()
	if err != nil {
		panic(err)
	}
	r.font, err = ttf.OpenFont("m8stealth57.ttf", 8*fontScale)
	if err != nil {
		panic(err)
	}

	r.glyphCache = make(map[glyphCacheKey]glyph, characterCommandCount)

	return r
}

func (r *sdlRenderer) quit() {
	_ = r.overlay.Destroy()
	_ = r.background.Destroy()
	_ = r.renderer.Destroy()
	_ = r.window.Destroy()
	r.font.Close()

	for _, glyph := range r.glyphCache {
		_ = glyph.texture.Destroy()
	}

	sdl.Quit()
}

func (r *sdlRenderer) toggleFullscreen() {
	var flags uint32
	if !r.fullscreen {
		flags = sdl.WINDOW_FULLSCREEN_DESKTOP
	}
	_ = r.window.SetFullscreen(flags)
	r.fullscreen = !r.fullscreen
}

func (r *sdlRenderer) render(screen *screen) {
	renderer := r.renderer

	_ = renderer.SetRenderTarget(nil)

	_ = renderer.SetDrawColor(0, 0, 0, 255)
	_ = renderer.Clear()

	// Draw rectangles

	_ = renderer.SetRenderTarget(r.background)

	for e := screen.rectangles.Front(); e != nil; e = e.Next() {
		command := e.Value.(DrawRectangleCommand)
		r.drawRectangle(command)
	}

	_ = renderer.SetRenderTarget(nil)
	_ = renderer.Copy(r.background, nil, nil)

	// Draw overlay: waveform and characters

	_ = renderer.SetRenderTarget(r.overlay)
	_ = renderer.SetDrawColor(0, 0, 0, 0)
	_ = renderer.Clear()

	r.drawWaveform(screen.waveform)

	// Draw characters

	for _, command := range screen.characters {
		if command.c == 0 {
			continue
		}
		r.drawCharacter(command)
	}

	_ = renderer.SetRenderTarget(nil)
	_ = renderer.Copy(r.overlay, nil, nil)

	renderer.Present()
}

func (r *sdlRenderer) drawCharacter(command DrawCharacterCommand) {
	renderer := r.renderer

	cacheKey := glyphCacheKey{
		c:     command.c,
		color: command.foreground,
	}
	g := r.glyphCache[cacheKey]
	if g.texture == nil {
		glyphSurface, err := r.font.RenderUTF8Solid(
			string(command.c),
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
		r.glyphCache[cacheKey] = g
		glyphSurface.Free()
	}

	scaleX := r.scaleX
	scaleY := r.scaleY

	var renderRect = sdl.Rect{
		X: int32(command.pos.x) * scaleX,
		Y: int32(command.pos.y) * scaleY,
		W: g.width * scaleX / fontScale,
		H: g.height * scaleY / fontScale,
	}

	_ = renderer.Copy(g.texture, nil, &renderRect)
}

func (r *sdlRenderer) drawRectangle(command DrawRectangleCommand) {
	renderer := r.renderer

	_ = renderer.SetDrawColor(
		command.color.r,
		command.color.g,
		command.color.b,
		0xff,
	)

	scaleX := r.scaleX
	scaleY := r.scaleY

	var renderRect = sdl.Rect{
		X: int32(command.pos.x) * scaleX,
		Y: int32(command.pos.y)*scaleY - scaleY*3,
		W: int32(command.size.width) * scaleX,
		H: int32(command.size.height) * scaleY,
	}

	_ = renderer.FillRect(&renderRect)
}

func (r *sdlRenderer) drawWaveform(command DrawOscilloscopeWaveformCommand) {
	renderer := r.renderer

	_ = renderer.SetDrawColor(
		command.color.r,
		command.color.g,
		command.color.b,
		0xff,
	)

	scaleX := r.scaleX
	scaleY := r.scaleY

	var renderRect sdl.Rect

	for x, y := range command.waveform {

		renderRect.X = int32(x) * scaleX
		renderRect.Y = int32(y) * scaleY
		renderRect.W = scaleX
		renderRect.H = scaleY

		_ = renderer.FillRect(&renderRect)
	}
}
