package helm

import (
	"fmt"
	"log"

	"github.com/mitchellh/go-homedir"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/client-go/tools/clientcmd"
)

// Runner is a helm runner type
type Runner struct {
	config *action.Configuration
}

// NewHelmClient returns a helm client
func NewHelmClient(bamespace string) (*Runner, error) {
	// get a restConfig
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	config, err := clientcmd.BuildConfigFromFlags("", fmt.Sprintf("%s/.kube/lab-techops-test-euc1-977445621284", home))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Client: %s", config.Host)
	return nil, nil
}

/*
usr, err := user.Current()
if err != nil {
	panic(err)
}

kubeconfig := usr.HomeDir + ".kube/config"
config, _ := clientcmd.BuildConfigFromFlags("", kubeconfig)

runner, err := NewRunner(config, "tfd-namespace", "tfd-namespace")
if err != nil {
	panic(err)
}

folderPath, _ := osext.ExecutableFolder()
chart, err := loader.Load(folderPath + "/helm-charts.tgz")
if err != nil {
	panic(err)
}

runner.Install(chart, nil)

fmt.Println("Successfully deployed")

}

// Runner represents a Helm action runner capable of performing Helm
// operations for a v2beta1.HelmRelease.
type Runner struct {
config *action.Configuration
}

// NewRunner constructs a new Runner configured to run Helm actions with the
// given rest.Config, and the release and storage namespace configured to the
// provided values.

func NewRunner(clusterCfg *rest.Config, releaseNamespace, storageNamespace string) (*Runner, error) {
cfg := new(action.Configuration)
if err := cfg.Init(&genericclioptions.ConfigFlags{
	APIServer:   &clusterCfg.Host,
	CAFile:      &clusterCfg.CAFile,
	BearerToken: &clusterCfg.BearerToken,
	Namespace:   &releaseNamespace,
}, storageNamespace, "secret", debugLogger(logger)); err != nil {
	return nil, err
}
return &Runner{config: cfg}, nil
}

func (r *Runner) Install(chart *chart.Chart, values chartutil.Values) (*release.Release, error) {
d, _ := time.ParseDuration("10m")
install := action.NewInstall(r.config)
install.ReleaseName = "tfd-patient-zero"
install.Namespace = "tfd-namespace"
install.Timeout = d
install.Wait = true
install.DisableHooks = false
install.DisableOpenAPIValidation = false
install.Replace = true
install.SkipCRDs = true

return install.Run(chart, values.AsMap())
}
*/
