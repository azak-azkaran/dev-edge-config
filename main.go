package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)
const (
	MODE_OPENSTACK = "openstack"
	MODE_KUBERNETES = "k8s"
)

var (
	LogConfig                    = zap.NewProductionConfig()
	Sugar     *zap.SugaredLogger = LogInit()
)

func LogInit() *zap.SugaredLogger {
	var err error
	LogConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	LogConfig.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	LogConfig.Encoding = "console"
	logger, err := LogConfig.Build()
	if err != nil {
		fmt.Println("Error building logger:", err)
	}
	defer logger.Sync() // flushes buffer, if any
	return logger.Sugar()
}

func FetchOpenstackInfo()  error{
viper.SetConfigName("config.secret")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("~/")
	viper.AddConfigPath("./")

	err := viper.ReadInConfig()
	if err != nil {
		Sugar.Panic("ERROR while reading config file", err)
		return err
	}

	projects := viper.GetStringMapString("project")
	pw := viper.GetString("password")
	for project, prefix := range projects {
		Sugar.Info("Project name:", project)
		Sugar.Info("Project prefix:", prefix)
		cmd := "openstack server list --os-password=" + pw + " --os-cloud=" + project
		job := *exec.Command("/bin/sh", "-c", cmd)
		job.Env = append(os.Environ(), "OS_CLOUD="+project)

		out, err := job.CombinedOutput()
		if err != nil {
			Sugar.Error("Error:", string(out), "\n", err)
		} else {
			CreateConfigs("~/.ssh/config.d/"+prefix+".gecgo.net", string(out), prefix)
		}
	}
	return nil
}

func FetchK8sInfo(ssh string){
	if ssh == "" {
		Sugar.Error("Error: url was not defined",)
		return
	}

	Sugar.Info("Fetchin kubectl config from server: ", ssh)
	clustername, k8sconfig := UpdateKubeConfig("", GetKubeConfig(ssh))
	
	WriteKubectlConfig("~/.kube/config.d/"+clustername, k8sconfig)

	conf := UpdateGlobalConfig("~/.kube/config", k8sconfig)
	WriteKubectlConfig("~/.kube/config", conf)
}

func main() {
	mode := flag.String("mode",MODE_OPENSTACK,"defines the mode in which to run ( default: openstack) ")
	sshurl := flag.String("url", "", "defines the url from which to get the k8s config, only used in k8s mode")
	flag.Parse()
	Sugar.Info("Mode: ", *mode)
	if strings.EqualFold(*mode, MODE_OPENSTACK){
		FetchOpenstackInfo()
	}else if strings.EqualFold(*mode, MODE_KUBERNETES) {
		FetchK8sInfo(*sshurl)
	} else{
		Sugar.Panic("Unknown mode, Dying Horribly")
		return
	}
	Sugar.Info("Happy Death")
}

func CreateConfigs(path string, table string, prefix string) {
	buffer, servers := readServers(table, prefix)
	if buffer == nil {
		Sugar.Error("Error: Could not read Openstack Config")
		return
	}
	writeSSHConfig(path, buffer)

	clustername, k8sconfig := CreateKubeConfig(prefix, servers)
	if k8sconfig == nil{
		return
	}
	WriteKubectlConfig("~/.kube/config.d/"+clustername, k8sconfig)

	conf := UpdateGlobalConfig("~/.kube/config", k8sconfig)
	WriteKubectlConfig("~/.kube/config", conf)
}

func ReplacePath(path string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		Sugar.Fatal(err)
		return "", err
	}
	path = strings.ReplaceAll(path, "~", usr.HomeDir)
	return path, nil
}
