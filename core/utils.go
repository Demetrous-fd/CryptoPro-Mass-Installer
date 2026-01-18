package core

import (
	"errors"
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

type SettingsDefaultBlock struct {
	NamePattern *string `json:"namePattern,omitempty"`
	PfxPassword *string `json:"pfxPassword,omitempty"`
	Exportable  *bool   `json:"exportable,omitempty"`
}

type SettingsArgsBlock struct {
	Exportable *bool `json:"exportable,omitempty"`
	SkipRoot   *bool `json:"skipRoot,omitempty"`
	SkipWait   *bool `json:"skipWait,omitempty"`
	Debug      *bool `json:"debug,omitempty"`
}

type Settings struct {
	Default SettingsDefaultBlock
	Args    SettingsArgsBlock
	Items   *[]*ESignatureInstallParams
}

func GetRootContainersFolder(virtualDiskPath string) (string, error) {
	var rootContainersFolder string
	if runtime.GOOS == "windows" {
		if virtualDiskPath == "" {
			return rootContainersFolder, fmt.Errorf("cant create virtual disk, virtualDiskPath not set")
		}

		disk, err := CreateVirtualDisk(virtualDiskPath)
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

func GetFilePath(path string, certPath string) (string, error) {
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

func IsContainerName(path string) bool {
	result := CONTAINER_PATTERN.FindString(path)
	return result != ""
}

func ClearDoubleSlashes(text string) string {
	doubleSlashCount := strings.Count(text, `\\`)
	if doubleSlashCount < 1 {
		return text
	}
	return strings.ReplaceAll(text, `\\`, `\`)
}

func ReplaceAttrsForLogs(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey && len(groups) == 0 {
		return slog.Attr{}
	}
	return a
}

func DeclOfNum(number int) string {
	if number%100 >= 11 && number%100 <= 14 {
		return "дней"
	}

	lastDigit := number % 10
	if lastDigit == 1 {
		return "день"
	} else if lastDigit < 5 {
		return "дня"
	}
	return "дней"
}

func IsPrivateKeyMalformed(containerPath string) bool {
	files := []string{
		"header.key",
		"masks.key",
		"masks2.key",
		"primary.key",
		"primary2.key",
	}

	for _, file := range files {
		filePath := filepath.Join(containerPath, file)
		fileInfo, err := os.Stat(filePath)
		if errors.Is(err, os.ErrNotExist) {
			slog.Debug(fmt.Sprintf("Контейнер[%s] поврежден: файл[%s] не существует", containerPath, file))
			return true
		} else if fileInfo.Size() <= 0 {
			slog.Debug(fmt.Sprintf("Контейнер[%s] поврежден: файл[%s] пустой", containerPath, file))
			return true
		}
	}
	return false
}
