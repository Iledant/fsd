package cmd

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	rootCmd = &cobra.Command{
		Use:   "fsd",
		Short: "Utilitaire pour le déploiement d'applications sur AWS",
	}
	cfgFile string
	cfg     config
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
	Path           string   `yaml:"path"`
	VersionCommand string   `yaml:"version_command"`
	VersionArgs    []string `yaml:"version_args"`
	Command        string   `yaml:"command"`
	Args           []string `yaml:"args"`
	Environment    []envVar `yaml:"environment"`
	Message        string
	No             bool
}

type envVar struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type deployCfg struct {
	Path      string `yaml:"path"`
	Dist      dist   `yaml:"dist"`
	AppSource string `yaml:"app_source"`
}

type dist struct {
	Source string `yaml:"source"`
	Dest   string `yaml:"dest"`
}

// Execute launches the commands
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(listCommand)
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
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			PrintErrMsg("Impossible de récupérer le chemin du dossier utilisateur : " +
				err.Error())
			os.Exit(1)
		}
		cfgFile = path.Join(userHomeDir, ".fsd.yaml")
	}
	content, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		PrintErrMsg("Impossible de lire le fichier de configuration : " + err.Error())
		os.Exit(1)
	}
	if err = yaml.Unmarshal(content, &cfg); err != nil {
		PrintErrMsg("Erreur de décodage du fichier de configuration " + err.Error())
		os.Exit(1)
	}
}
