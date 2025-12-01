package core

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	cades "github.com/Demetrous-fd/CryptoPro-Adapter"
	"github.com/gocarina/gocsv"
	"golang.org/x/exp/slog"
)

func InstallESignatureFromFile(certPath string, rootContainersFolder string, settings Settings) {
	items := []*ESignatureInstallParams{}

	if settings.Items != nil && len(*settings.Items) > 0 {
		items = *settings.Items
	} else if _, err := os.Stat("data.csv"); errors.Is(err, os.ErrNotExist) {
		pair, err := FindDigitalSignaturePairs(certPath)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		items = pair
	} else {
		gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
			r := csv.NewReader(in)
			r.Comma = ';'
			r.Comment = '#'
			return r
		})

		in, err := os.Open("data.csv")
		if err != nil {
			slog.Error(err.Error())
			return
		}
		defer in.Close()

		if err := gocsv.UnmarshalFile(in, &items); err != nil {
			slog.Error(err.Error())
			return
		}
	}
	for _, installParams := range items {
		containerPath := filepath.Join(certPath, installParams.ContainerPath)
		installParams.ContainerPath = containerPath

		if installParams.ContainerName == "" && settings.Default.NamePattern != nil {
			installParams.ContainerName = *settings.Default.NamePattern
		}

		if filepath.Ext(installParams.ContainerPath) == ".pfx" && installParams.PfxPassword == nil && settings.Default.PfxPassword != nil {
			installParams.PfxPassword = settings.Default.PfxPassword
		}

		certificatePath := filepath.Join(certPath, installParams.CertificatePath)
		installParams.CertificatePath = certificatePath

		if installParams.Exportable == nil {
			installParams.Exportable = settings.Default.Exportable
		}

		InstallESignature(rootContainersFolder, installParams)
	}
}

func InstallESignatureCLI(certPath string, rootContainersFolder string, installParams *ESignatureInstallParams, waitFlag bool) error {
	if installParams.CertificatePath == "" {
		slog.Error("Не указан путь до сертификата, используйте флаг -cert для указания пути")
		return errors.New("certificate not set")
	}
	if installParams.ContainerPath == "" {
		slog.Error("Не указан путь до контейнера, используйте флаг -cont для указания пути")
		return errors.New("container not set")
	}

	containerPath, err := GetFilePath(installParams.ContainerPath, certPath)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	installParams.ContainerPath = containerPath

	certificatePath, err := GetFilePath(installParams.CertificatePath, certPath)
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

func InstallRootCertificates(certsFolderPath string) {
	rootFolder := filepath.Join(certsFolderPath, "root")
	if _, err := os.Stat(rootFolder); errors.Is(err, os.ErrNotExist) {
		slog.Debug(fmt.Sprintf("Root folder[%s] not exists", rootFolder))
		return
	}

	folderEntity, err := os.ReadDir(rootFolder)
	if err != nil {
		slog.Debug(fmt.Sprintf("Cant get entities from root folder[%s], error: %s", rootFolder, err.Error()))
		return
	}

	for _, entity := range folderEntity {
		if entity.IsDir() {
			continue
		}

		filename := entity.Name()
		path := filepath.Join(rootFolder, filename)

		if strings.HasSuffix(filename, ".p7b") {
			err = InstallRootCertificate(path)
			if err == nil {
				slog.Info(fmt.Sprintf("Корневой сертификат[%s] установлен", filename))
			}
			continue
		}

		if strings.HasSuffix(filename, ".cer") {
			thumbprint, err := cades.GetCertificateThumbprintFromFile(path)
			if err != nil {
				slog.Debug(fmt.Sprintf("Cant get thumbprint from file[%s], error: %s", path, err.Error()))
				continue
			}

			certExists, _ := IsCertificateExists(thumbprint, "uRoot")

			if !certExists {
				err = InstallRootCertificate(path)
				if err == nil {
					slog.Info(fmt.Sprintf("Корневой сертификат[%s] установлен", filename))
				}
			} else {
				slog.Debug(fmt.Sprintf("Root certificate[%s] exists in store[uRoot]", path))
			}
		}
	}
}
