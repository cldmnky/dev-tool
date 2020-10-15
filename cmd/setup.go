/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"github.com/cldmnky/dev-tool/pkg/rancher"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type setupOptions struct {
	config *config.RancherCluster
}

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "setup rancher",
	Long:  `Check if rancher is setup, create project and namespace if it does not exist`,
	Run: func(cmd *cobra.Command, args []string) {
		env, _ := cmd.Parent().Flags().GetString("environment")
		cluster, _ := cmd.Parent().Flags().GetString("cluster")
		project, _ := cmd.Parent().Flags().GetString("project")
		namespace, _ := cmd.Parent().Flags().GetString("namespace")
		o := &setupOptions{}
		r := getRancher(env)
		o.config = r
		rancher, err := rancher.GetClient(o.config.URL, o.config.Token)
		if err != nil {
			fmt.Printf("%s", err)
		}
		p, err := rancher.EnsureProject(cluster, project)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		fmt.Printf("Rancher project setup: %s\n", p.ID)
		n, err := rancher.EnsureNamespace(namespace, cluster, project)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		fmt.Printf("Project namespace setup: %s\n", n.ID)
		fmt.Printf("Generating kubeconfig for %s", cluster)
		currentCluster, err := rancher.GetCluster(cluster)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		kubeConfig, err := rancher.GetKubeConfig(currentCluster)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		// write kubeconfig
		home := getHomeDir()
		if err = os.Mkdir(fmt.Sprintf("%s/%s/kubeconfig", home, configDir), 0700); err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		ioutil.WriteFile(fmt.Sprintf("%s/%s/kubeconfig/%s", home, configDir, cluster), []byte(kubeConfig), 0600)

	},
}

func init() {
	rancherCmd.AddCommand(setupCmd)
}

func getRancher(environment string) *config.RancherCluster {
	config := &config.Config{}
	viper.Unmarshal(config)

	found := false
	for _, cluster := range config.Rancher.Clusters {
		if cluster.Environment == environment {
			found = true
			return &cluster
		}
	}
	if !found {
		fmt.Printf("Could not find a rancher cluster to use: %s\n", environment)
		os.Exit(1)
	}
	return nil
}
