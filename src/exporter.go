package main

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

func exportContainerToPfxCLI(certPath string, rootContainersFolder string, params *ExportContainerParams) error {
	if params.ContainerPath == "" {
		slog.Error("Не указан имя/путь до контейнера, используйте флаг -cont для указания имени/пути")
		return errors.New("container not set")
	}

	var container *cades.Container
	var containerBeforeExists bool
	var err error
	isContainer := isContainerName(params.ContainerPath)

	if !isContainer {
		containerPath, errPath := getFilePath(params.ContainerPath, certPath)
		if errPath != nil {
			params.ContainerPath = clearDoubleSlashes(params.ContainerPath)
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
		params.ContainerPath = clearDoubleSlashes(params.ContainerPath)
		container, err = GetContainer(params.ContainerPath)
		if err != nil {
			slog.Error(err.Error())
			return err
		}
	}

	var containerName string
	if params.ContainerName == "" {
		containerNameRaw := strings.Split(container.ContainerName, `\`)
		containerName = containerNameRaw[len(containerNameRaw)-1]
	} else {
		containerName = params.ContainerName
	}

	var pfxLocation string
	if params.PfxLocation == "" {
		if isContainer {
			pfxLocation = filepath.Join(filepath.Dir(certPath), fmt.Sprintf("%s.pfx", containerName))
		} else {
			pfxLocation = filepath.Join(filepath.Dir(params.ContainerPath), fmt.Sprintf("%s.pfx", containerName))
		}
	} else {
		pfxLocation = params.PfxLocation
	}

	container, err = RenameContainer(container, containerName)
	if err != nil {
		slog.Debug(fmt.Sprintf("Не удалось переименовать контейнер [%s] -> [%s]", container.ContainerName, containerName))
	}

	pfxPath, err := ExportContainerToPfx(container, pfxLocation, params.PfxPassword)
	if err != nil {
		slog.Warn(fmt.Sprintf("Не удалось экспортировать контейнер: %s", container.ContainerName))
		slog.Warn(fmt.Sprintf("Error: %s", err))
		return err
	}

	slog.Info(fmt.Sprintf("Контейнер[%s] экспортирован в файл: %s", containerName, pfxPath))

	m := cades.CadesManager{}
	container, err = m.GetContainer(containerName)
	if err == nil && !isContainer && !containerBeforeExists {
		DeleteContainer(container)
	}

	return nil
}
