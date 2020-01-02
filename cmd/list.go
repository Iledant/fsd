package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCommand = &cobra.Command{
	Use:   "list",
	Short: "Liste les applications du fichier de configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if len(cfg.Application) == 0 {
			PrintErrMsg("Aucune application configurée")
			return
		}
		PrintSuccessMsg("Applications configurées :")
		for i, a := range cfg.Application {
			fmt.Printf("  %d : %s\n", i+1, a.Name)
		}
	},
}
