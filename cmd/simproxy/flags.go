package main

import (
	"flag"
)

type CommandLineOptions struct {
	Config string
}

func setupFlagSet(name string, options *CommandLineOptions) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.StringVar(&options.Config, "config", "/etc/simproxy.yml", "Config file path")
	return fs
}
