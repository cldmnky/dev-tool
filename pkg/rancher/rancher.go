package rancher

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/rancher/cli/cliclient"
	"github.com/rancher/cli/config"
	"github.com/rancher/norman/types"
	clusterClient "github.com/rancher/types/client/cluster/v3"
	client "github.com/rancher/types/client/management/v3"
)

const (
	authTokenURL = "%s/v3-public/authTokens/%s"
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

// GetToken returns a Rancher management token using a ADFS login flow
func GetToken(rancherURI string) (client.Token, error) {
	token := client.Token{}
	// Generate a private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return token, err
	}
	publicKey := privateKey.PublicKey
	marshalKey, err := json.Marshal(publicKey)
	if err != nil {
		return token, err
	}
	encodedKey := base64.StdEncoding.EncodeToString(marshalKey)
	id, err := generateKey()
	if err != nil {
		return token, err
	}
	responseType := "json"
	tokenURL := fmt.Sprintf(authTokenURL, rancherURI, id)
	req, err := http.NewRequest(http.MethodGet, tokenURL, bytes.NewBuffer(nil))
	if err != nil {
		return token, err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("accept", "application/json")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{Transport: tr, Timeout: 300 * time.Second}

	loginRequest := fmt.Sprintf("%s/login?requestId=%s&publicKey=%s&responseType=%s",
		rancherURI, id, encodedKey, responseType)

	fmt.Printf("\nLogin to Rancher Server at %s \n", loginRequest)
	openbrowser(loginRequest)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// timeout for user to login and get token
	timeout := time.NewTicker(15 * time.Minute)
	defer timeout.Stop()

	poll := time.NewTicker(10 * time.Second)
	defer poll.Stop()

	for {
		select {
		case <-poll.C:
			res, err := client.Do(req)
			if err != nil {
				return token, err
			}
			content, err := ioutil.ReadAll(res.Body)
			if err != nil {
				res.Body.Close()
				return token, err
			}
			res.Body.Close()
			err = json.Unmarshal(content, &token)
			if err != nil {
				return token, err
			}
			if token.Token == "" {
				continue
			}
			decoded, err := base64.StdEncoding.DecodeString(token.Token)
			if err != nil {
				return token, err
			}
			decryptedBytes, err := privateKey.Decrypt(nil, decoded, &rsa.OAEPOptions{Hash: crypto.SHA256})
			if err != nil {
				panic(err)
			}
			token.Token = string(decryptedBytes)
			// delete token
			req, err = http.NewRequest(http.MethodDelete, tokenURL, bytes.NewBuffer(nil))
			if err != nil {
				return token, err
			}
			req.Header.Set("content-type", "application/json")
			req.Header.Set("accept", "application/json")
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
			client = &http.Client{Transport: tr, Timeout: 150 * time.Second}
			res, err = client.Do(req)
			if err != nil {
				// log error and use the token if login succeeds
				fmt.Printf("DeleteToken: %v", err)
			}
			return token, nil

		case <-timeout.C:
			break

		case <-interrupt:
			fmt.Printf("received interrupt\n")
			break
		}

		return token, nil
	}
}

// CreateToken generates a token for the cli
func (c *RancherClient) CreateToken() (*client.Token, error) {
	tokenOpts := &client.Token{Description: "dev-tool"}
	token, err := c.Client.ManagementClient.Token.Create(tokenOpts)
	if err != nil {
		return nil, err
	}
	return token, err
}

// GetCluster returns Servers for the user
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

func generateKey() (string, error) {
	characters := "abcdfghjklmnpqrstvwxz12456789"
	tokenLength := 32
	token := make([]byte, tokenLength)
	for i := range token {
		r, err := rand.Int(rand.Reader, big.NewInt(int64(len(characters))))
		if err != nil {
			return "", err
		}
		token[i] = characters[r.Int64()]
	}
	return string(token), nil
}

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}
