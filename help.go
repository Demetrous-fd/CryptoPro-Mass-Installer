package main

import (
	"flag"
	"fmt"
	"os"
)

func DefaultHelpUsage() {
	intro := `
Использование:
  cpmass [flags] <command> [command flags]`
	fmt.Fprintln(os.Stderr, intro)

	fmt.Fprintln(os.Stderr, "\nCommands:")
	fmt.Fprintln(os.Stderr, "  install - Установка электронной подписи")

	fmt.Fprintln(os.Stderr, "\nFlags:")
	flag.PrintDefaults()

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Запустите `cpmass <command> -h` чтобы получить справку по определенной команде\n\n")
}

func InstallHelpUsage() {
	intro := `
Использование:
  cpmass install -cont "..." -cert "..." [flags]`
	fmt.Fprintln(os.Stderr, intro)

	fmt.Fprintln(os.Stderr, "\nFlags:")
	InstallFlagSet.PrintDefaults()

	fmt.Fprintln(os.Stderr)
}
