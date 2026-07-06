//go:build !windows

package main

// ApplyAutoStart is a no-op on non-Windows platforms.
func (a *App) ApplyAutoStart(enabled bool) error {
	return nil
}
