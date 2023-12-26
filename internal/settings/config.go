package settings

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"sync"
)

type Config struct {
	AppName  string `json:"app_name"`
	LogLevel string `json:"log_level"`
}

var (
	once     sync.Once
	instance *Config
)

func LoadConfigJson(filename string) (*Config, error) {
	once.Do(func() {
		configData, err := ioutil.ReadFile(filename)
		if err != nil {
			panic(err)
		}

		var config Config
		err = json.Unmarshal(configData, &config)
		if err != nil {
			panic(err)
		}

		instance = &config
	})

	return instance, nil
}

func LoadConfigYaml(filename string) (*Config, error) {
	once.Do(func() {
		// 初始化 Viper
		viper.SetConfigFile(filename)
		viper.AddConfigPath(".")

		if err := viper.ReadInConfig(); err != nil {
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}

		var config Config
		instance = &config
		config.AppName = viper.GetString("app_name")
		config.LogLevel = viper.GetString("log_level")
	})

	return instance, nil
}

func GetConfig() *Config {
	return instance
}
