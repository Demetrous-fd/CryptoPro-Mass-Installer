package core

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	cades "github.com/Demetrous-fd/CryptoPro-Adapter"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
	"golang.org/x/text/encoding/charmap"
)

func IsCertificateExists(thumbprint string, store string) (bool, error) {
	manager := cades.CadesManager{}
	exists, err := manager.IsCertificateExists(thumbprint, store)
	slog.Debug(fmt.Sprintf("Certificate with thumbprint[%s] is exists: %v", thumbprint, exists))
	return exists, err
}

func IsCertificateWithContainerExists(thumbprint string, store string) (bool, error) {
	manager := cades.CadesManager{}
	certs, err := manager.GetCertificatesInfo(thumbprint, store)
	if err != nil {
		slog.Debug(fmt.Sprintf("Certificate with thumbprint[%s] not exists: %s", thumbprint, err))
		return false, err
	}

	slog.Debug(fmt.Sprintf("Certificate with thumbprint[%s] exists: %v", thumbprint, len(certs) > 1))

	for _, c := range certs {
		if c.ContainerLink {
			slog.Debug(fmt.Sprintf("Thumbprint[%s]: %s", c.Thumbprint, c.Container))
			container, err := GetContainer(c.Container)
			if err != nil || (container == &cades.Container{}) {
				slog.Debug(fmt.Sprintf("Thumbprint[%s] container[%s] not exists", c.Thumbprint, c.Container))
				return false, err
			} else {
				slog.Debug(fmt.Sprintf("Thumbprint[%s] container[%s] exists", c.Thumbprint, c.Container))
				return true, nil
			}
		}
	}

	return false, err
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

var cachedUserSid string

func RenameContainer(container *cades.Container, newContainerName string) (*cades.Container, error) {
	m := cades.CadesManager{}

	decoder := charmap.Windows1251.NewDecoder()
	newContainerNameUtf8, err := decoder.String(newContainerName)
	if err != nil {
		slog.Debug(fmt.Sprintf("Cant decode cp1251 string[%s]: %s", newContainerName, err))
		return m.RenameContainer(container, newContainerNameUtf8)
	}

	user, err := user.Current()

	var username string
	// user.Name в некоторых случаях отсутствует
	if user.Name == "" {
		username = strings.Split(user.Username, `\`)[1]
	} else {
		username = user.Name
	}

	if err != nil {
		slog.Debug(fmt.Sprintf("Cant get username for direct rename, use cryptopro utils: %s", err))
		return m.RenameContainer(container, newContainerNameUtf8)
	}

	if strings.Contains(container.UniqueContainerName, "REGISTRY") {
		if user.Uid != "" {
			cachedUserSid = user.Uid
		} else if cachedUserSid == "" {
			slog.Debug(fmt.Sprintf("User sid not in cache: %s", user.Name))

			userSid, _ := cades.GetUserSid(username)
			if userSid != "" {
				cachedUserSid = userSid
				slog.Debug(fmt.Sprintf("Set user sid to cache: %s -> %s", user.Name, cachedUserSid))
			}
		}

		if cachedUserSid == "" {
			slog.Debug(fmt.Sprintf("Cant get user sid for direct rename, use cryptopro utils: %s", err))
			return m.RenameContainer(container, newContainerNameUtf8)
		}

		containerNameRaw := strings.Split(container.ContainerName, `\`)
		containerName := containerNameRaw[len(containerNameRaw)-1]

		ok, err := cades.DirectRenameContainerRegistry(cachedUserSid, containerName, newContainerName)
		if !ok {
			slog.Debug(fmt.Sprintf("Error in DirectRenameContainerRegistry, use cryptopro utils: %s", err))
			return m.RenameContainer(container, newContainerNameUtf8)
		}

		return &cades.Container{
			ContainerName:       fmt.Sprintf(`\\.\REGISTRY\%s`, newContainerNameUtf8),
			UniqueContainerName: fmt.Sprintf(`\\.\REGISTRY\REGISTRY\\%s`, newContainerNameUtf8),
		}, nil
	}

	if strings.Contains(container.UniqueContainerName, "HDIMAGE") {
		ok, err := cades.DirectRenameContainerHDImage(username, container.UniqueContainerName, newContainerName)
		if err != nil {
			slog.Debug(fmt.Sprintf("Error in DirectRenameContainerHDImage, use cryptopro utils: %s", err))
			return m.RenameContainer(container, newContainerNameUtf8)
		}

		if ok {
			newContainer, err := m.GetContainer(fmt.Sprintf(`\\.\HDIMAGE\%s`, newContainerNameUtf8))
			if err != nil {
				slog.Debug(fmt.Sprintf("Error after DirectRenameContainerHDImage, use cryptopro utils: %s", err))
				return m.RenameContainer(container, newContainerNameUtf8)
			}
			return newContainer, nil
		}
	}

	return m.RenameContainer(container, newContainerNameUtf8)
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
	err := m.InstallCertificate(path, "uRoot", false)
	return err
}

func AbsorbCertificatesFromContainers() error {
	m := cades.CadesManager{}
	_, err := m.AbsorbCertificates("")
	return err
}

type ESignatureInstallParams struct {
	ContainerPath   string  `json:"containerPath" csv:"pfx,container"`
	ContainerName   string  `json:"name,omitempty"`
	CertificatePath string  `json:"certificatePath" csv:"cert"`
	PfxPassword     *string `json:"pfxPassword,omitempty" csv:"password,pfx_password,omitempty"`
	Exportable      *bool   `json:"exportable,omitempty"`
}

func InstallESignature(rootContainersFolder string, installParams *ESignatureInstallParams) error {
	slog.Debug(fmt.Sprintf("rootContainersFolder: %s, installParams: %v", rootContainersFolder, installParams))
	certificateFilename := filepath.Base(installParams.CertificatePath)
	containerFilename := filepath.Base(installParams.ContainerPath)

	if _, err := os.Stat(installParams.ContainerPath); errors.Is(err, os.ErrNotExist) {
		slog.Debug(err.Error())
		slog.Warn(fmt.Sprintf("Файл/Директория контейнера не найден: %s", installParams.ContainerPath))
		return err
	}

	if _, err := os.Stat(installParams.CertificatePath); errors.Is(err, os.ErrNotExist) {
		slog.Warn(fmt.Sprintf("Файл сертификата не найден: %s", installParams.CertificatePath))
		return err
	}

	thumbprint, err := cades.GetCertificateThumbprintFromFile(installParams.CertificatePath)
	if err != nil {
		slog.Warn(fmt.Sprintf("Не удалось получить thumbprint сертификата: %s", certificateFilename))
		return err
	}

	ok, err := IsCertificateWithContainerExists(thumbprint, "")
	if ok {
		slog.Warn(fmt.Sprintf("Контейнер и сертификат[%s] с thumbprint[%s] существует в хранилище.", certificateFilename, thumbprint))
		return err
	}

	var container *cades.Container
	if filepath.Ext(installParams.ContainerPath) != ".pfx" {
		container, err = InstallContainerFromFolder(installParams.ContainerPath, rootContainersFolder, "")
		if err != nil {
			slog.Warn(fmt.Sprintf("Не удалось установить контейнер из папки: %s", containerFilename))
			return err
		} else {
			slog.Debug(fmt.Sprintf("Контейнер установлен из папки[%s], Имя контейнера:'%s'", containerFilename, container.ContainerName))
		}

		if installParams.Exportable != nil && !*installParams.Exportable {
			id := uuid.New()
			pfxName := fmt.Sprintf("%s-temp.pfx", id.String())
			pfxPath := filepath.Join(rootContainersFolder, pfxName)

			ok, _ := LinkCertWithContainer(installParams.CertificatePath, container.ContainerName)
			if ok {
				pfxPath, err = ExportContainerToPfx(container, pfxPath, "")
				if err == nil {
					containerFilename = pfxName
					emptyPassword := ""
					installParams.PfxPassword = &emptyPassword
					installParams.ContainerPath = pfxPath
					defer os.Remove(pfxPath)

					oldContainer := container
					defer DeleteContainer(oldContainer)
				}
			}
		}
	}

	if filepath.Ext(installParams.ContainerPath) == ".pfx" {
		pfxResult, err := InstallContainerFromPfx(installParams.ContainerPath, *installParams.PfxPassword, *installParams.Exportable)
		if err != nil {
			slog.Warn(fmt.Sprintf("Не удалось установить контейнер из pfx файла %s", containerFilename))
			if strings.Contains(pfxResult.Output, "unrecognized option `-pfx") {
				slog.Warn("Установка контейнеров из pfx файлов доступна с версии КриптоПро CSP 4.0.9944 R3 (Xenocrates) от 22.02.2018.")
			}
			return err
		} else {
			slog.Debug(fmt.Sprintf("Контейнер установлен из Pfx[%s], Имя контейнера:'%s'", containerFilename, pfxResult.Container.ContainerName))
		}
		container = &pfxResult.Container
	}

	if installParams.ContainerName != "" {
		certificateRaw, _ := os.ReadFile(installParams.CertificatePath)
		certificateX509, _ := cades.LoadCertificate(certificateRaw)
		gostCertificate, _ := cades.ParseGostCertificate(certificateX509)

		oldContainerName := container.ContainerName
		newContainerName := FormatNewName(installParams.ContainerName, gostCertificate)

		newContainer, err := RenameContainer(container, newContainerName)
		if errors.Is(err, cades.ErrContainerNotExportable) {
			slog.Warn(fmt.Sprintf("Контейнер[%s] не экспортируемый", container.ContainerName))
			if filepath.Ext(installParams.ContainerPath) == ".pfx" {
				DeleteContainer(container)
			}
			return err
		} else if err != nil {
			slog.Warn(fmt.Sprintf("Не удалось переименовать контейнер [%s] -> [%s]", container.ContainerName, newContainerName))
			if filepath.Ext(installParams.ContainerPath) == ".pfx" {
				DeleteContainer(container)
			}
			return err
		} else {
			container = newContainer
			slog.Debug(fmt.Sprintf("Контейнер [%s] переименован в [%s]", oldContainerName, container.ContainerName))
		}
	}

	isCertLink, err := LinkCertWithContainer(installParams.CertificatePath, container.ContainerName)
	if err != nil || !isCertLink {
		DeleteContainer(container)
		slog.Warn(fmt.Sprintf("Не удалось установить сертификат[%s] в закрытый контейнер[%s], изменения отменены", certificateFilename, container.UniqueContainerName))
		return err
	} else {
		slog.Info(fmt.Sprintf("Сертификат[%s] установлен в закрытый контейнер:'%s'", certificateFilename, container.ContainerName))
	}
	return nil
}
