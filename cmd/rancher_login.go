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
	"log"
	"os"

	"github.com/cldmnky/dev-tool/pkg/config"
	"github.com/cldmnky/dev-tool/pkg/rancher"
	"github.com/spf13/cobra"
)

type loginOptions struct {
	config *config.RancherCluster
}

// setupCmd represents the setup command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "rancher login",
	Long:  `Login to Rancher and get a login token`,
	Run: func(cmd *cobra.Command, args []string) {
		env, _ := cmd.Parent().Flags().GetString("environment")
		o := &loginOptions{}
		r := getRancher(env)
		o.config = r
		token, err := rancher.GetToken(o.config.URL)
		if err != nil {
			fmt.Printf("could not get token: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Got token: %+v\n", token.Token)
		c, err := rancher.GetClient(o.config.URL, token.Token)
		if err != nil {
			os.Exit(1)
		}
		cluster, err := c.GetCluster("tooling-igaming-test-euc1")
		if err != nil {
			log.Fatalf("Error getting cluster: %v", err)
		}
		fmt.Printf("Got cluster: %s\n", cluster.Name)
		//rancher.GetClient(o.config.URL)
	},
}

func init() {
	rancherCmd.AddCommand(loginCmd)
}
