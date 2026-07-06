//go:build windows

package main

import (
	"os"
	"strconv"

	"golang.org/x/sys/windows/registry"
)

const autoStartRegistryKey = `Software\Microsoft\Windows\CurrentVersion\Run`
const autoStartRegistryValue = "SugarMusic"

// ApplyAutoStart registers or removes the application from the Windows Run registry key.
func (a *App) ApplyAutoStart(enabled bool) error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, autoStartRegistryKey, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	if enabled {
		return key.SetStringValue(autoStartRegistryValue, strconv.Quote(exePath))
	}
	return key.DeleteValue(autoStartRegistryValue)
}
