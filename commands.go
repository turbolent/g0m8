package main

// # M8 SLIP Serial Send command list
//
// 251 (0xFB) - Joypad key pressed state (hardware M8 only)
//    - sends the keypress state as a single byte in hardware pin order: LEFT|UP|DOWN|SELECT|START|RIGHT|OPT|EDIT
// 252 (0xFC) - Draw oscilloscope waveform command:
//    zero bytes if off - uint8 r, uint8 g, uint8 b, followed by 320 byte value array containing the waveform
// 253 (0xFD) - Draw character command:
//    12 bytes. char c, int16 x position, int16 y position, uint8 r, uint8 g, uint8 b, uint8 r_background, uint8 g_background, uint8 b_background
// 254 (0xFE) - Draw rectangle command:
//    12 bytes. int16 x position, int16 y position, int16 width, int16 height, uint8 r, uint8 g, uint8 b

import (
	"encoding/binary"
	"fmt"
)

func decodeInt16(data []byte) int16 {
	return int16(binary.LittleEndian.Uint16(data))
}

// Position
//
type Position struct {
	x int16
	y int16
}

func decodePosition(data []byte) Position {
	return Position{
		x: decodeInt16(data[0:2]),
		y: decodeInt16(data[2:4]),
	}
}

// Size
//
type Size struct {
	width  int16
	height int16
}

func decodeSize(data []byte) Size {
	return Size{
		width:  decodeInt16(data[0:2]),
		height: decodeInt16(data[2:4]),
	}
}

// Color
//
type Color struct {
	r uint8
	g uint8
	b uint8
}

func decodeColor(data []byte) Color {
	return Color{
		r: data[0],
		g: data[1],
		b: data[2],
	}
}

// Command
//
type Command interface {
	isCommand()
}

// DrawRectangleCommand
//
type DrawRectangleCommand struct {
	pos   Position
	size  Size
	color Color
}

const drawRectangleCommand = 0xFE
const drawRectangleCommandDataLength = 12

func (DrawRectangleCommand) isCommand() {}

// DrawCharacterCommand
//
type DrawCharacterCommand struct {
	c          byte
	pos        Position
	foreground Color
	background Color
}

const drawCharacterCommand = 0xFD
const drawCharacterCommandDataLength = 12

func (DrawCharacterCommand) isCommand() {}

// DrawOscilloscopeWaveformCommand
//
type DrawOscilloscopeWaveformCommand struct {
	color    Color
	waveform [320]byte
}

const drawOscilloscopeWaveformCommand = 0xFC
const drawOscilloscopeWaveformCommandMinDataLength = 1 + 3
const drawOscilloscopeWaveformCommandMaxDataLength = 1 + 3 + 320

func (DrawOscilloscopeWaveformCommand) isCommand() {}

// JoypadKeyPressedStateCommand
//
type JoypadKeyPressedStateCommand struct {
	key byte
}

const joypadKeyPressedStateCommand = 0xFB
const joypadKeyPressedStateCommandDataLength = 2

func (JoypadKeyPressedStateCommand) isCommand() {}

// decodeCommand decodes the given M8 SLIP command packet
//
func decodeCommand(data []byte) (Command, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("invalid packet: missing command")
	}

	commandByte := data[0]
	switch commandByte {
	case drawCharacterCommand:
		if length != drawCharacterCommandDataLength {
			return nil, fmt.Errorf(
				"invalid draw character packet: expected length %d, got %d",
				drawCharacterCommandDataLength,
				length,
			)
		}
		return DrawCharacterCommand{
			c:          data[1],
			pos:        decodePosition(data[2:]),
			foreground: decodeColor(data[6:]),
			background: decodeColor(data[9:]),
		}, nil

	case drawRectangleCommand:
		if length != drawRectangleCommandDataLength {
			return nil, fmt.Errorf(
				"invalid draw rectangle packet: expected length %d, got %d",
				drawRectangleCommandDataLength,
				length,
			)
		}
		return DrawRectangleCommand{
			pos:   decodePosition(data[1:]),
			size:  decodeSize(data[5:]),
			color: decodeColor(data[9:]),
		}, nil

	case drawOscilloscopeWaveformCommand:

		if length < drawOscilloscopeWaveformCommandMinDataLength {
			return nil, fmt.Errorf(
				"invalid draw oscilloscope waveform packet: expected min length %d, got %d",
				drawOscilloscopeWaveformCommandMinDataLength,
				length,
			)
		}

		if length > drawOscilloscopeWaveformCommandMaxDataLength {
			return nil, fmt.Errorf(
				"invalid draw oscilloscope waveform packet: expected max length %d, got %d",
				drawOscilloscopeWaveformCommandMaxDataLength,
				length,
			)
		}

		waveform := [320]byte{}
		copy(waveform[:], data[4:])

		return DrawOscilloscopeWaveformCommand{
			color:    decodeColor(data[1:]),
			waveform: waveform,
		}, nil

	case joypadKeyPressedStateCommand:
		if length != joypadKeyPressedStateCommandDataLength {
			return nil, fmt.Errorf(
				"invalid joypad key pressed state packet: expected length %d, got %d",
				joypadKeyPressedStateCommandDataLength,
				length,
			)
		}
		return JoypadKeyPressedStateCommand{
			key: data[1],
		}, nil

	default:
		return nil, fmt.Errorf("unknown command byte: 0x%x", commandByte)
	}
}
