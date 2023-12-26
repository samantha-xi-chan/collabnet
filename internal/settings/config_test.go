package settings

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
)

// TestLoadConfig 测试 LoadConfig 函数
func TestLoadConfig(t *testing.T) {
	// 创建一个临时的配置文件
	tempFile, err := ioutil.TempFile("", "config_test_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	// 创建一个临时的配置对象
	tempConfig := Config{
		AppName:  "TestApp",
		LogLevel: "debug",
		// 添加其他配置项...
	}

	// 将配置对象转为 JSON 并写入临时文件
	configData, err := json.Marshal(tempConfig)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(tempFile.Name(), configData, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 测试加载配置
	cfg, err := LoadConfig(tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// 验证配置项是否正确加载
	if cfg.AppName != tempConfig.AppName {
		t.Errorf("Expected AppName %s, got %s", tempConfig.AppName, cfg.AppName)
	}
	if cfg.LogLevel != tempConfig.LogLevel {
		t.Errorf("Expected LogLevel %s, got %s", tempConfig.LogLevel, cfg.LogLevel)
	}
	// 验证其他配置项...

	// 测试 GetConfig 函数是否返回正确的全局配置实例
	globalCfg := GetConfig()
	if globalCfg != cfg {
		t.Error("GetConfig does not return the correct global configuration instance")
	}
}
