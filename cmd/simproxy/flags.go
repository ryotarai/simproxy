package main

import (
	"flag"
)

func setupFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	return fs
}
