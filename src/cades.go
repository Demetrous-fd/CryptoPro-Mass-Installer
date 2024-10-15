package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cades "github.com/Demetrous-fd/CryptoPro-Adapter"
	"golang.org/x/exp/slog"
)

func IsCertificateExists(thumbprint string, store string) (bool, error) {
	manager := cades.CadesManager{}
	exists, err := manager.IsCertificateExists(thumbprint, store)
	slog.Debug(fmt.Sprintf("Certificate with thumbprint is exists: %v", exists))
	return exists, err
}

func DeleteCertificate(thumbprint string) bool {
	manager := cades.CadesManager{}
	result, _ := manager.DeleteCertificate(thumbprint)
	return result
}

func DeleteContainer(container *cades.Container) bool {
	manager := cades.CadesManager{}
	result, _ := manager.DeleteContainer(container)
	return result
}

type DeleteCertificateResult struct {
	Certificate bool `json:"certificate"`
	Container   bool `json:"container"`
}

func DeleteESignature(thumbprint string, container *cades.Container) *DeleteCertificateResult {
	result := &DeleteCertificateResult{}
	result.Certificate = DeleteCertificate(thumbprint)
	result.Container = DeleteContainer(container)
	return result
}

func InstallContainerFromPfx(path string, password string, exportable bool) (*cades.InstallPfxResult, error) {
	m := cades.CadesManager{}
	result, err := m.InstallPfx(path, password, exportable)
	slog.Debug(fmt.Sprintf("Install Pfx result: %+v", result))

	return result, err
}

func InstallContainerFromFolder(path string, rootContainersFolder string, containerName string) (*cades.Container, error) {
	m := cades.CadesManager{}
	result, err := m.InstallContainerFromFolder(path, rootContainersFolder, "", containerName)
	slog.Debug(fmt.Sprintf("Install Container from folder result: %+v", result))
	return result, err
}

func RenameContainer(container *cades.Container, newContainerName string) (*cades.Container, error) {
	m := cades.CadesManager{}
	result, err := m.RenameContainer(container, newContainerName)
	return result, err
}

func GetContainer(containerName string) (*cades.Container, error) {
	m := cades.CadesManager{}
	result, err := m.GetContainer(containerName)
	return result, err
}

func ExportContainerToPfx(container *cades.Container, filePath string, password string) (string, error) {
	m := cades.CadesManager{}

	_, err := m.ExportContainerToPfx(filePath, container.UniqueContainerName, password)
	if err != nil {
		slog.Debug(fmt.Sprintf("[Folder to pfx] Не удалось экспортировать контейнер[%s] в pfx[%s] файл, error: %s", container.ContainerName, filePath, err))
		return "", err
	}

	slog.Debug(fmt.Sprintf("[Folder to pfx] Контейнер[%s] экспортирован в pfx[%s] файл", container.ContainerName, filePath))
	return filePath, nil
}

func LinkCertWithContainer(path, containerName string) (bool, error) {
	m := cades.CadesManager{}
	result, err := m.LinkCertWithContainer(path, containerName)
	slog.Debug(fmt.Sprintf("Link certificate with container result: %+v", result))

	return result, err
}

func InstallRootCertificate(path string) error {
	m := cades.CadesManager{}
	err := m.InstallCertificate(path, "uRoot")
	return err
}

type ESignatureInstallParams struct {
	ContainerPath   string `csv:"pfx,container"`
	ContainerName   string
	CertificatePath string `csv:"cert"`
	PfxPassword     string `csv:"password,pfx_password,omitempty"`
	Exportable      bool
}

