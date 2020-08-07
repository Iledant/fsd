package cmd

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var testSaveCmd = &cobra.Command{
	Use:   "test_save <nom de l'application>",
	Short: "Sauvegarde la base de l'application pour tests ultérieurs",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			errMsg := "Nom de l'application absent"
			PrintErrMsg(errMsg)
			return errors.New(errMsg)
		}
		for _, app := range cfg.Application {
			if args[0] == app.Name {
				if app.LocalDatabase.User == "" || app.LocalDatabase.Password == "" ||
					app.LocalDatabase.Address == "" || app.LocalDatabase.Port == "" ||
					app.LocalDatabase.Name == "" || app.TestRepo == "" {
					errMsg := "Configuration incomplète de l'application \"" + args[0] + "\" "
					PrintErrMsg(errMsg)
					return errors.New(errMsg)
				}
				return nil
			}
		}
		errMsg := "Impossible de trouver l'application \"" + args[0] +
			"\" dans le fichier de configuration"
		PrintErrMsg(errMsg)
		return errors.New(errMsg)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, a := range cfg.Application {
			if a.Name == args[0] {
				return testSave(a)
			}
		}
		return nil
	},
}

func testSave(c fullStackCfg) error {
	PrintSuccessMsg("Dump de la base Local")
	localPostgresString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
		c.LocalDatabase.User, c.LocalDatabase.Password, c.LocalDatabase.Address,
		c.LocalDatabase.Port, c.LocalDatabase.Name)
	cmd := exec.Command(cfg.PostgreSQLPath+`pg_dump.exe`,
		"-d", localPostgresString, "-Fc", "-f", c.TestRepo)
	out, err := cmd.Output()
	if err != nil {
		var errMsg *exec.ExitError
		if errors.As(err, &errMsg) {
			PrintErrMsg(string(errMsg.Stderr))
		} else {
			PrintErrMsg("Erreur d'exécution de pg_dump : " + err.Error())
		}
		return err
	}
	if len(out) > 0 {
		fmt.Print(string(out))
	} else {
		fmt.Println("Fin du backup")
	}
	return nil
}
