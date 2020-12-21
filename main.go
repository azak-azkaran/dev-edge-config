package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	LogConfig = zap.NewProductionConfig()
	Sugar     *zap.SugaredLogger = LogInit()
)


func LogInit() *zap.SugaredLogger{
	var err error
	LogConfig = zap.NewProductionConfig()
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


func main (){
	viper.SetConfigName("config.secret")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("~/")
	viper.AddConfigPath("./")
	//viper.SetConfigFile("./config.secret.yaml")

	err:= viper.ReadInConfig()
	if err != nil {
		Sugar.Panic("ERROR while reading config file", err)
	}

	projects := viper.GetStringMapString("project")
	pw := viper.GetString("password")
	for project , prefix := range projects {
		Sugar.Info("Project name:", project)
		Sugar.Info("Project prefix:", prefix)
		cmd := "openstack server list --os-password="+ pw+ " --os-cloud="+ project
		job := *exec.Command("/bin/sh", "-c",cmd)
		job.Env = append(os.Environ(), "OS_CLOUD="+project)

		out,err := job.CombinedOutput()
		if err != nil {
			Sugar.Error("Error:" , string( out ), "\n" ,err)
		}else{
			readServers(out, prefix)
		}
	}
}
