package main

// #include <termios.h>
// #include <unistd.h>
import "C"

import (
	"log"
	"os"

	"golang.org/x/sys/unix"
)

func openSerialPort(device string) *os.File {
	f, err := os.OpenFile(device, unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0666)
	if err != nil {
		log.Fatal(err)
	}

	fd := C.int(f.Fd())
	if C.isatty(fd) != 1 {
		_ = f.Close()
		log.Fatalf("device is not a TTY: %s", device)
	}

	var settings C.struct_termios
	_, err = C.tcgetattr(fd, &settings)
	if err != nil {
		_ = f.Close()
		log.Fatal(err)
	}

	C.cfmakeraw(&settings)
	_, err = C.tcsetattr(fd, C.TCSANOW, &settings)
	if err != nil {
		_ = f.Close()
		log.Fatal(err)
	}

	return f
}
