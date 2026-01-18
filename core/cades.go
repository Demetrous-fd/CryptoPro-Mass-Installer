package core

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	cades "github.com/Demetrous-fd/CryptoPro-Adapter"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
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

func RenameContainer(container *cades.Container, containerName ContainerName) (*cades.Container, error) {
	m := cades.CadesManager{}
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
		return m.RenameContainer(container, containerName.Normal)
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
			return m.RenameContainer(container, containerName.Normal)
		}

		containerNameRaw := strings.Split(container.ContainerName, `\`)
		currentContainerName := containerNameRaw[len(containerNameRaw)-1]

		ok, err := cades.DirectRenameContainerRegistry(cachedUserSid, currentContainerName, containerName.Windows1251)
		if !ok {
			slog.Debug(fmt.Sprintf("Error in DirectRenameContainerRegistry, use cryptopro utils: %s", err))
			return m.RenameContainer(container, containerName.Normal)
		}

		return &cades.Container{
			ContainerName:       fmt.Sprintf(`\\.\REGISTRY\%s`, containerName.Normal),
			UniqueContainerName: fmt.Sprintf(`\\.\REGISTRY\REGISTRY\\%s`, containerName.Normal),
		}, nil
	}

	if strings.Contains(container.UniqueContainerName, "HDIMAGE") {
		ok, err := cades.DirectRenameContainerHDImage(username, container.UniqueContainerName, containerName.Windows1251)
		if err != nil {
			slog.Debug(fmt.Sprintf("Error in DirectRenameContainerHDImage, use cryptopro utils: %s", err))
			return m.RenameContainer(container, containerName.Normal)
		}

		if ok {
			newContainer, err := m.GetContainer(fmt.Sprintf(`\\.\HDIMAGE\%s`, containerName.Normal))
			if err != nil {
				slog.Debug(fmt.Sprintf("Error after DirectRenameContainerHDImage, use cryptopro utils: %s", err))
				return m.RenameContainer(container, containerName.Normal)
			}
			return newContainer, nil
		}
	}

	return m.RenameContainer(container, containerName.Normal)
}

func GetContainer(containerName string) (*cades.Container, error) {
	m := cades.CadesManager{}
	result, err := m.GetContainer(containerName)
	return result, err
}

func ExportContainerToPfxByThumbprint(container *cades.Container, thumbprint string, filePath string, password string) (string, error) {
	m := cades.CadesManager{}

	_, err := m.ExportContainerToPfxByThumbprint(filePath, thumbprint, password)
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
		slog.Error(fmt.Sprintf("Файл/Директория контейнера не найден: %s", installParams.ContainerPath))
		return err
	}

	if _, err := os.Stat(installParams.CertificatePath); errors.Is(err, os.ErrNotExist) {
		slog.Error(fmt.Sprintf("Файл сертификата не найден: %s", installParams.CertificatePath))
		return err
	}

	thumbprint, err := cades.GetCertificateThumbprintFromFile(installParams.CertificatePath)
	if err != nil {
		slog.Error(fmt.Sprintf("Не удалось получить thumbprint сертификата[%s]", certificateFilename))
		return err
	}

	ok, err := IsCertificateWithContainerExists(thumbprint, "")
	if ok {
		slog.Warn(fmt.Sprintf("Контейнер с сертификатом[%s] существует в хранилище.", certificateFilename))
		return err
	}

	certificateRaw, err := os.ReadFile(installParams.CertificatePath)
	if err != nil {
		slog.Error(fmt.Sprintf("Не удалось прочитать сертификат[%s]", installParams.CertificatePath))
		slog.Debug(fmt.Sprintf("Cant read[%s]: %s", installParams.CertificatePath, err))
		return err
	}

	certificateX509, err := cades.LoadCertificate(certificateRaw)
	if err != nil {
		slog.Error(fmt.Sprintf("Не удалось прочитать сертификат[%s]", installParams.CertificatePath))
		slog.Debug(fmt.Sprintf("Cant load[%s]: %s", installParams.CertificatePath, err))
		return err
	}

	gostCertificate, err := cades.ParseGostCertificate(certificateX509)
	if err != nil {
		slog.Error(fmt.Sprintf("Не удалось прочитать сертификат[%s]", installParams.CertificatePath))
		slog.Debug(fmt.Sprintf("Cant parse[%s]: %s", installParams.CertificatePath, err))
		return err
	}

	containerSubject := FormatNewName("#subject.surname #subject.initials - #subject.title", gostCertificate)

	now := time.Now()
	if now.After(gostCertificate.NotAfter) {
		slog.Warn(fmt.Sprintf(
			"У сертификата[%s] истек срок действия (был действителен до %s)",
			certificateFilename, gostCertificate.NotAfter.Format("02.01.2006 15:04:05"),
		))
	}

	expireAfterHours := gostCertificate.NotAfter.Sub(now).Hours()
	if expireAfterHours > 0 {
		expireAfterDays := int(expireAfterHours / 24)
		if expireAfterDays <= 30 {
			slog.Warn(fmt.Sprintf(
				"Сертификат[%s] истекает через %d %s (действителен до %s)",
				certificateFilename, expireAfterDays, DeclOfNum(expireAfterDays),
				gostCertificate.NotAfter.Format("02.01.2006 15:04:05"),
			))
		}
	}

	var container *cades.Container
	if filepath.Ext(installParams.ContainerPath) != ".pfx" {
		if IsPrivateKeyMalformed(installParams.ContainerPath) {
			slog.Error(fmt.Sprintf("Контейнер[%s] поврежден (Владелец: %s)", containerFilename, containerSubject.Normal))
			return os.ErrInvalid
		}

		container, err = InstallContainerFromFolder(installParams.ContainerPath, rootContainersFolder, "")
		if err != nil {
			slog.Error(fmt.Sprintf("Не удалось установить контейнер[%s] (Владелец: %s)", containerFilename, containerSubject.Normal))
			return err
		} else {
			slog.Debug(fmt.Sprintf("Контейнер[%s] установлен, имя[%s]", containerFilename, container.ContainerName))
		}

		if installParams.Exportable != nil && !*installParams.Exportable {
			id := uuid.New()
			pfxName := fmt.Sprintf("%s-temp.pfx", id.String())
			pfxPath := filepath.Join(rootContainersFolder, pfxName)

			ok, _ := LinkCertWithContainer(installParams.CertificatePath, container.ContainerName)
			if ok {
				pfxPath, err = ExportContainerToPfxByThumbprint(container, thumbprint, pfxPath, "")
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
			slog.Error(fmt.Sprintf("Не удалось установить контейнер из pfx файла[%s] (Владелец: %s)", containerFilename, containerSubject.Normal))
			if strings.Contains(pfxResult.Output, "unrecognized option `-pfx") {
				slog.Warn("Установка контейнеров из pfx файлов доступна с версии КриптоПро CSP 4.0.9944 R3 (Xenocrates) от 22.02.2018.")
			}
			return err
		}

		if pfxResult.Container.ContainerName == "" && pfxResult.Container.UniqueContainerName == "" {
			slog.Error(fmt.Sprintf(
				"Не удалось установить контейнер из pfx файла[%s], отсутствует закрытый ключ (Владелец: %s)",
				containerFilename, containerSubject.Normal,
			))
			return err
		}

		slog.Debug(fmt.Sprintf("Контейнер установлен из Pfx[%s], Имя контейнера:'%s'", containerFilename, pfxResult.Container.ContainerName))
		container = &pfxResult.Container
	}

	if installParams.ContainerName != "" {
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
			slog.Error(fmt.Sprintf("Не удалось переименовать контейнер [%s] -> [%s]", container.ContainerName, newContainerName.Normal))
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
		slog.Error(fmt.Sprintf(
			"Не удалось установить сертификат[%s] в контейнер[%s], изменения отменены",
			certificateFilename, container.UniqueContainerName,
		))
		return err
	} else {
		slog.Info(fmt.Sprintf("Установлен контейнер[%s]", container.ContainerName))
	}
	return nil
}
