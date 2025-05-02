package main

import (
	"fmt"
	"runtime/debug"
)

// outputs the version number of the main module
func version() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return fmt.Sprintf("Version: unknown")
	}

	if buildInfo.Main.Version != "" {
		return fmt.Sprintf("Version: %s\n", buildInfo.Main.Version)
	}
	return fmt.Sprintln("Version: unknown")
}
