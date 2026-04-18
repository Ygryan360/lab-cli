package tui

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// ANSI color codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"

	FgBlack   = "\033[30m"
	FgRed     = "\033[31m"
	FgGreen   = "\033[32m"
	FgYellow  = "\033[33m"
	FgBlue    = "\033[34m"
	FgMagenta = "\033[35m"
	FgCyan    = "\033[36m"
	FgWhite   = "\033[37m"

	FgBrightBlack   = "\033[90m"
	FgBrightRed     = "\033[91m"
	FgBrightGreen   = "\033[92m"
	FgBrightYellow  = "\033[93m"
	FgBrightBlue    = "\033[94m"
	FgBrightMagenta = "\033[95m"
	FgBrightCyan    = "\033[96m"
	FgBrightWhite   = "\033[97m"

	BgBlack  = "\033[40m"
	BgBlue   = "\033[44m"
	BgCyan   = "\033[46m"
)

// Cursor controls
const (
	ClearScreen    = "\033[2J"
	CursorHome     = "\033[H"
	CursorHide     = "\033[?25l"
	CursorShow     = "\033[?25h"
	ClearLine      = "\033[2K"
)

func MoveCursor(row, col int) string {
	return fmt.Sprintf("\033[%d;%dH", row, col)
}

func ClearToEnd() string {
	return "\033[J"
}

// Color helpers
func Colored(color, text string) string {
	return color + text + Reset
}

func ColoredBg(fg, bg, text string) string {
	return bg + fg + text + Reset
}

func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func PadRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

func PadLeft(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// TermSize returns current terminal width and height
func TermSize() (int, int) {
	ws := &winsize{}
	ret, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(os.Stdout.Fd()),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))
	if int(ret) == -1 {
		return 80, 24
	}
	return int(ws.Col), int(ws.Row)
}

// Box draws a titled box
func Box(title string, width int) (top, bottom string) {
	inner := width - 2
	titleStr := ""
	if title != "" {
		titleStr = "─ " + Bold + title + Reset + " "
	}
	// approximate visual length (ignore escape codes)
	titleLen := len(title) + 4
	lineLen := inner - titleLen
	if lineLen < 0 {
		lineLen = 0
	}
	top = "╭" + titleStr + strings.Repeat("─", lineLen) + "╮"
	bottom = "╰" + strings.Repeat("─", inner) + "╯"
	return
}

func HLine(width int) string {
	return strings.Repeat("─", width)
}
