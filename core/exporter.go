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
	ContainerPath string
	ContainerName string
	PfxPassword   string
	PfxLocation   string
}

func ExportContainerToPfxCLI(certFolderPath string, rootContainersFolder string, params *ExportContainerParams) error {
	if params.ContainerPath == "" {
		slog.Error("Не указан имя/путь до контейнера, используйте флаг -cont для указания имени/пути")
		return errors.New("container not set")
	}

	var container *cades.Container
	var containerBeforeExists bool
	var err error
	isContainer := IsContainerName(params.ContainerPath)

	if !isContainer {
		containerPath, errPath := GetFilePath(params.ContainerPath, certFolderPath)
		if errPath != nil {
			params.ContainerPath = ClearDoubleSlashes(params.ContainerPath)
			container, err = GetContainer(params.ContainerPath)
			if err != nil {
				slog.Error(errPath.Error())
				slog.Error(err.Error())
				return errPath
			}
			isContainer = true

		} else {
			containerFilename := filepath.Base(containerPath)

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

	if params.ContainerName != "" {
		container, err = RenameContainer(container, params.ContainerName)

		if errors.Is(err, cades.ErrContainerNotExportable) {
			slog.Debug(fmt.Sprintf("Контейнер[%s] не экспортируемый", container.ContainerName))
			return cades.ErrContainerNotExportable
		} else if err != nil {
			slog.Debug(fmt.Sprintf("Не удалось переименовать контейнер [%s] -> [%s]", container.ContainerName, params.ContainerName))
		}
	}

	pfxPath, err := ExportContainerToPfx(container, pfxLocation, params.PfxPassword)
	if err != nil {
		slog.Warn(fmt.Sprintf("Не удалось экспортировать контейнер: %s", container.ContainerName))
		slog.Warn(fmt.Sprintf("Error: %s", err))
		return err
	}

	slog.Info(fmt.Sprintf("Контейнер[%s] экспортирован в файл: %s", params.ContainerName, pfxPath))

	m := cades.CadesManager{}
	container, err = m.GetContainer(params.ContainerName)
	if err == nil && !isContainer && !containerBeforeExists {
		DeleteContainer(container)
	}

	return nil
}
