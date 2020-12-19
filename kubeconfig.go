package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

var kubemaster Server

type UserData struct {
	ClientCertificate string `yaml:"client-certificate-data"`
	ClientKey         string `yaml:"client-key-data"`
}

type KubectlUsers struct {
	Name string   `yaml:"name"`
	User UserData `yaml:"user"`
}

type ClusterData struct {
	CertificateAuthority string `yaml:"certificate-authority-data"`
	Server               string `yaml:"server"`
}

type KubectlCluster struct {
	Cluster ClusterData `yaml:"cluster"`
	Name    string      `yaml:"name"`
}

type ContextData struct {
	Cluster string `yaml:"cluster"`
	User    string `yaml:"user"`
}

type KubectlContext struct {
	Name    string      `yaml:"name"`
	Context ContextData `yaml:"context"`
}

type KubectlConfig struct {
	Kind           string           `yaml:"kind"`
	Preference     []string         `yaml:"preference,omitempty"`
	CurrentContext string   `yaml:"current-context"`
	Users          []KubectlUsers   `yaml:"users"`
	Clusters       []KubectlCluster `yaml:"clusters"`
	Contexts       []KubectlContext `yaml:"contexts"`
	apiVersion     string           `yaml:"apiVersion"`
}

const (
	ERROR_KUBE_CONFIG_NOT_FOUND = "cat: /root/.kube/config: No such file or directory"
)

func GetKubeConfig() *KubectlConfig {
	cmd := "ssh dev-edge." + kubemaster.Name + " sudo cat /root/.kube/config"
	job := *exec.Command("bash", "-c", cmd)
	//job.Path = "/Users/gec/workspace-gec/deployment/ansible/"
	job.Env = append(os.Environ(), "optimist")
	output, err := job.CombinedOutput()

	if err != nil {
		if strings.Contains(string(output), ERROR_KUBE_CONFIG_NOT_FOUND) {
			Sugar.Error("Kube Config not found\n", err)
			return nil
		}
		Sugar.Error("Error running: ", cmd, "\n", err, ":\n", string(output))
		return nil
	}

	var o string
	foundApi := false
	split := strings.Split(string(output), "\n")
	for _, line := range split {
		if !foundApi {
			foundApi = strings.Contains(line, "apiVersion:")
		}
		if foundApi {
			o = o + line + "\n"
		}
	}

	err = ioutil.WriteFile("./tmp.yml", []byte(o), 0644)
	if err != nil {
		Sugar.Error(err)
		return nil
	}

	var conf KubectlConfig

	err = yaml.Unmarshal([]byte(o), &conf)
	if err != nil {
		Sugar.Error(err)
		return nil
	}

	return &conf
}

func SetKubemaster(servers []*Server) {
	for _, s := range servers {
		if strings.Contains(s.Name, "switch0") {
			Sugar.Info("Setting Kubemaster to: ", s.IP)
			kubemaster = *s
		}
	}
}
