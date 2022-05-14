package main

import (
	"container/list"
)

const charactersPerRow = 41
const characterCommandCount = charactersPerRow * 48

type screen struct {
	characters [characterCommandCount]DrawCharacterCommand
	rectangles *list.List
	waveform   DrawOscilloscopeWaveformCommand
}

func newScreen() *screen {
	return &screen{
		rectangles: list.New(),
	}
}

func (s *screen) update(command Command) bool {
	switch command := command.(type) {
	case DrawRectangleCommand:
		s.rectangles.PushBack(command)
		if s.rectangles.Len() >= 1024 {
			s.rectangles.Remove(s.rectangles.Front())
		}
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
