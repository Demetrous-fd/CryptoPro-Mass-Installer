package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"lazydeus/CryptoMassInstall/core"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-colorable"
	slogmulti "github.com/samber/slog-multi"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
)

var (
	versionFlag             *bool
	debugFlag               *bool
	skipWaitFlag            *bool
	skipRootFlag            *bool
	containerPathInstallArg *string
	containerNameInstallArg *string
	certificatePathArg      *string
	pfxPasswordInstallArg   *string
	containerExportableArg  *bool
	InstallFlagSet          *flag.FlagSet
)

const (
	MASS_VERSION = "1.6.1"
)

func init() {
	flag.Usage = DefaultHelpUsage
	versionFlag = flag.Bool("version", false, "Отобразить версию программы")
	debugFlag = flag.Bool("debug", false, "Включить отладочную информацию в консоли")
	skipWaitFlag = flag.Bool("skip-wait", false, "Пропустить ожидание перед выходом")
	skipRootFlag = flag.Bool("skip-root", false, "Пропустить установку корневых сертификатов")
	containerExportableArg = flag.Bool("exportable", false, "Разрешить экспорт контейнеров")

	InstallFlagSet = flag.NewFlagSet("install", flag.ExitOnError)
	InstallFlagSet.Usage = InstallHelpUsage
	containerPathInstallArg = InstallFlagSet.String("cont", "", "[Требуется] Путь до pfx/папки контейнера")
	certificatePathArg = InstallFlagSet.String("cert", "", "[Требуется] Путь до файла сертификата")
	containerNameInstallArg = InstallFlagSet.String("name", "", "Название контейнера")
	pfxPasswordInstallArg = InstallFlagSet.String("pfx_pass", "", "Пароль от pfx контейнера")
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
		default:
		}
	}

	pwd, err := os.Getwd()
	if err != nil {
		code = 1
		return
	}

	var settings core.Settings
	settingsPath := filepath.Join(pwd, "settings.json")
	settingsFile, err := os.ReadFile(settingsPath)
	if err == nil {
		err = json.Unmarshal(settingsFile, &settings)
		if err == nil {
			if settings.Args.Debug != nil {
				debugFlag = settings.Args.Debug
			}

			if settings.Args.SkipWait != nil {
				skipWaitFlag = settings.Args.SkipWait
			}

			if settings.Default.Exportable != nil {
				settings.Args.Exportable = settings.Default.Exportable
			}

			if settings.Args.Exportable != nil {
				containerExportableArg = settings.Args.Exportable
			} else {
				settings.Default.Exportable = containerExportableArg
			}

			if settings.Args.SkipRoot != nil {
				skipRootFlag = settings.Args.SkipRoot
			}
		}
	}

	if *versionFlag {
		fmt.Printf("CryptoPro Mass Installer version %s\n", MASS_VERSION)
		fmt.Println("Repository: https://github.com/Demetrous-fd/CryptoPro-Mass-Installer")
		fmt.Println("Maintainer: Lazydeus (Demetrous-fd)")
		return
	}

	loggerLevel := &slog.LevelVar{}
	if *debugFlag {
		loggerLevel.Set(slog.LevelDebug)
	}

	loggerOptions := &tint.Options{
		AddSource:   *debugFlag,
		Level:       loggerLevel,
		ReplaceAttr: core.ReplaceAttrsForLogs,
	}

	logsPath := filepath.Join(pwd, "logs")
	_ = os.Mkdir(logsPath, os.ModePerm)

	now := time.Now()
	logFile, err := os.Create(filepath.Join(logsPath, fmt.Sprintf("logger-%s.log", now.Format("02-01-2006 15-04-05"))))
	if err != nil {
		code = 1
		slog.Error(err.Error())
		return
	}
	defer logFile.Close()

	var consoleLoggerHandler io.Writer
	if runtime.GOOS == "windows" {
		consoleLoggerHandler = colorable.NewColorableStdout()
	} else {
		consoleLoggerHandler = os.Stdout
	}

	logger := slog.New(
		slogmulti.Fanout(
			slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug}),
			tint.NewHandler(consoleLoggerHandler, loggerOptions),
		),
	)
	slog.SetDefault(logger)
	slog.Debug(fmt.Sprintf("CryptoPro Mass Installer version %s", MASS_VERSION))

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
	if slices.Contains(flagArgs, "install") {
		installParams := &core.ESignatureInstallParams{
			ContainerPath:   *containerPathInstallArg,
			ContainerName:   *containerNameInstallArg,
			CertificatePath: *certificatePathArg,
			PfxPassword:     pfxPasswordInstallArg,
			Exportable:      containerExportableArg,
		}
		err := core.InstallESignatureCLI(certsPath, rootContainersFolder, installParams, false)
		if err != nil {
			code = 2
			slog.Error(err.Error())
			return
		} else {
			core.AbsorbCertificatesFromContainers()
		}
	} else {
		if !*skipRootFlag {
			core.InstallRootCertificates(certsPath)
		}

		core.InstallESignatureFromFile(certsPath, rootContainersFolder, settings)
		core.AbsorbCertificatesFromContainers()
		if !*skipWaitFlag {
			if runtime.GOOS == "windows" {
				core.DeleteVirtualDisk(rootContainersFolder)
			} // На случай если пользователь вручную закроет окно

			fmt.Print("\n\n\nУстановка сертификатов завершена, нажмите Enter:")
			fmt.Scanln()
		}
	}
}
