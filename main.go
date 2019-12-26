package main

import (
	"fmt"
	"os"

	"github.com/Iledant/fsd/cmd"
	"gopkg.in/gookit/color.v1"
)

func main() {
	if err := cmd.Execute(); err != nil {
		color.Error.Print("Erreur : " + err.Error())
		color.Reset()
		fmt.Println()
		os.Exit(1)
	}
}
