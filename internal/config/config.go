package config

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"strconv"
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

func GetMinioDsn() (string, error) {
	value := os.Getenv("MINIO_DSN")
	if value == "" {
		v := viper.GetString("depend.minio_dsn")
		return v, nil
	} else {
		return value, nil
	}
}

func GetDependMsgRpc() (string, error) {
	value := os.Getenv("MSG_RPC")
	if value == "" {
		v := viper.GetString("depend.msg_rpc")
		return v, nil
	}

	return value, nil
}

func GetFirstParty() bool { // 仅 node_manager 使用
	v := viper.GetBool("biz.first_party")
	return v
}

func GetTaskConcurrent() int { // 仅 server 使用
	value := os.Getenv("BIZ_TASK_CONCURRENT")
	if value == "" {
		return viper.GetInt("biz.task_concurrent")
	} else {
		num, err := strconv.Atoi(value)
		if err != nil {
			log.Println("strconv.Atoi error: ", err)
			return 0
		} else {
			return num
		}
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
		return hostName //+ "-" + idgen.GetRandStr()
		//return hostName + "-" + idgen.GetRandStr()
	} else {
		return podName
	}
}

func GetBizSchedServer() (string, error) {
	v := viper.GetString("biz.sched_server")
	return v, nil
}
