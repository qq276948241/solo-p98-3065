package main

import (
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

type coord struct {
	X short
	Y short
}

type smallRect struct {
	Left   short
	Top    short
	Right  short
	Bottom short
}

type consoleScreenBufferInfo struct {
	Size              coord
	CursorPosition    coord
	Attributes        word
	Window            smallRect
	MaximumWindowSize coord
}

const (
	stdOutputHandle = ^uintptr(11) + 1
)

var (
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	procSetConsoleCursorPosition   = kernel32.NewProc("SetConsoleCursorPosition")
	procFillConsoleOutputCharacterW = kernel32.NewProc("FillConsoleOutputCharacterW")
	procFillConsoleOutputAttribute  = kernel32.NewProc("FillConsoleOutputAttribute")
)

var stdoutHandle uintptr

func InitRender() {
	stdoutHandle, _, _ = procGetStdHandle.Call(stdOutputHandle)
}

func ClearScreen() {
	var csbi consoleScreenBufferInfo
	procGetConsoleScreenBufferInfo.Call(
		stdoutHandle,
		uintptr(unsafe.Pointer(&csbi)),
	)
	consoleSize := dword(csbi.Size.X) * dword(csbi.Size.Y)
	home := coord{X: 0, Y: 0}
	var written dword
	procFillConsoleOutputCharacterW.Call(
		stdoutHandle,
		uintptr(' '),
		uintptr(consoleSize),
		*(*uintptr)(unsafe.Pointer(&home)),
		uintptr(unsafe.Pointer(&written)),
	)
	procFillConsoleOutputAttribute.Call(
		stdoutHandle,
		uintptr(csbi.Attributes),
		uintptr(consoleSize),
		*(*uintptr)(unsafe.Pointer(&home)),
		uintptr(unsafe.Pointer(&written)),
	)
	procSetConsoleCursorPosition.Call(
		stdoutHandle,
		*(*uintptr)(unsafe.Pointer(&home)),
	)
}

func MoveCursor(x, y int) {
	pos := coord{X: short(x), Y: short(y)}
	procSetConsoleCursorPosition.Call(
		stdoutHandle,
		*(*uintptr)(unsafe.Pointer(&pos)),
	)
}

func WriteString(s string) {
	if len(s) == 0 {
		return
	}
	buf := syscall.StringToUTF16(s)
	var written dword
	procWriteConsoleW.Call(
		stdoutHandle,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)-1),
		uintptr(unsafe.Pointer(&written)),
		0,
	)
}

func Render(snake *Snake, food *Food, score, highScore, speedLevel int, gameOver bool) {
	MoveCursor(0, 0)

	var sb strings.Builder

	topBorder := "+" + strings.Repeat("-", MapWidth) + "+"
	sb.WriteString(topBorder)
	for i := 0; i < 3; i++ {
		sb.WriteString(" ")
	}
	switch speedLevel {
	case 1:
		sb.WriteString("═══ 贪吃蛇 ═══")
	case 2:
		sb.WriteString("═★═ 贪吃蛇 ═★═")
	default:
		sb.WriteString("═♛═ 贪吃蛇 ═♛═")
	}
	sb.WriteString("\r\n")

	for y := 0; y < MapHeight; y++ {
		sb.WriteString("|")
		for x := 0; x < MapWidth; x++ {
			ch := " "
			if x == snake.Head().X && y == snake.Head().Y {
				ch = "@"
			} else if snake.Occupies(Point{X: x, Y: y}) {
				ch = "o"
			} else if x == food.Pos.X && y == food.Pos.Y {
				ch = "*"
			}
			sb.WriteString(ch)
		}
		sb.WriteString("|")

		for i := 0; i < 3; i++ {
			sb.WriteString(" ")
		}

		switch y {
		case 2:
			sb.WriteString("当前分: " + strconv.Itoa(score))
		case 4:
			sb.WriteString("最高分: " + strconv.Itoa(highScore))
		case 6:
			sb.WriteString("速度档: Lv." + strconv.Itoa(speedLevel))
		case 10:
			sb.WriteString("↑ ↓ ← → 移动")
		case 12:
			sb.WriteString("R 重开  Q 退出")
		case 16:
			if gameOver {
				sb.WriteString("  ⚠ 游戏结束! ⚠")
			}
		case 17:
			if gameOver {
				sb.WriteString(" 按 R 重新开始")
			}
		}

		sb.WriteString("\r\n")
	}

	sb.WriteString(topBorder)
	sb.WriteString("   得分+10 / 食物\r\n")

	WriteString(sb.String())
}
