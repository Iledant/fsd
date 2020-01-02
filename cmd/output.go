package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
)

// ANSIReset switch back uses of ANSI escapes codes
const ANSIReset = "\u001B[0m"

// ANSIGreen is the ANSI escape code for green text
const ANSIGreen = "\u001B[32m"

// ANSIYellow is the ANSI escape code for yellow text
const ANSIYellow = "\u001B[33m"

// ANSIBlack is the ANSI escape code for black text (to be used with background colors)
const ANSIBlack = "\u001B[30m"

// ANSIRed is the ANSI escape code for red text (to be used with background colors)
const ANSIRed = "\u001B[31m"

// ANSIRedBackground is the ANSI escape code for red background
const ANSIRedBackground = "\u001B[41m"

// PrintErrMsg print an error to the console using a dedicated color schema
func PrintErrMsg(s string) {
	fmt.Println(ANSIRed + s + ANSIReset)
}

// PrintSuccessMsg print a green message to the console
func PrintSuccessMsg(s string) {
	fmt.Println(ANSIGreen + s + ANSIReset)
}

func askUser(name string, required bool) (string, error) {
	fmt.Print(ANSIYellow + name + " : " + ANSIReset)
	scanner := bufio.NewScanner(os.Stdin)

	scanner.Scan()
	scan := scanner.Text()

	if required && scan == "" {
		PrintErrMsg(name + " nécessaire")
		return "", errors.New(name + " nécessaire")
	}
	return scan, nil
}
