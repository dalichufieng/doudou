package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	AppName        string
	AppPort        string
	AppVersion     string
	OsWrite        string
	OsRead         string
	OsReqwrite     string
	OsReqread      string
	RequestTimeOut int64
}

func InitConf() *Config {
	workDir, _ := os.Getwd()
	log.Println("workDir：", workDir)

	viper.SetConfigType("yaml")
	viper.AddConfigPath(workDir)
	viper.SetConfigName("app")
	err := viper.ReadInConfig()
	if err != nil {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if ok {
			log.Println("找不到配置文件。。。")
			return nil
		} else {
			log.Println("配置文件出错。。。")
			return nil
		}
	}

	//打印获取到的配置文件key
	log.Println("打印获取到的配置文件key :", viper.AllKeys())

	//fmt.Println("app.name ：",viper.GetString("app.name"))
	//fmt.Println("app.port ：",viper.GetString("app.port"))
	//fmt.Println("app.version ：",viper.GetString("app.version"))
	//
	//fmt.Println("mysql.port ：",viper.GetString("mysql.port"))
	//fmt.Println("mysql.host ：",viper.GetString("mysql.host"))

	return &Config{
		AppName:        viper.GetString("app.name"),
		AppPort:        viper.GetString("app.port"),
		AppVersion:     viper.GetString("app.version"),
		OsWrite:        viper.GetString("os.write"),
		OsRead:         viper.GetString("os.read"),
		OsReqwrite:     viper.GetString("os.reqwrite"),
		OsReqread:      viper.GetString("os.reqread"),
		RequestTimeOut: viper.GetInt64("request.timeout"),
	}
}
