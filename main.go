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

// TODO : faire une liste des applications dans le fichier de configuration
