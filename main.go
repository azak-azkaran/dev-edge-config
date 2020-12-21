package main

import (
	"fmt"

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

	projects := viper.GetStringMap("project")
	pw := viper.GetString("password")
	Sugar.Info(pw)
	Sugar.Info(projects)
}
