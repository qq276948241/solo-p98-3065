package main

import (
	"syscall"
	"unsafe"
)

type (
	wchar uint16
	dword uint32
	word  uint16
	short int16
)

const (
	stdInputHandle      = ^uintptr(10) + 1
	enableProcessedInput = 0x0001
	enableLineInput      = 0x0002
	enableEchoInput      = 0x0004
	enableWindowInput    = 0x0008
	enableMouseInput     = 0x0010
	enableInsertMode     = 0x0020
	enableQuickEditMode  = 0x0040
	enableExtendedFlags  = 0x0080
	keyEvent             = 0x0001
	vkUp       = 0x26
	vkDown     = 0x28
	vkLeft     = 0x25
	vkRight    = 0x27
	vkSpace    = 0x20
	vkR        = 0x52
	vkQ        = 0x51
)

type inputRecord struct {
	eventType word
	_         [2]byte
	event     [16]byte
}

type keyEventRecord struct {
	keyDown         int32
	repeatCount     word
	virtualKeyCode  word
	virtualScanCode word
	unicodeChar     wchar
	controlKeyState dword
}

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle               = kernel32.NewProc("GetStdHandle")
	procSetConsoleMode             = kernel32.NewProc("SetConsoleMode")
	procGetConsoleMode             = kernel32.NewProc("GetConsoleMode")
	procGetNumberOfConsoleInputEvents = kernel32.NewProc("GetNumberOfConsoleInputEvents")
	procPeekConsoleInputW          = kernel32.NewProc("PeekConsoleInputW")
	procReadConsoleInputW          = kernel32.NewProc("ReadConsoleInputW")
	procWriteConsoleW              = kernel32.NewProc("WriteConsoleW")
)

type Direction int

const (
	DirUp Direction = iota
	DirDown
	DirLeft
	DirRight
	DirNone
)

type KeyAction int

const (
	ActionNone KeyAction = iota
	ActionDir
	ActionRestart
	ActionQuit
	ActionPause
)

type KeyEvent struct {
	Action    KeyAction
	Direction Direction
}

var stdinHandle uintptr
var originalMode dword

func InitKeyboard() {
	stdinHandle, _, _ = procGetStdHandle.Call(stdInputHandle)
	procGetConsoleMode.Call(stdinHandle, uintptr(unsafe.Pointer(&originalMode)))
	mode := originalMode &^ (enableEchoInput | enableLineInput)
	mode |= enableWindowInput
	procSetConsoleMode.Call(stdinHandle, uintptr(mode))
}

func RestoreKeyboard() {
	procSetConsoleMode.Call(stdinHandle, uintptr(originalMode))
}

func ReadKeyEvent() KeyEvent {
	var num dword
	procGetNumberOfConsoleInputEvents.Call(stdinHandle, uintptr(unsafe.Pointer(&num)))
	if num == 0 {
		return KeyEvent{Action: ActionNone}
	}

	var ir inputRecord
	var read dword
	procReadConsoleInputW.Call(
		stdinHandle,
		uintptr(unsafe.Pointer(&ir)),
		1,
		uintptr(unsafe.Pointer(&read)),
	)

	if read == 0 || ir.eventType != keyEvent {
		return KeyEvent{Action: ActionNone}
	}

	kr := (*keyEventRecord)(unsafe.Pointer(&ir.event[0]))
	if kr.keyDown == 0 {
		return KeyEvent{Action: ActionNone}
	}

	switch kr.virtualKeyCode {
	case vkUp:
		return KeyEvent{Action: ActionDir, Direction: DirUp}
	case vkDown:
		return KeyEvent{Action: ActionDir, Direction: DirDown}
	case vkLeft:
		return KeyEvent{Action: ActionDir, Direction: DirLeft}
	case vkRight:
		return KeyEvent{Action: ActionDir, Direction: DirRight}
	case vkSpace:
		return KeyEvent{Action: ActionPause}
	case vkR:
		return KeyEvent{Action: ActionRestart}
	case vkQ:
		return KeyEvent{Action: ActionQuit}
	}

	return KeyEvent{Action: ActionNone}
}
