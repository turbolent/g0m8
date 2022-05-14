package main

import (
	"fmt"
)

const slipEnd = 0xC0
const slipEsc = 0xDB
const slipEscEnd = 0xDC
const slipEscEsc = 0xDD

func decodeSLIP(data []byte, handle func(packet []byte)) (rest []byte, err error) {
	var packet []byte

	escaped := false

	var index, lastEndIndex int
	var b byte

	for index, b = range data {
		switch b {
		case slipEnd:
			lastEndIndex = index + 1
			if len(packet) > 0 {
				handle(packet)
				packet = nil
			}
			continue
		case slipEsc:
			escaped = true
			continue
		case slipEscEnd:
			if escaped {
				b = slipEnd
				escaped = false
			}
		case slipEscEsc:
			if escaped {
				b = slipEsc
				escaped = false
			}
		default:
			if escaped {
				return data, fmt.Errorf("SLIP protocol error")
			}
		}
		packet = append(packet, b)
	}

	return data[lastEndIndex:], nil
}
