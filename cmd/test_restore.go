package cmd

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var testRestoreCmd = &cobra.Command{
	Use:   "restore_test <nom de l'application>",
	Short: "Restore la base de données de test",
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
				return testRestore(a)
			}
		}
		return nil
	},
}

func testRestore(c fullStackCfg) error {
	PrintSuccessMsg("Restauration à partir du repo")
	localPostgresString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
		c.LocalDatabase.User, c.LocalDatabase.Password, c.LocalDatabase.Address,
		c.LocalDatabase.Port, c.LocalDatabase.Name)
	cmd := exec.Command(`C:\Program Files\PostgreSQL\11\bin\pg_restore.exe`,
		"-d", localPostgresString, "-cO", c.TestRepo)
	out, err := cmd.Output()
	if err != nil {
		var errMsg *exec.ExitError
		if errors.As(err, &errMsg) {
			PrintErrMsg(string(errMsg.Stderr))
		} else {
			PrintErrMsg("Erreur d'exécution de pg_restore : " + err.Error())
		}
		return err
	}
	if len(out) > 0 {
		fmt.Print(string(out))
	} else {
		fmt.Println("Restauration terminée")
	}
	return nil
}
