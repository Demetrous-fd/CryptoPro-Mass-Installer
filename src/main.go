package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/gocarina/gocsv"
	"github.com/kardianos/osext"
	"golang.org/x/exp/slog"
)

var (
	debugFlag *bool
	fastFlag  *bool
)

func init() {
	debugFlag = flag.Bool("debug", false, "Enable debug output")
	fastFlag = flag.Bool("fast", false, "Enable fast mode, can be errors")
}

func main() {
	flag.Parse()

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
		panic(err)
	}
	defer logFile.Close()

	w := io.MultiWriter(os.Stdout, logFile)
	var handler slog.Handler = slog.NewTextHandler(w, loggerOptions)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = ';'
		r.Comment = '#'
		return r
	})

	pwd, err := osext.ExecutableFolder()
	if err != nil {
		panic(err)
	}
	certPath := filepath.Join(pwd, "certs")
	_ = os.MkdirAll(certPath, os.ModePerm)

	in, err := os.Open("data.csv")
	if err != nil {
		panic(err)
	}
	defer in.Close()

	certificates := []*Cert{}
	if err := gocsv.UnmarshalFile(in, &certificates); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	for _, cert := range certificates {
		if *fastFlag {
			wg.Add(1)
			go func(cert Cert) {
				defer wg.Done()
				InstallCertificate(certPath, &cert)
			}(*cert)
		} else {
			InstallCertificate(certPath, cert)
			slog.Info("")
		}
	}

	wg.Wait()
	fmt.Print("\n\n\nУстановка сертификатов завершена, нажмите Enter:")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
}
