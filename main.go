package main

import (
	"flag"
	"fmt"
	"lazydeus/CryptoMassInstall/core"
	"os"
	"path/filepath"
	"runtime"

	slogmulti "github.com/samber/slog-multi"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
)

var (
	versionFlag             *bool
	debugFlag               *bool
	waitFlag                *bool
	skipRootFlag            *bool
	containerPathInstallArg *string
	containerPathExportArg  *string
	containerNameInstallArg *string
	containerNameExportArg  *string
	certificatePathArg      *string
	pfxPasswordInstallArg   *string
	pfxPasswordExportArg    *string
	pfxLocationArg          *string
	containerExportableArg  *bool
	InstallFlagSet          *flag.FlagSet
	ExporterFlagSet         *flag.FlagSet
)

func init() {
	flag.Usage = DefaultHelpUsage
	versionFlag = flag.Bool("version", false, "Отобразить версию программы")
	debugFlag = flag.Bool("debug", false, "Включить отладочную информацию в консоли")
	waitFlag = flag.Bool("wait", true, "Перед выходом ожидать нажатия клавиши enter")
	skipRootFlag = flag.Bool("skip-root", false, "Пропустить установку корневых сертификатов")
	containerExportableArg = flag.Bool("exportable", false, "Разрешить экспорт контейнеров")

	InstallFlagSet = flag.NewFlagSet("install", flag.ExitOnError)
	InstallFlagSet.Usage = InstallHelpUsage
	containerPathInstallArg = InstallFlagSet.String("cont", "", "[Требуется] Путь до pfx/папки контейнера")
	certificatePathArg = InstallFlagSet.String("cert", "", "[Требуется] Путь до файла сертификата")
	containerNameInstallArg = InstallFlagSet.String("name", "", "Название контейнера")
	pfxPasswordInstallArg = InstallFlagSet.String("pfx_pass", "", "Пароль от pfx контейнера")

	ExporterFlagSet = flag.NewFlagSet("export", flag.ExitOnError)
	ExporterFlagSet.Usage = ExporterHelpUsage
	containerPathExportArg = ExporterFlagSet.String("cont", "", "[Требуется] Название контейнера или путь до папки")
	containerNameExportArg = ExporterFlagSet.String("name", "", "Новое название контейнера")
	pfxPasswordExportArg = ExporterFlagSet.String("pass", "", "Пароль от pfx контейнера")
	pfxLocationArg = ExporterFlagSet.String("o", "", "Путь до нового pfx контейнера")
}

func main() {
	code := 0
	defer func() {
		os.Exit(code)
	}()

	flag.Parse()
	flagArgs := flag.Args()
	if len(flagArgs) > 1 {
		cmd := flagArgs[0]
		args := flagArgs[1:]
		switch cmd {
		case "install":
			InstallFlagSet.Parse(args)
		case "export":
			ExporterFlagSet.Parse(args)
		default:
		}
	}

	if *versionFlag {
		fmt.Println("Mass version 1.4.1")
		fmt.Println("Repository: https://github.com/Demetrous-fd/CryptoPro-Mass-Installer")
		fmt.Println("Maintainer: Lazydeus (Demetrous-fd)")
		return
	}

	loggerLevel := &slog.LevelVar{}
	if *debugFlag {
		loggerLevel.Set(slog.LevelDebug)
	}
	loggerOptions := &slog.HandlerOptions{
		AddSource: *debugFlag,
		Level:     loggerLevel,
	}

	logFile, err := os.Create("logger.log")
	if err != nil {
		code = 1
		slog.Error(err.Error())
		return
	}
	defer logFile.Close()

	logger := slog.New(
		slogmulti.Fanout(
			slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug}),
			slog.NewTextHandler(os.Stdout, loggerOptions),
		),
	)
	slog.SetDefault(logger)

	pwd, err := os.Getwd()
	if err != nil {
		code = 1
		slog.Error(err.Error())
		return
	}

	certsPath := filepath.Join(pwd, "certs")
	_ = os.Mkdir(certsPath, os.ModePerm)

	rootContainersFolder, err := core.GetRootContainersFolder(certsPath)
	if err != nil {
		code = 1
		slog.Error(err.Error())
		return
	}
	if runtime.GOOS == "windows" {
		defer core.DeleteVirtualDisk(rootContainersFolder)
	}

	// Parsing subcommand flags
	if slices.Contains(flagArgs, "export") {
		exportParams := &core.ExportContainerParams{
			ContainerPath: *containerPathExportArg,
			ContainerName: *containerNameExportArg,
			PfxPassword:   *pfxPasswordExportArg,
			PfxLocation:   *pfxLocationArg,
		}
		err := core.ExportContainerToPfxCLI(certsPath, rootContainersFolder, exportParams)
		if err != nil {
			code = 2
			slog.Error(err.Error())
			return
		}

	} else if slices.Contains(flagArgs, "install") {
		installParams := &core.ESignatureInstallParams{
			ContainerPath:   *containerPathInstallArg,
			ContainerName:   *containerNameInstallArg,
			CertificatePath: *certificatePathArg,
			PfxPassword:     *pfxPasswordInstallArg,
			Exportable:      *containerExportableArg,
		}
		err := core.InstallESignatureCLI(certsPath, rootContainersFolder, installParams, false)
		if err != nil {
			code = 2
			slog.Error(err.Error())
			return
		}
	} else {
		if !*skipRootFlag {
			core.InstallRootCertificates(certsPath)
		}

		core.InstallESignatureFromFile(certsPath, rootContainersFolder, *containerExportableArg)

		if *waitFlag {
			if runtime.GOOS == "windows" {
				core.DeleteVirtualDisk(rootContainersFolder)
			} // На случай если пользователь вручную закроет окно

			fmt.Print("\n\n\nУстановка сертификатов завершена, нажмите Enter:")
			fmt.Scanln()
		}
	}
}
