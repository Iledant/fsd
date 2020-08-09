package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var dbCopyCmd = &cobra.Command{
	Use:   "db_copy <nom de l'application>",
	Short: "Copie en local la base de données d'AWS",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			errMsg := "Nom de l'application absent"
			PrintErrMsg(errMsg)
			return errors.New(errMsg)
		}
		for _, app := range cfg.Application {
			if args[0] == app.Name {
				if app.AWSDatabase.User == "" || app.AWSDatabase.Password == "" ||
					app.AWSDatabase.Address == "" || app.AWSDatabase.Port == "" ||
					app.AWSDatabase.Name == "" || app.LocalDatabase.User == "" ||
					app.LocalDatabase.Password == "" || app.LocalDatabase.Address == "" ||
					app.LocalDatabase.Port == "" || app.LocalDatabase.Name == "" {
					errMsg := "Configuration incomplète des bases de données de l'application \"" +
						args[0] + "\" "
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
				return dbCopy(a)
			}
		}
		return nil
	},
}

func dbCopy(c fullStackCfg) error {
	tmpFile, err := ioutil.TempFile("", "db.*.dump")
	if err != nil {
		PrintErrMsg("Impossible de créer le fichier provisoire")
		return err
	}
	defer os.Remove(tmpFile.Name())

	PrintSuccessMsg("Dump de la base Local")
	awsPostgresString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
		c.AWSDatabase.User, c.AWSDatabase.Password, c.AWSDatabase.Address,
		c.AWSDatabase.Port, c.AWSDatabase.Name)
	cmd := exec.Command(cfg.PostgreSQLPath+`pg_dump.exe`,
		"-d", awsPostgresString, "-Fc", "-f", tmpFile.Name())
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

	PrintSuccessMsg("Restauration de la base AWS")
	localPostgresString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
		c.LocalDatabase.User, c.LocalDatabase.Password, c.LocalDatabase.Address,
		c.LocalDatabase.Port, c.LocalDatabase.Name)
	cmd = exec.Command(cfg.PostgreSQLPath+`pg_restore.exe`,
		"-d", localPostgresString, "-cO", tmpFile.Name())
	out, err = cmd.Output()
	if err != nil {
		var errMsg *exec.ExitError
		if errors.As(err, &errMsg) {
			if strings.Contains(string(errMsg.Stderr), "FATAL") {
				PrintErrMsg(string(errMsg.Stderr))
				return err
			}
			fmt.Println(string(errMsg.Stderr))
		} else {
			PrintErrMsg("Erreur d'exécution de pg_restore : " + err.Error())
			return err
		}
	}
	if len(out) > 0 {
		fmt.Print(string(out))
	} else {
		fmt.Println("Fin de la restauration")
	}

	if err = tmpFile.Close(); err != nil {
		PrintErrMsg("Impossible de fermer le fichier provisoire")
		return err
	}
	return nil
}
