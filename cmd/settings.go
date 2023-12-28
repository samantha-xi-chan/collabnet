package main

import (
	"collab-net-v2/internal/settings"
	"fmt"
)

func main() {
	// 加载配置
	//cfg, err := settings.LoadConfigJson("config.json")
	cfg, err := settings.LoadConfigYaml("app.yaml")
	if err != nil {
		panic(err)
	}

	// 在全局范围内可以随时访问配置
	fmt.Printf("App Name: %s\n", cfg.AppName)
	fmt.Printf("Log Level: %s\n", cfg.LogLevel)

	// 这里可以调用其他函数，传递配置对象作为参数
	// doSomething(cfg)
}
