package tui

import (
	"os"
	"syscall"
	"unsafe"
)

type termios struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Cc     [20]uint8
	Ispeed uint32
	Ospeed uint32
}

var originalTermios termios

func EnableRawMode() error {
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(os.Stdin.Fd()),
		syscall.TCGETS,
		uintptr(unsafe.Pointer(&originalTermios))); errno != 0 {
		return errno
	}

	raw := originalTermios
	// Disable echo, canonical mode, signals
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	raw.Iflag &^= syscall.IXON | syscall.ICRNL | syscall.BRKINT | syscall.INPCK | syscall.ISTRIP
	raw.Oflag &^= syscall.OPOST
	raw.Cflag |= syscall.CS8
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0

	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(os.Stdin.Fd()),
		syscall.TCSETS,
		uintptr(unsafe.Pointer(&raw))); errno != 0 {
		return errno
	}
	return nil
}

func DisableRawMode() {
	syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(os.Stdin.Fd()),
		syscall.TCSETS,
		uintptr(unsafe.Pointer(&originalTermios)))
}

// Key constants
type Key int

const (
	KeyNone Key = iota
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyEnter
	KeyEsc
	KeyTab
	KeyBackspace
	KeyCtrlC
	KeyCtrlD
	KeyRune
)

type KeyEvent struct {
	Type Key
	Ch   rune
}

func ReadKey() KeyEvent {
	buf := make([]byte, 4)
	n, _ := os.Stdin.Read(buf)
	if n == 0 {
		return KeyEvent{Type: KeyNone}
	}

	switch {
	case n == 1 && buf[0] == 13:
		return KeyEvent{Type: KeyEnter}
	case n == 1 && buf[0] == 27:
		return KeyEvent{Type: KeyEsc}
	case n == 1 && buf[0] == 9:
		return KeyEvent{Type: KeyTab}
	case n == 1 && (buf[0] == 127 || buf[0] == 8):
		return KeyEvent{Type: KeyBackspace}
	case n == 1 && buf[0] == 3:
		return KeyEvent{Type: KeyCtrlC}
	case n == 1 && buf[0] == 4:
		return KeyEvent{Type: KeyCtrlD}
	case n == 3 && buf[0] == 27 && buf[1] == '[':
		switch buf[2] {
		case 'A':
			return KeyEvent{Type: KeyUp}
		case 'B':
			return KeyEvent{Type: KeyDown}
		case 'C':
			return KeyEvent{Type: KeyRight}
		case 'D':
			return KeyEvent{Type: KeyLeft}
		}
	case n == 1 && buf[0] >= 32 && buf[0] < 127:
		return KeyEvent{Type: KeyRune, Ch: rune(buf[0])}
	}

	return KeyEvent{Type: KeyNone}
}
