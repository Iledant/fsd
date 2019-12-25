package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	deployCmd = &cobra.Command{
		Use:   "fsd",
		Short: "fsd compile le back- et le frontend et déploie aur AWS",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Appel de la commande")
		},
	}
	cfgFile string
	cfg     config
)

type config struct {
	Application []fullStackCfg `yaml:"application"`
}

type fullStackCfg struct {
	Name     string  `yaml:"name"`
	BackEnd  partCfg `yaml:"backend"`
	FrontEnd partCfg `yaml:"frontend"`
	Deploy   partCfg `yaml:"deploy"`
}

type partCfg struct {
	Path    string `yaml:"path"`
	Command string `yaml:"command"`
}

// Execute launches the deploy command
func Execute() error {
	return deployCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	deployCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"fichier de configuration, par défaut $HOME/.fsd.yaml")
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
	if err = yaml.Unmarshal(content,&cfg) ; err !=nil {
		fmt.Printf("erreur de décodage du fichier de configuration %v", err)
		os.Exit(1)
	}
}
