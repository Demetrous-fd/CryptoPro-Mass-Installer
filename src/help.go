package main

import (
	"flag"
	"fmt"
	"os"
)

func defaultHelpUsage() {
	intro := `
Использование:
  mass [flags] <command> [command flags]`
	fmt.Fprintln(os.Stderr, intro)

	fmt.Fprintln(os.Stderr, "\nCommands:")
	fmt.Fprintln(os.Stderr, "  install - Установка электронной подписи")
	fmt.Fprintln(os.Stderr, "  export - Экспортирование электронной подписи в pfx")

	fmt.Fprintln(os.Stderr, "\nFlags:")
	flag.PrintDefaults()

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Запустите `mass <command> -h` чтобы получить справку по определенной команде\n\n")
}

func installHelpUsage() {
	intro := `
Использование:
  mass install -cont "..." -cert "..." [flags]`
	fmt.Fprintln(os.Stderr, intro)

	fmt.Fprintln(os.Stderr, "\nFlags:")
	installFlagSet.PrintDefaults()

	fmt.Fprintln(os.Stderr)
}

func exporterHelpUsage() {
	intro := `
Использование:
  mass export -cont "..." [flags]`
	fmt.Fprintln(os.Stderr, intro)

	fmt.Fprintln(os.Stderr, "\nFlags:")
	exporterFlagSet.PrintDefaults()

	fmt.Fprintln(os.Stderr)
}