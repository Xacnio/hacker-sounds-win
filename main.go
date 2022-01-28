package main

import (
	"fmt"
	"syscall"
	"unicode/utf8"
	"unsafe"

	"github.com/TheTitanrain/w32"
	wi32 "github.com/jcollie/w32"
)

var (
	moduser32 = syscall.NewLazyDLL("user32.dll")

	procGetKeyboardLayout     = moduser32.NewProc("GetKeyboardLayout")
	procToUnicodeEx           = moduser32.NewProc("ToUnicodeEx")
	procGetKeyState           = moduser32.NewProc("GetKeyState")
)

func NewKeyListener() Keylistener {
	kl := Keylistener{}

	return kl
}

type Keylistener struct {
	lastKey int
}

type Key struct {
	Empty   bool
	Rune    rune
	Keycode int
}

func ValidKey(i int) bool {
	switch i {
	case 0x20, 0x0D, 0x25, 0x26, 0x27, 0x28, 0x08, 0x09,
	0x23, 0x24, 0x21, 0x22:
		return true
	}
	return false
}

func (kl *Keylistener) GetKey() Key {
	activeKey := 0
	var keyState uint16

	for i := 0; i < 256; i++ {
		keyState = w32.GetAsyncKeyState(i)

		if keyState&(1<<15) != 0 && !(i < 0x2F && !ValidKey(i)) && (i < 160 || i > 165) && (i < 91 || i > 93) {
			activeKey = i
			break
		}
	}

	if activeKey != 0 {
		kl.lastKey = activeKey
		return kl.ParseKeycode(activeKey, keyState)
	}

	return Key{Empty: true}
}

func (kl Keylistener) ParseKeycode(keyCode int, keyState uint16) Key {
	key := Key{Empty: false, Keycode: keyCode}
	outBuf := make([]uint16, 1)
	kbState := make([]uint8, 256)
	kbLayout, _, _ := procGetKeyboardLayout.Call(uintptr(0))
	if w32.GetAsyncKeyState(w32.VK_SHIFT)&(1<<15) != 0 {
		kbState[w32.VK_SHIFT] = 0xFF
	}
	capitalState, _, _ := procGetKeyState.Call(uintptr(w32.VK_CAPITAL))
	if capitalState != 0 {
		kbState[w32.VK_CAPITAL] = 0xFF
	}

	if w32.GetAsyncKeyState(w32.VK_CONTROL)&(1<<15) != 0 {
		kbState[w32.VK_CONTROL] = 0xFF
	}

	if w32.GetAsyncKeyState(w32.VK_MENU)&(1<<15) != 0 {
		kbState[w32.VK_MENU] = 0xFF
	}

	_, _, _ = procToUnicodeEx.Call(
		uintptr(keyCode),
		uintptr(0),
		uintptr(unsafe.Pointer(&kbState[0])),
		uintptr(unsafe.Pointer(&outBuf[0])),
		uintptr(1),
		uintptr(1),
		uintptr(kbLayout))

	key.Rune, _ = utf8.DecodeRuneInString(syscall.UTF16ToString(outBuf))

	return key
}

func main() {
	go func() {
		kl := NewKeyListener()
		for {
			ke := kl.GetKey()
			if ke.Keycode != 0 {
				if ke.Keycode == w32.VK_SPACE {
					wi32.PlaySound("sounds/spacebar.wav", wi32.HMODULE(0), wi32.SND_FILENAME)
				} else if ke.Keycode == w32.VK_BACK {
					wi32.PlaySound("sounds/delete.wav", wi32.HMODULE(0), wi32.SND_FILENAME)
				} else if ke.Keycode == w32.VK_RETURN {
					wi32.PlaySound("sounds/enter.wav", wi32.HMODULE(0), wi32.SND_FILENAME)
				} else if ke.Keycode >= w32.VK_LEFT && ke.Keycode <= w32.VK_DOWN {
					wi32.PlaySound("sounds/arrow.wav", wi32.HMODULE(0), wi32.SND_FILENAME)
				} else if ke.Keycode == w32.VK_TAB {
					wi32.PlaySound("sounds/tab.wav", wi32.HMODULE(0), wi32.SND_FILENAME)
				} else {
					wi32.PlaySound("sounds/key.wav", wi32.HMODULE(0), wi32.SND_FILENAME)
				}
			}
		}
	}()
	fmt.Scanln()
}