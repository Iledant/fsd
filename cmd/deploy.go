package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

var (
	deployCmd = &cobra.Command{
		Use:           "deploy <nom de l'application>",
		Short:         "Compile le backend et le frontend et le déploie sur AWS",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				errMsg := "Nom de l'application à déployer absent"
				PrintErrMsg(errMsg)
				return errors.New(errMsg)
			}
			for _, app := range cfg.Application {
				if args[0] == app.Name {
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
					return launch(a)
				}
			}
			return nil
		},
	}
	noBackend, noFrontend bool
)

type command struct {
	AppName string
	AppArgs []string
}

func launch(c fullStackCfg) error {
	c.BackEnd.Message = "Compilation du backend"
	c.BackEnd.No = noBackend
	c.FrontEnd.Message = "Compilation du frontend"
	c.FrontEnd.No = noFrontend

	bCh, fCh := launchPart(c.BackEnd), launchPart(c.FrontEnd)
	b, f := <-bCh, <-fCh

	if b != nil {
		return b
	}
	if f != nil {
		return f
	}

	PrintSuccessMsg("Déploiement")
	return launchDeploy(c.Deploy)
}

func launchPart(p partCfg) <-chan error {
	e := make(chan error)

	go func() {
		defer close(e)
		if p.No {
			e <- nil
			return
		}
		PrintSuccessMsg(p.Message)
		for _, env := range p.Environment {
			if err := os.Setenv(env.Name, env.Value); err != nil {
				e <- err
				return
			}
		}
		cmd := exec.Command(p.Command, p.Args...)
		cmd.Dir = p.Path
		out, err := cmd.Output()
		if err != nil {
			var errMsg *exec.ExitError
			if errors.As(err, &errMsg) {
				PrintErrMsg(string(errMsg.Stderr))
			} else {
				PrintErrMsg("Erreur d'exécution de" + p.Command + " : " + err.Error())
			}
			e <- err
			return
		}
		fmt.Print(out)
		e <- nil
	}()

	return e
}

func getEBVersion() string {
	cmd := exec.Command("eb", "status")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	idx := strings.Index(string(out), "Deployed Version:")
	if idx == -1 {
		return ""
	}
	extract := out[idx+18:]
	var i, j int
	for i = 0; i < len(extract); i++ {
		if extract[i] == '\n' {
			j = i - 1
			break
		}
	}
	if j <= 0 {
		return ""
	}
	return string(extract[:j])
}

func getGitVersion() string {
	cmd := exec.Command("git", "describe")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	l := len(out) - 1
	for ; l >= 0; l-- {
		if out[l] != 10 && out[l] != 13 {
			break
		}
	}
	return string(out[:l+1])
}

func launchDeploy(d deployCfg) error {
	if err := os.Chdir(d.Path); err != nil {
		PrintErrMsg("Erreur lors du changement de répertoire vers " + d.Path +
			" : " + err.Error())
		return err
	}

	dest := path.Join(d.Path, d.Dist.Dest)
	if err := os.RemoveAll(dest); err != nil {
		PrintErrMsg("Erreur lors de la suppression du répertoire dist :" + err.Error())
		return err
	}
	if err := os.Mkdir(dest, os.ModeDir|os.ModePerm); err != nil {
		PrintErrMsg("Erreur lors du changement de répertoire vers " + d.Path +
			" : " + err.Error())
		return err
	}
	if err := copyFilesAndDirs(d.Dist.Source, dest); err != nil {
		return err
	}

	ebVersion := getEBVersion()
	gitVersion := getGitVersion()
	fmt.Printf("Version eb %s, version git %s\n", ebVersion, gitVersion)

	version, err := askUser("Numéro de version", true)
	if err != nil {
		return err
	}

	comment, err := askUser("Commentaire", true)
	if err != nil {
		return err
	}

	commands := []command{
		{
			AppName: "git",
			AppArgs: []string{"add", "."}},
		{
			AppName: "git",
			AppArgs: []string{"update-index", "--chmod=+x", "bin/application"}},
		{
			AppName: "git",
			AppArgs: []string{"commit", "-m", comment}},
		{
			AppName: "git",
			AppArgs: []string{"tag", "-a", version, "-m", comment}},
		{
			AppName: "eb",
			AppArgs: []string{"deploy", "-l", version, "-m", comment}},
	}

	for _, c := range commands {
		cmd := exec.Command(c.AppName, c.AppArgs...)
		out, err := cmd.Output()
		if err != nil {
			describe := c.AppName
			if len(c.AppArgs) > 0 {
				describe = describe + c.AppArgs[0]
			}
			var errMsg *exec.ExitError
			if errors.As(err, &errMsg) {
				PrintErrMsg("Erreur d'exécution de " + describe + " : " + string(errMsg.Stderr))
			} else {
				PrintErrMsg("Erreur d'exécution de " + describe + " : " + err.Error())
			}
			return err
		}
		fmt.Print(out)
	}

	return nil
}

func copyFilesAndDirs(src, dest string) error {
	var dirs []string
	files, err := ioutil.ReadDir(src)
	if err != nil {
		PrintErrMsg("Impossible de lire le contenu du répertoire" + src + " : " +
			err.Error())
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			if err = os.Mkdir(path.Join(dest, file.Name()), os.ModeDir|os.ModePerm); err != nil {
				PrintErrMsg("Impossible de créer le répertoire " + file.Name() + " : " +
					err.Error())
				return err
			}
			dirs = append(dirs, file.Name())
			continue
		}
		if file.Mode().IsRegular() {
			src, err := os.Open(path.Join(src, file.Name()))
			if err != nil {
				PrintErrMsg("Impossible d'ouvrir le fichier " + file.Name() + " : " +
					err.Error())
				return err
			}
			defer src.Close()
			dst, err := os.Create(path.Join(dest, file.Name()))
			if err != nil {
				PrintErrMsg("Impossible de créer le fichier " + file.Name() + " : " +
					err.Error())
				return err
			}
			defer dst.Close()
			_, err = io.Copy(dst, src)
			if err != nil {
				PrintErrMsg("Impossible de copier le fichier " + file.Name() + " : " +
					err.Error())
				return err
			}
		}
	}
	for _, d := range dirs {
		if err = copyFilesAndDirs(path.Join(src, d), path.Join(dest, d)); err != nil {
			return err
		}
	}
	return nil
}
