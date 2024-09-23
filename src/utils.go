package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"golang.org/x/exp/slog"
)

var CONTAINER_PATTERN = regexp.MustCompile(`(?m)^\\\\\.\\.*\\.*$`)

func getRootContainersFolder(virtualDiskPath string) (string, error) {
	var rootContainersFolder string
	if runtime.GOOS == "windows" {
		if virtualDiskPath == "" {
			return rootContainersFolder, fmt.Errorf("cant create virtual disk, virtualDiskPath not set")
		}

		disk, err := createVirtualDisk(virtualDiskPath)
		if err != nil {
			return disk, err
		}
		slog.Debug(fmt.Sprintf("Virtual disk %s created", disk))
		rootContainersFolder = disk

	} else {
		user, err := user.Current()
		if err != nil {
			return rootContainersFolder, err
		}
		rootContainersFolder = fmt.Sprintf("/var/opt/cprocsp/keys/%s", user.Username)
	}

	return rootContainersFolder, nil
}

func getFilePath(path string, certPath string) (string, error) {
	var filePath string
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	filePath = filepath.Join(certPath, path)
	if _, err := os.Stat(filePath); err == nil {
		return filePath, nil
	}

	filePath = filepath.Join(filepath.Dir(certPath), path)
	if _, err := os.Stat(filePath); err == nil {
		return filePath, nil
	}

	return filePath, fmt.Errorf("не удалось найти контейнер: %s", path)
}

func isContainerName(path string) bool {
	result := CONTAINER_PATTERN.FindString(path)
	return result != ""
}

func clearDoubleSlashes(text string) string {
	doubleSlashCount := strings.Count(text, `\\`)
	if doubleSlashCount < 1 {
		return text
	}
	return strings.ReplaceAll(text, `\\`, `\`)
}
