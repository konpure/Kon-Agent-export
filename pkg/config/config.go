package config

import (
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Storage StorageConfig `yaml:"storage"`
	Log     LogConfig     `yaml:"log"`
}

type ServerConfig struct {
	QUICPort     int           `yaml:"quic_port"`
	HTTPPort     int           `yaml:"http_port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type       string        `yaml:"type"`
	MaxSize    int           `yaml:"max_size"`
	ExpireTime time.Duration `yaml:"expire_time"`
	FilePath   string        `yaml:"file_path"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

// LoadConfig 从文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Failed to read config file: %v", err)
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Printf("Failed to unmarshal config: %v", err)
		return nil, err
	}

	// 设置默认值
	setDefaults(&config)

	return &config, nil
}

// 设置默认配置值
func setDefaults(config *Config) {
	if config.Server.QUICPort == 0 {
		config.Server.QUICPort = 7843
	}
	if config.Server.HTTPPort == 0 {
		config.Server.HTTPPort = 8080
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 10 * time.Second
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 10 * time.Second
	}

	if config.Storage.Type == "" {
		config.Storage.Type = "memory"
	}
	if config.Storage.MaxSize == 0 {
		config.Storage.MaxSize = 10000
	}
	if config.Storage.ExpireTime == 0 {
		config.Storage.ExpireTime = 24 * time.Hour
	}
	if config.Storage.FilePath == "" {
		config.Storage.FilePath = "./data/"
	}

	if config.Log.Level == "" {
		config.Log.Level = "info"
	}
}
