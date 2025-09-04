package internal

import (
	"os/user"
	"path/filepath"
	"strings"
)

func Resolve(path string) (string, error) {
	resolvedPath := ""
	user, err := user.Current()
	if err != nil {
		return resolvedPath, err
	}
	userDir := user.HomeDir

	if path == "~" {
		resolvedPath = userDir
	} else if strings.HasPrefix(path, "~/") {
		resolvedPath = filepath.Join(userDir, path[2:])
	} else {
		resolvedPath = path
	}
	absolutePath, err := filepath.Abs(resolvedPath)
	if err != nil {
		return "", err
	}

	return absolutePath, nil
}