func InstallESignature(rootContainersFolder string, installParams *ESignatureInstallParams) {
	slog.Debug(fmt.Sprintf("rootContainersFolder: %s, installParams: %v", rootContainersFolder, installParams))
	certificateFilename := filepath.Base(installParams.CertificatePath)
	containerFilename := filepath.Base(installParams.ContainerPath)
	var err error

	if _, err := os.Stat(installParams.ContainerPath); errors.Is(err, os.ErrNotExist) {
		slog.Debug(err.Error())
		slog.Error(fmt.Sprintf("Файл/Директория контейнера не найден: %s", installParams.ContainerPath))
		return
	}

	if _, err := os.Stat(installParams.CertificatePath); errors.Is(err, os.ErrNotExist) {
		slog.Error(fmt.Sprintf("Файл сертификата не найден: %s", installParams.CertificatePath))
		return
	}

	thumbprint, err := cades.GetCertificateThumbprintFromFile(installParams.CertificatePath)
	if err != nil {
		slog.Error(fmt.Sprintf("Не удалось получить thumbprint сертификата: %s", certificateFilename))
		return
	}

	ok, _ := IsCertificateExists(thumbprint, "")
	if ok {
		slog.Warn(fmt.Sprintf("Сертификат[%s] с thumbprint:%s существует в хранилище.", certificateFilename, thumbprint))
		return
	}

	var container *cades.Container
	if !strings.Contains(containerFilename, ".pfx") {
		container, err = InstallContainerFromFolder(installParams.ContainerPath, rootContainersFolder, "")
		if err != nil {
			slog.Error(fmt.Sprintf("Не удалось установить контейнер из папки: %s", containerFilename))
			return
		} else {
			slog.Debug(fmt.Sprintf("Контейнер установлен из папки[%s], Имя контейнера:'%s'", containerFilename, container.ContainerName))
		}

		if !installParams.Exportable {
			containerName := strings.Trim(certificateFilename, ".cer")
			pfxName := fmt.Sprintf("%s-temp.pfx", containerName)
			pfxPath := filepath.Join(rootContainersFolder, pfxName)

			container, err = RenameContainer(container, containerName)
			if err != nil {
				slog.Debug(fmt.Sprintf("Не удалось переименовать контейнер [%s] -> [%s]", container.ContainerName, containerName))
			}

			pfxPath, err = ExportContainerToPfx(container, pfxPath, "")
			if err == nil {
				containerFilename = pfxName
				installParams.PfxPassword = ""
				installParams.ContainerPath = pfxPath
				DeleteContainer(container)

				defer os.Remove(pfxPath)
			}
		}
	}

	if strings.Contains(containerFilename, ".pfx") {
		pfxResult, err := InstallContainerFromPfx(installParams.ContainerPath, installParams.PfxPassword, installParams.Exportable)
		if err != nil {
			slog.Error(fmt.Sprintf("Не удалось установить контейнер из pfx файла %s", containerFilename))
			return
		} else {
			slog.Debug(fmt.Sprintf("Контейнер установлен из Pfx[%s], Имя контейнера:'%s'", containerFilename, pfxResult.Container.ContainerName))
		}
		container = &pfxResult.Container
	}

	var newContainerName string
	oldContainerName := container.ContainerName
	if installParams.ContainerName == "" {
		newContainerName = strings.Replace(certificateFilename, ".cer", "", 1)
	} else {
		newContainerName = installParams.ContainerName
	}

	newContainer, err := RenameContainer(container, newContainerName)
	if err != nil {
		slog.Error(fmt.Sprintf("Не удалось переименовать контейнер [%s] -> [%s]", container.ContainerName, newContainerName))
		DeleteContainer(container)
		return
	}
	container = newContainer
	slog.Debug(fmt.Sprintf("Контейнер [%s] переименован в [%s]", oldContainerName, container.ContainerName))

	isCertLink, err := LinkCertWithContainer(installParams.CertificatePath, container.ContainerName)
	if err != nil || !isCertLink {
		DeleteContainer(container)
		slog.Error(fmt.Sprintf("Не удалось установить сертификат[%s] в закрытый контейнер[%s], изменения отменены", certificateFilename, container.UniqueContainerName))
		return
	} else {
		slog.Info(fmt.Sprintf("Сертификат[%s] установлен в закрытый контейнер:'%s'", certificateFilename, container.ContainerName))
	}
}
