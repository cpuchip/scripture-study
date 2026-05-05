//go:build !windows

package main

// configureUTF8Stdout is a no-op on platforms where stdout is already
// UTF-8 by default.
func configureUTF8Stdout() {}
