package config

import (
	"collab-net-v2/package/util/idgen"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
)

const PATH = "config/app.yaml"

func init() {
	podName := os.Getenv("POD_NAME")
	if podName == "" {
		log.Println("Failed to get POD_NAME environment variable")

		// 初始化 Viper
		viper.SetConfigFile(PATH)
		viper.AddConfigPath(".")

		if err := viper.ReadInConfig(); err != nil {
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}
	} else {
		log.Printf("Pod Name: %s\n", podName)

	}

	log.Println("config init()")
}

func GetMySqlDsn() (string, error) {
	value := os.Getenv("MYSQL_DSN")
	if value == "" {
		v := viper.GetString("depend.mysql_dsn")
		return v, nil
	} else {
		return value, nil
	}
}

func GetMqDsn() (string, error) {
	value := os.Getenv("MQ_DSN")
	if value == "" {
		v := viper.GetString("depend.mq_dsn")
		return v, nil
	} else {
		return value, nil
	}
}

func GetLogServer() string {
	logServer := os.Getenv("LOG_SERVER")
	if logServer == "" {
		v := viper.GetString("debug.log_server")
		return v
	} else {
		log.Printf("logServer: %s\n", logServer)
		return logServer
	}
}

func GetRunningInstance() string {
	podName := os.Getenv("POD_NAME")
	if podName == "" {
		hostName, _ := os.Hostname()
		return hostName + "-" + idgen.GetRandStr()
	} else {
		return podName
	}
}

func GetBizSchedServer() (string, error) {
	v := viper.GetString("biz.sched_server")
	return v, nil
}
