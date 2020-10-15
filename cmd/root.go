/*
Copyright © 2020 Betsson tooling

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cldmnky/dev-tool/pkg/config"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var cfgFile string

const (
	configDir      = ".dev-tool"
	configFileName = "dev-tool.yaml"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dev-tool",
	Short: "Doing fun things at Betsson",
	Long:  `THe dev-tool can do lot's of things, try dev-tool help for starters.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dev-tool/dev-tool.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home := getHomeDir()
		// Search config in home directory with name ".dev-tool" (without extension).
		viper.AddConfigPath(fmt.Sprintf("%s/%s", home, configDir))
		viper.AddConfigPath(".")
		viper.SetConfigName(configFileName)
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in, else create it
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		if err := createDefaultConfig(configFileName); err != nil {
			fmt.Printf("Error creating config file: %s", err)
			os.Exit(1)
		}
	}
}

func getHomeDir() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println("Unable to locate home directory")
		os.Exit(1)
	}
	return home
}

func createDefaultConfig(configFile string) error {
	home := getHomeDir()
	defaultConfig := &config.Config{}
	defaultConfig.Rancher.Clusters = []config.RancherCluster{
		{
			Environment: "prod",
			URL:         "https://kube-api.prod",
			Token:       "changeme",
		},
		{
			Environment: "test",
			URL:         "https://kube-api.test",
			Token:       "changeme",
		},
	}
	newConfig, _ := yaml.Marshal(&defaultConfig)
	configPath := fmt.Sprintf("%s/%s/%s", home, configDir, configFile)
	os.Mkdir(fmt.Sprintf("%s/%s", home, configDir), 0700)
	ioutil.WriteFile(configPath, newConfig, 0600)
	return nil
}
