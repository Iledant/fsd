package main

import (
	"os"

	"github.com/Iledant/fsd/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
