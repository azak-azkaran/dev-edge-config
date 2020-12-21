package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

func main() {
	viper.SetConfigName("config.secret")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("~/")
	viper.AddConfigPath("./")
	//viper.SetConfigFile("./config.secret.yaml")

	err := viper.ReadInConfig()
	if err != nil {
		Sugar.Panic("ERROR while reading config file", err)
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
}

func CreateConfigs(path string, table string, prefix string) {
	buffer, servers := readServers(table, prefix)
	if buffer == nil {
		Sugar.Error("Error: Could not read Openstack Config")
		return
	}
	writeSSHConfig(path, buffer)

	clustername, k8sconfig := CreateKubeConfig(prefix, servers)
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
