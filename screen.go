package main

const charactersPerRow = 41
const characterCommandCount = charactersPerRow * 48

type screen struct {
	characters [characterCommandCount]DrawCharacterCommand
	rectangles []DrawRectangleCommand
	waveform   DrawOscilloscopeWaveformCommand
}

func (s *screen) update(command Command) bool {
	switch command := command.(type) {
	case DrawRectangleCommand:
		s.rectangles = append(s.rectangles, command)
		return true

	case DrawCharacterCommand:
		x := command.pos.x / 8
		y := command.pos.y / 10
		index := y*charactersPerRow + x
		s.characters[index] = command
		return true

	case DrawOscilloscopeWaveformCommand:
		if command.Equals(s.waveform) {
			return false
		}
		s.waveform = command
	}

	return false
}

func (s *screen) prepare() {
	s.rectangles = s.rectangles[:]
}
