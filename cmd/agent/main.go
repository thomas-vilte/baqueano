package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func main() {
	flags := flag.NewFlagSet("baqueano", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	showVersion := flags.Bool("version", false, "Print version and exit")

	if err := flags.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		os.Exit(1)
	}

	if flags.NArg() != 0 {
		flags.Usage()
		os.Exit(1)
	}

	if !*showVersion {
		flags.Usage()
		return
	}

	if _, err := fmt.Fprintln(os.Stdout, "baqueano dev"); err != nil {
		os.Exit(1)
	}
}
