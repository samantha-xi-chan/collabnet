package config

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
)

const PATH = "config/app.yaml"

func init() {
	// 初始化 Viper
	viper.SetConfigFile(PATH)
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	log.Println("config init()")
}

func GetMySqlDsn() (string, error) {
	v := viper.GetString("depend.mysql_dsn_biz")
	return v, nil
}

func GetMqDsn() (string, error) {
	v := viper.GetString("depend.mq_dsn")
	return v, nil
}
