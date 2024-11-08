package core

import (
	"fmt"

	"os"
	"os/exec"

	"golang.org/x/exp/slog"
	"golang.org/x/text/encoding/charmap"
)

func ExecuteSubst(args ...string) error {
	cmd := exec.Command("subst", args...)

	stdoutStderr, err := cmd.CombinedOutput()
	slog.Debug(fmt.Sprintf("subst start with args: %q", cmd.Args))

	d := charmap.CodePage866.NewDecoder()
	data, errDecode := d.Bytes(stdoutStderr)
	if errDecode != nil {
		return errDecode
	}

	slog.Debug(fmt.Sprintf("subst output: %s", data))

	return err
}

func CreateVirtualDisk(virtualDiskPath string) (string, error) {
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var availableDiskLetter string
	var err error

	for _, letter := range letters {
		diskLetter := fmt.Sprintf("%s:", string(letter))
		if _, err := os.Stat(diskLetter); err == nil {
			continue
		}

		availableDiskLetter = diskLetter
		err = ExecuteSubst(availableDiskLetter, virtualDiskPath)
		if err == nil {
			break
		}
	}
	diskLetter := fmt.Sprintf(`%s\`, availableDiskLetter)
	return diskLetter, err
}

func DeleteVirtualDisk(diskLetter string) error {
	letter := diskLetter[:2]

	err := ExecuteSubst(letter, "/D")
	if err != nil {
		slog.Error(fmt.Sprintf("Не удалось автоматически удалить виртуальный диск, выполните команду в консоли: subst %s /D", letter))
		return err
	}
	slog.Debug(fmt.Sprintf("Virtual disk %s deleted", diskLetter))
	return err
}
