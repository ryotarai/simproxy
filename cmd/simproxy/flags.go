package main

import (
	"flag"
)

type CommandLineOptions struct {
	Config      string
	ShowVersion bool
}

func setupFlagSet(name string, options *CommandLineOptions) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.StringVar(&options.Config, "config", "", "Config file path")
	fs.BoolVar(&options.ShowVersion, "version", false, "Show version")
	return fs
}
