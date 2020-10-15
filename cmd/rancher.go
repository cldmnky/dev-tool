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

	"github.com/spf13/cobra"
)

// rancherCmd represents the rancher command
var rancherCmd = &cobra.Command{
	Use:   "rancher",
	Short: "Manage rancher",
	Long:  `.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("rancher called")
	},
}

func init() {
	rootCmd.AddCommand(rancherCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// rancherCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// rancherCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rancherCmd.PersistentFlags().StringP("project", "p", "", "Rancher project")
	rancherCmd.PersistentFlags().StringP("cluster", "c", "", "Kubernetes cluster")
	rancherCmd.PersistentFlags().StringP("namespace", "n", "", "Kubernetes namespace")
	rancherCmd.PersistentFlags().StringP("environment", "e", "", "Environment")
	rancherCmd.MarkPersistentFlagRequired("project")
	rancherCmd.MarkPersistentFlagRequired("cluster")
	rancherCmd.MarkPersistentFlagRequired("namespace")
	rancherCmd.MarkPersistentFlagRequired("environment")
}
