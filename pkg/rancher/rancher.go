package rancher

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/rancher/cli/cliclient"
	"github.com/rancher/cli/config"
	"github.com/rancher/norman/types"
	clusterClient "github.com/rancher/types/client/cluster/v3"
	client "github.com/rancher/types/client/management/v3"
)

type RancherClient struct {
	Client   *cliclient.MasterClient
	Settings *config.ServerConfig
}

// GetClient imports a client
func GetClient(rancherURI string, token string) (*RancherClient, error) {
	serverConfig := &config.ServerConfig{}
	uri, err := url.ParseRequestURI(rancherURI)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse SERVERURL (%s), make sure it is a valid HTTPS URL (e.g. https://rancher.yourdomain.com or https://1.1.1.1). Error: %s", rancherURI, err)
	}
	uri.Path = ""
	serverConfig.URL = uri.String()
	auth := strings.Split(token, ":")
	serverConfig.AccessKey = auth[0]
	serverConfig.SecretKey = auth[1]
	serverConfig.TokenKey = token

	client, err := cliclient.NewManagementClient(serverConfig)
	if err != nil {
		return nil, fmt.Errorf("Unable to get a client for the rancher server: %s", err)
	}

	rancherClient := &RancherClient{}
	rancherClient.Client = client
	rancherClient.Settings = serverConfig
	return rancherClient, nil
}

// GetServers returns Servers for the user
func (c *RancherClient) GetCluster(clusterName string) (*client.Cluster, error) {
	clusters, err := c.Client.ManagementClient.Cluster.List(clusterListOpts())
	if err != nil {
		return nil, fmt.Errorf("Error listing clusters: %s", err)
	}
	for _, cluster := range clusters.Data {
		if cluster.Name == clusterName {
			return &cluster, nil
		}
	}

	return nil, &rancherError{fmt.Sprintf("Could not find cluster: %s", clusterName), notFoundErr}
}

func (c *RancherClient) GetProject(projectName string, cluster *client.Cluster) (*client.Project, error) {
	filter := &types.ListOpts{}
	//filter.Filters["clusterId"] = cluster.ID
	projects, err := c.Client.ManagementClient.Project.List(filter)
	if err != nil {
		return nil, fmt.Errorf("Error listing projects: %s", err)
	}
	for _, project := range projects.Data {
		if project.Name == projectName {
			return &project, nil
		}
	}
	return nil, &rancherError{fmt.Sprintf("Could not find cluster: %s", projectName), notFoundErr}
}

func (c *RancherClient) CreateProject(projectName string, description string, cluster *client.Cluster) (*client.Project, error) {
	newProject := &client.Project{
		Name:        projectName,
		ClusterID:   cluster.ID,
		Description: description,
	}
	project, err := c.Client.ManagementClient.Project.Create(newProject)
	if err != nil {
		return nil, fmt.Errorf("Could not create project: %s", err)
	}
	return project, nil
}

func (c *RancherClient) GetNamespace(namespaceName string, project *client.Project) (*clusterClient.Namespace, error) {
	filter := &types.ListOpts{}
	c.Settings.Project = project.ID
	cclient, err := cliclient.NewClusterClient(c.Settings)
	namespaces, err := cclient.ClusterClient.Namespace.List(filter)
	if err != nil {
		return nil, fmt.Errorf("Error listing namespaces: %s", err)
	}
	for _, namespace := range namespaces.Data {
		if namespace.ProjectID == project.ID && namespace.Name == namespaceName {
			return &namespace, nil
		}
	}
	return nil, &rancherError{fmt.Sprintf("Could not find namespace: %s", namespaceName), notFoundErr}
}

func (c *RancherClient) CreateNamespace(namespaceName string, project *client.Project) (*clusterClient.Namespace, error) {
	c.Settings.Project = project.ID
	cclient, err := cliclient.NewClusterClient(c.Settings)
	newNamespace := &clusterClient.Namespace{
		Name:      namespaceName,
		ProjectID: project.ID,
	}
	namespace, err := cclient.ClusterClient.Namespace.Create(newNamespace)
	if err != nil {
		return nil, fmt.Errorf("Could not create namespace: %s", namespaceName)
	}
	return namespace, nil
}

func (c *RancherClient) GetKubeConfig(cluster *client.Cluster) (string, error) {
	kubeConfig, err := c.Client.ManagementClient.Cluster.ActionGenerateKubeconfig(cluster)
	if err != nil {
		return "", fmt.Errorf("Could not generate kubecnfig for: %s", cluster.Name)
	}
	return kubeConfig.Config, nil
}

func (c *RancherClient) EnsureNamespace(namespaceName string, clusterName string, projectName string) (*clusterClient.Namespace, error) {
	cluster, err := c.GetCluster(clusterName)
	if err != nil {
		return nil, err
	}
	project, err := c.GetProject(projectName, cluster)
	if err != nil {
		return nil, err
	}
	namespace, err := c.GetNamespace(namespaceName, project)
	if err != nil {
		if err, ok := err.(*rancherError); ok {
			if err.notFound() {
				// namespace not found, create it
				newNamespace, err := c.CreateNamespace(namespaceName, project)
				if err != nil {
					return nil, fmt.Errorf("Could not ensure namespace", err)
				}
				return newNamespace, nil
			}
		}
		return nil, err
	}
	return namespace, nil
}

func (c *RancherClient) EnsureProject(clusterName string, projectName string) (*client.Project, error) {
	cluster, err := c.GetCluster(clusterName)
	if err != nil {
		// Error getting cluster, bail out
		return nil, err
	}
	project, err := c.GetProject(projectName, cluster)
	if err != nil {
		if err, ok := err.(*rancherError); ok {
			if err.notFound() {
				// not found, create the project
				newProject, err := c.CreateProject(projectName, projectName, cluster)
				if err != nil {
					return nil, fmt.Errorf("Could not ensure project: %s", err)
				}
				return newProject, nil
			}
		}
		return nil, err
	}
	return project, nil
}

func clusterListOpts() *types.ListOpts {
	return &types.ListOpts{
		Filters: map[string]interface{}{
			"limit":        -1,
			"all":          true,
			"removed_null": "1",
			"system":       false,
			"state_ne": []string{
				"inactive",
				"stopped",
				"removing",
			},
		},
	}
}
