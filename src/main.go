package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/kardianos/osext"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
)

var (
	versionFlag            *bool
	debugFlag              *bool
	waitFlag               *bool
	containerPathArg       *string
	containerNameArg       *string
	certificatePathArg     *string
	pfxPasswordArg         *string
	pfxLocationArg         *string
	containerExportableArg *bool
	installFlagSet         *flag.FlagSet
	exporterFlagSet        *flag.FlagSet
)

func init() {
	flag.Usage = defaultHelpUsage
	versionFlag = flag.Bool("version", false, "Отобразить версию программы")
	debugFlag = flag.Bool("debug", false, "Включить отладочную информацию")
	waitFlag = flag.Bool("wait", true, "Перед выходом ожидать нажатия клавиши enter")
	containerExportableArg = flag.Bool("exportable", false, "Разрешить экспорт контейнеров")

	installFlagSet = flag.NewFlagSet("install", flag.ExitOnError)
	installFlagSet.Usage = installHelpUsage
	containerPathArg = installFlagSet.String("cont", "", "[Требуется] Путь до pfx/папки контейнера")
	certificatePathArg = installFlagSet.String("cert", "", "[Требуется] Путь до файла сертификата")
	containerNameArg = installFlagSet.String("name", "", "Название контейнера")
	pfxPasswordArg = installFlagSet.String("pfx_pass", "", "Пароль от pfx контейнера")

	exporterFlagSet = flag.NewFlagSet("export", flag.ExitOnError)
	exporterFlagSet.Usage = exporterHelpUsage
	containerPathArg = exporterFlagSet.String("cont", "", "[Требуется] Название контейнера или путь до папки")
	containerNameArg = exporterFlagSet.String("name", "", "Название контейнера")
	pfxPasswordArg = exporterFlagSet.String("pass", "", "Пароль от pfx контейнера")
	pfxLocationArg = exporterFlagSet.String("o", "", "Путь до нового pfx контейнера")
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
		switch cmd {
		case "install":
			installFlagSet.Parse(flagArgs[1:])
		case "export":
			exporterFlagSet.Parse(flagArgs[1:])
		default:
		}
	}

	if *versionFlag {
		fmt.Println("Mass version 1.2.0")
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

	w := io.MultiWriter(os.Stdout, logFile)
	var handler slog.Handler = slog.NewTextHandler(w, loggerOptions)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	pwd, err := osext.ExecutableFolder()
	if err != nil {
		code = 1
		slog.Error(err.Error())
		return
	}
	certPath := filepath.Join(pwd, "certs")
	_ = os.Mkdir(certPath, os.ModePerm)

	rootContainersFolder, err := getRootContainersFolder(certPath)
	if err != nil {
		code = 1
		slog.Error(err.Error())
		return
	}
	if runtime.GOOS == "windows" {
		defer deleteVirtualDisk(rootContainersFolder)
	}

	// Parsing subcommand flags
	if slices.Contains(flagArgs, "export") {
		exportParams := &ExportContainerParams{
			ContainerPath: *containerPathArg,
			ContainerName: *containerNameArg,
			PfxPassword:   *pfxPasswordArg,
			PfxLocation:   *pfxLocationArg,
		}
		err := exportContainerToPfxCLI(certPath, rootContainersFolder, exportParams)
		if err != nil {
			code = 2
			slog.Error(err.Error())
			return
		}

	} else if slices.Contains(flagArgs, "install") {
		installParams := &ESignatureInstallParams{
			ContainerPath:   *containerPathArg,
			ContainerName:   *containerNameArg,
			CertificatePath: *certificatePathArg,
			PfxPassword:     *pfxPasswordArg,
			Exportable:      *containerExportableArg,
		}
		err := installESignatureCLI(certPath, rootContainersFolder, installParams, false)
		if err != nil {
			code = 2
			slog.Error(err.Error())
			return
		}
	} else {
		installESignatureFromFile(certPath, rootContainersFolder, *waitFlag, *containerExportableArg)
	}
}
