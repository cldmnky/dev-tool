package config

// Config defines the config
type Config struct {
	Rancher struct {
		Clusters []RancherCluster `yaml:"clusters"`
	} `yaml:"rancher"`
	Kubernetes struct {
		Clusters []KubernetesCluster `yaml:"clusters"`
	} `yaml:"kubernetes"`
}

type RancherCluster struct {
	Environment string `yaml:"environment"`
	URL         string `yaml:"url"`
	Token       string `yaml:"token"`
}

type KubernetesCluster struct {
	ClusterName string `yaml:"clustername"`
	KubeConfig  string `yaml:"kubeconfig"`
}
