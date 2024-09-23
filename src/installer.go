package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gocarina/gocsv"
	"golang.org/x/exp/slog"
)

func installESignatureFromFile(certPath string, rootContainersFolder string, waitFlag bool, exportable bool) {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = ';'
		r.Comment = '#'
		return r
	})

	in, err := os.Open("data.csv")
	if err != nil {
		panic(err)
	}
	defer in.Close()

	certificates := []*ESignatureInstallParams{}
	if err := gocsv.UnmarshalFile(in, &certificates); err != nil {
		panic(err)
	}

	for _, installParams := range certificates {
		containerPath := filepath.Join(certPath, installParams.ContainerPath)
		installParams.ContainerPath = containerPath

		certificatePath := filepath.Join(certPath, installParams.CertificatePath)
		installParams.CertificatePath = certificatePath
		installParams.Exportable = exportable
		InstallESignature(rootContainersFolder, installParams)
	}

	if waitFlag {
		fmt.Print("\n\n\nУстановка сертификатов завершена, нажмите Enter:")
		fmt.Scanln()
	}
}

func installESignatureCLI(certPath string, rootContainersFolder string, installParams *ESignatureInstallParams, waitFlag bool) error {
	if installParams.CertificatePath == "" {
		slog.Error("Не указан путь до сертификата, используйте флаг -cert для указания пути")
		return errors.New("certificate not set")
	}
	if installParams.ContainerPath == "" {
		slog.Error("Не указан путь до контейнера, используйте флаг -cont для указания пути")
		return errors.New("container not set")
	}

	containerPath, err := getFilePath(installParams.ContainerPath, certPath)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	installParams.ContainerPath = containerPath

	certificatePath, err := getFilePath(installParams.CertificatePath, certPath)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	installParams.CertificatePath = certificatePath

	InstallESignature(rootContainersFolder, installParams)
	if waitFlag {
		fmt.Print("\n\n\nУстановка сертификатов завершена, нажмите Enter:")
		fmt.Scanln()
	}
	return nil
}
