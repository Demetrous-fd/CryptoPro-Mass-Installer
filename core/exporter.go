package core

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	cades "github.com/Demetrous-fd/CryptoPro-Adapter"
	"golang.org/x/exp/slog"
)

type ExportContainerParams struct {
	ContainerPath   string
	ContainerName   string
	CertificatePath string
	PfxPassword     string
	PfxLocation     string
}

func ExportContainerToPfxCLI(certFolderPath string, rootContainersFolder string, params *ExportContainerParams) error {
	if params.ContainerPath == "" {
		slog.Error("Не указан имя/путь до контейнера, используйте флаг -cont для указания имени/пути")
		return errors.New("container not set")
	}

	var container *cades.Container
	var containerBeforeExists bool
	var err error

	if !IsContainerName(params.ContainerPath) {
		containerPath, errPath := GetFilePath(params.ContainerPath, certFolderPath)
		if errPath != nil {
			params.ContainerPath = ClearDoubleSlashes(params.ContainerPath)
			container, err = GetContainer(params.ContainerPath)
			if err != nil {
				slog.Error(errPath.Error())
				slog.Error(err.Error())
				return errPath
			}
			containerBeforeExists = true

		} else {
			containerFilename := filepath.Base(containerPath)
			if params.CertificatePath == "" {
				slog.Warn("Укажите путь до сертификата через параметр `-cert <путь до сертификата>`")
				return errors.New("certificate not set")
			}

			container, err = InstallContainerFromFolder(containerPath, rootContainersFolder, params.ContainerName)
			if err != nil {
				if !errors.Is(err, cades.ErrContainerExists) {
					slog.Warn(fmt.Sprintf("Не удалось установить контейнер из папки: %s", containerFilename))
					return err
				}
				containerBeforeExists = true
			}

			slog.Debug(fmt.Sprintf("Контейнер установлен из папки[%s], Имя контейнера:'%s'", containerFilename, container.ContainerName))
		}
	} else {
		params.ContainerPath = ClearDoubleSlashes(params.ContainerPath)
		container, err = GetContainer(params.ContainerPath)
		containerBeforeExists = true
		if err != nil {
			slog.Error(err.Error())
			return err
		}
	}

	filename := filepath.Base(params.PfxLocation)
	if !strings.Contains(filename, ".pfx") {
		containerNameRaw := strings.Split(container.ContainerName, `\`)
		containerName := containerNameRaw[len(containerNameRaw)-1]
		filename = fmt.Sprintf("%s.pfx", containerName)
	}

	var pfxLocation string
	if params.PfxLocation == "" {
		pfxLocation = filepath.Join(filepath.Dir(certFolderPath), filename)
	} else {
		pfxLocation = filepath.Join(filepath.Dir(params.PfxLocation), filename)
	}

	if params.ContainerName != "" && !containerBeforeExists {
		container, err = RenameContainer(container, params.ContainerName)

		if err != nil {
			if errors.Is(err, cades.ErrContainerNotExportable) {
				slog.Debug(fmt.Sprintf("Контейнер[%s] не экспортируемый", container.ContainerName))
				return cades.ErrContainerNotExportable
			}
			slog.Debug(fmt.Sprintf("Не удалось переименовать контейнер [%s] -> [%s]", container.ContainerName, params.ContainerName))
		}
	} else if params.ContainerName != "" {
		containerStorageName := strings.ReplaceAll(container.ContainerName, `\\.\`, "")
		containerStorageName = strings.Split(containerStorageName, `\`)[0]

		location := fmt.Sprintf(`\\.\%s\%s`, containerStorageName, params.ContainerName)
		if location == container.ContainerName || location == container.UniqueContainerName {
			slog.Debug(fmt.Sprintf("RenameContainer: The new container name matches the old one. OLD[%s] = NEW[%s]", container.ContainerName, location))
		}

		m := cades.CadesManager{}
		container, err = m.CopyContainer(container, location)

		if err != nil {
			if errors.Is(err, cades.ErrContainerNotExportable) {
				slog.Debug(fmt.Sprintf("Контейнер[%s] не экспортируемый", container.ContainerName))
				return cades.ErrContainerNotExportable
			}
			slog.Debug(fmt.Sprintf("Не удалось переименовать контейнер [%s] -> [%s]", container.ContainerName, params.ContainerName))
		} else {
			defer DeleteContainer(container)
		}
	}

	if params.CertificatePath != "" {
		_, err := LinkCertWithContainer(params.CertificatePath, container.ContainerName)
		if err != nil {
			slog.Debug(fmt.Sprintf("Не удалось установить сертификат[%s] в закрытый контейнер[%s]", params.CertificatePath, container.UniqueContainerName))
		}
	}

	pfxPath, err := ExportContainerToPfx(container, pfxLocation, params.PfxPassword)
	if err != nil {
		slog.Warn(fmt.Sprintf("Не удалось экспортировать контейнер: %s", container.ContainerName))
		slog.Warn(fmt.Sprintf("Error: %s", err))
		return err
	}

	slog.Info(fmt.Sprintf("Контейнер[%s] экспортирован в файл: %s", container.ContainerName, pfxPath))

	return nil
}
