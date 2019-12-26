package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"gopkg.in/gookit/color.v1"
	"gopkg.in/yaml.v2"
)

var (
	rootCmd = &cobra.Command{
		Use:   "fsd",
		Short: "Utilitaire pour le déploiement d'applications sur AWS",
	}
	deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Compile le backend et le frontend et le déploie sur AWS",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("deploy ne prend qu'un argument : le nom de l'application")
			}
			for _, app := range cfg.Application {
				if args[0] == app.Name {
					return nil
				}
			}
			return errors.New("application non trouvée dans le fichier de configuration")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, a := range cfg.Application {
				if a.Name == args[0] {
					return launch(a)
				}
			}
			return fmt.Errorf("impossible de trouver l'application %s", args[0])
		},
	}
	cfgFile               string
	cfg                   config
	noBackend, noFrontend bool
)

type config struct {
	Application []fullStackCfg `yaml:"application"`
}

type fullStackCfg struct {
	Name     string    `yaml:"name"`
	BackEnd  partCfg   `yaml:"backend"`
	FrontEnd partCfg   `yaml:"frontend"`
	Deploy   deployCfg `yaml:"deploy"`
}

type partCfg struct {
	Path        string   `yaml:"path"`
	Command     string   `yaml:"command"`
	Args        []string `yaml:"args"`
	Environment []envVar `yaml:"environment"`
}

type deployCfg struct {
	Path      string   `yaml:"path"`
	Command   string   `yaml:"command"`
	Args      []string `yaml:"args"`
	Dist      dist     `yaml:"dist"`
	AppSource string   `yaml:"app_source"`
}

type dist struct {
	Source string `yaml:"source"`
	Dest   string `yaml:"dest"`
}

type envVar struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// Execute launches the deploy command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(deployCmd)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"fichier de configuration, par défaut ~/.fsd.yaml")
	deployCmd.PersistentFlags().BoolVarP(&noBackend, "noBack", "b", false,
		"éviter de recompiler le backend")
	deployCmd.PersistentFlags().BoolVarP(&noFrontend, "noFront", "f", false,
		"éviter de recompiler le frontend")
}

func initConfig() {
	var err error
	if cfgFile == "" {
		cfgFile, err = os.UserHomeDir()
		if err != nil {
			fmt.Printf("impossible de récupérer le chemin du dossier utilisateur %v", err)
			os.Exit(1)
		}
	}
	content, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		fmt.Printf("impossible de lire le fichier de configuration %v", err)
		os.Exit(1)
	}
	if err = yaml.Unmarshal(content, &cfg); err != nil {
		fmt.Printf("erreur de décodage du fichier de configuration %v", err)
		os.Exit(1)
	}
}

func launch(c fullStackCfg) error {
	if !noBackend {
		color.Info.Println("Compilation du backend")
		if err := launchPart(c.BackEnd); err != nil {
			return err
		}
	}
	if !noFrontend {
		color.Info.Println("Compilation du frontend")
		if err := launchPart(c.FrontEnd); err != nil {
			return err
		}
	}
	color.Info.Println("Déploiement")
	// return launchDeploy(c.Deploy)
	return nil
}

func launchPart(p partCfg) error {
	if err := os.Chdir(p.Path); err != nil {
		return fmt.Errorf("changement de répertoire %s : %v", p.Path, err)
	}
	cmd := exec.Command(p.Command)
	for _, a := range p.Args {
		cmd.Args = append(cmd.Args, a)
	}
	for _, e := range p.Environment {
		os.Setenv(e.Name, e.Value)
	}
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("erreur d'exécution : %v", err)
	}
	fmt.Printf("%s", out)
	return nil
}

func launchDeploy(d deployCfg) error {
	if err := os.Chdir(d.Path); err != nil {
		return fmt.Errorf("changement de répertoire %s : %v", d.Path, err)
	}
	dest := d.Path + "\\" + d.Dist.Dest
	if err := os.RemoveAll(dest); err != nil {
		return fmt.Errorf("suppression du répertoire dist : %v", err)
	}
	if err := os.Mkdir(dest, os.ModeDir|os.ModePerm); err != nil {
		return fmt.Errorf("changement de répertoire %s : %v", d.Path, err)
	}
	if err := copyFilesAndDirs(d.Dist.Source, dest); err != nil {
		return err
	}
	cmd := exec.Command(d.Command)
	for _, a := range d.Args {
		cmd.Args = append(cmd.Args, a)
	}
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("erreur d'exécution : %v", err)
	}
	fmt.Printf("%s", out)
	return nil
}

func copyFilesAndDirs(src, dest string) error {
	var dirs []string
	files, err := ioutil.ReadDir(src)
	if err != nil {
		return fmt.Errorf("impossible de lire le contenu du répertoire %s : %v", src, err)
	}
	for _, file := range files {
		if file.IsDir() {
			if err = os.Mkdir(dest+"\\"+file.Name(), os.ModeDir|os.ModePerm); err != nil {
				return fmt.Errorf("impossible de créer le répertoire %s : %v", file.Name(), err)
			}
			dirs = append(dirs, file.Name())
			continue
		}
		if file.Mode().IsRegular() {
			src, err := os.Open(src + "\\" + file.Name())
			if err != nil {
				return fmt.Errorf("impossible d'ouvrir le fichier %s : %v", file.Name(), err)
			}
			defer src.Close()
			dst, err := os.Create(dest + "\\" + file.Name())
			if err != nil {
				return fmt.Errorf("impossible d'ouvrir le fichier %s : %v", file.Name(), err)
			}
			defer dst.Close()
			_, err = io.Copy(dst, src)
			if err != nil {
				return fmt.Errorf("impossible de copier le fichier %s : %v", file.Name(), err)
			}
		}
	}
	for _, d := range dirs {
		if err = copyFilesAndDirs(src+"\\"+d, dest+"\\"+d); err != nil {
			return err
		}
	}
	return nil
}
