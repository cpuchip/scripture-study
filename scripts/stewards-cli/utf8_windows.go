//go:build windows

package main

import (
	"syscall"
)

// configureUTF8Stdout sets the Windows console output code page to
// UTF-8 (65001) so that strings written to stdout containing em-dashes
// and other non-ASCII characters render correctly instead of mojibake.
// The OS resets the code page when the process exits.
func configureUTF8Stdout() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleOutputCP := kernel32.NewProc("SetConsoleOutputCP")
	_, _, _ = setConsoleOutputCP.Call(uintptr(65001))
}
