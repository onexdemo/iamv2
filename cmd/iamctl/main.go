package main

import (
	"os"

	"github.com/onexdemo/iamv2/internal/iamctl/cmd"
)

func main() {
	if err := cmd.NewDefaultCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
