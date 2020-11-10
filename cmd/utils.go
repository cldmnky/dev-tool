package cmd

import (
	"fmt"
	"os"

	"github.com/cldmnky/dev-tool/pkg/config"
	"github.com/spf13/viper"
)

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
