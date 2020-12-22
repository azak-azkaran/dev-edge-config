package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

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
	CurrentContext string           `yaml:"current-context"`
	Users          []KubectlUsers   `yaml:"users"`
	Clusters       []KubectlCluster `yaml:"clusters"`
	Contexts       []KubectlContext `yaml:"contexts"`
	apiVersion     string           `yaml:"apiVersion"`
}

const (
	ERROR_KUBE_CONFIG_NOT_FOUND = "cat: /root/.kube/config: No such file or directory"
)

func GetKubeConfig(kubemaster string) *KubectlConfig {
	cmd := "ssh "+ kubemaster+ " sudo cat /root/.kube/config"
	Sugar.Info("Calling: ", cmd)
	job := *exec.Command("bash", "-c", cmd)
	//job.Path = "/Users/gec/workspace-gec/deployment/ansible/"
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

	var conf KubectlConfig

	err = yaml.Unmarshal([]byte(o), &conf)
	if err != nil {
		Sugar.Error(err)
		return nil
	}

	return &conf
}

func SetKubemaster(servers []*Server) *Server {
	for _, s := range servers {
		if strings.Contains(s.Name, "switch0") {
			Sugar.Info("Setting Kubemaster to: ", s.IP)
			return s
		}
	}
	return nil
}

func WriteKubectlConfig(path string, content *KubectlConfig) bool {
	path, err := ReplacePath(path)
	if err != nil {
		return false
	}

	if FileExists(path) {
		Sugar.Warn("Backup old kubectl config for cluster: " , content.CurrentContext)
		err := os.Rename(path, path+".backup")
		if err != nil {
			Sugar.Error(err)
			return false
		}
	}
	Sugar.Debug("Writing KubectlConfig:\n", content)
	b, err := yaml.Marshal(content)

	_, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		Sugar.Error(err)
		return false
	}
	err = ioutil.WriteFile(path, b, 0644)
	if err != nil {
		Sugar.Error(err)
		return false
	}
	return true
}

func UpdateKubeConfig(prefix string ,k8sconfig *KubectlConfig) (string, *KubectlConfig){
	if k8sconfig == nil {
		Sugar.Error("Error while creating cluster config")
		return "", nil
	}
	currentContext := k8sconfig.CurrentContext
	Sugar.Info("Replacing")
	if strings.ContainsAny(currentContext, "dev-edge") {
		currentContext = strings.ReplaceAll(currentContext, "dev-edge", prefix)
	}
	currentContext = strings.ReplaceAll(currentContext, "kubernetes-admin@", "")
	k8sconfig.CurrentContext = currentContext
	k8sconfig.Contexts[0].Name = currentContext
	k8sconfig.Contexts[0].Context.Cluster = currentContext
	k8sconfig.Clusters[0].Name = currentContext

	var username string
	if prefix == "" {
		username = k8sconfig.Contexts[0].Context.User + "-" + currentContext
	}else{
		username = k8sconfig.Contexts[0].Context.User + "-" + prefix
	}
	k8sconfig.Contexts[0].Context.User = username
	k8sconfig.Users[0].Name = username
	return k8sconfig.CurrentContext, k8sconfig
}

func CreateKubeConfig(prefix string, servers []*Server) (string, *KubectlConfig) {
	kubemaster := SetKubemaster(servers)
	if kubemaster == nil {
		Sugar.Error("could not retrieve switch0")
		return "", nil
	}
	k8sconfig := GetKubeConfig(prefix+"." + kubemaster.Name)
	
	return UpdateKubeConfig(prefix, k8sconfig)
}

func UpdateGlobalConfig(path string, k8sconfig *KubectlConfig) *KubectlConfig{
	path, err := ReplacePath(path)
	if err != nil {
		return nil
	}

	if FileExists(path) {
		Sugar.Warn("Backup Global KubectlConfig")
		err := os.Rename(path, path+".backup")
		if err != nil {
			Sugar.Error(err)
			return nil
		}
		path = path+".backup"
	}
	var conf KubectlConfig

	Sugar.Debug("Reading path: ", path)
	content, err := ioutil.ReadFile(path) 
	if err != nil {
		Sugar.Error(err)
		return nil
	}

	err = yaml.Unmarshal([]byte(content), &conf)
	if err != nil {
		Sugar.Error(err)
		return nil
	}

	found := false
	for i, context := range conf.Contexts {
		if context.Name == k8sconfig.CurrentContext {
			found = true
			Sugar.Info("Updating Context")
			conf.Contexts[i] = k8sconfig.Contexts[0]
		}
	}
	if !found {
		Sugar.Info("Appending Context information")
		conf.Contexts = append(conf.Contexts, k8sconfig.Contexts[0])
	}

	found = false
	for i, cluster := range conf.Clusters {
		if cluster.Name == k8sconfig.Clusters[0].Name {
			found = true
			Sugar.Info("Updating Clusters")
			conf.Clusters[i] = k8sconfig.Clusters[0]
		}
	}
	if !found {
		Sugar.Info("Appending Cluster information")
		conf.Clusters = append(conf.Clusters, k8sconfig.Clusters[0])
	}

	found = false
	for i, user:= range conf.Users {
		if user.Name == k8sconfig.Users[0].Name {
			found = true
			Sugar.Info("Updating Clusters")
			conf.Users[i] = k8sconfig.Users[0]
		}
	}

	if !found {
		Sugar.Info("Appending User Information")
		conf.Users = append(conf.Users, k8sconfig.Users[0])
	}
	return &conf
}
