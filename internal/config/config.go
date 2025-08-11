package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	AWSEndpoint        string
	AWSRegion          string
	AWSAccessKeyID     string
	AWSAccessSecretKey string
	AWSBucket          string
	LocalDir           string
	RemoteDir          string
}

var Cfg Config

func LoadConfig(configPath string) {
	viper.SetConfigType("yaml")

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	Cfg = Config{
		AWSEndpoint:        viper.GetString("aws.endpoint"),
		AWSRegion:          viper.GetString("aws.region"),
		AWSAccessKeyID:     viper.GetString("aws.access_key_id"),
		AWSAccessSecretKey: viper.GetString("aws.secret_access_key"),
		AWSBucket:          viper.GetString("aws.bucket"),
		LocalDir:           viper.GetString("local_dir"),
		RemoteDir:          viper.GetString("remote_dir"),
	}

	// Validation
	if Cfg.AWSEndpoint == "" {
		log.Fatal("aws.endpoint is required")
	}
	if Cfg.AWSRegion == "" {
		log.Fatal("aws.region is required")
	}
	if Cfg.AWSAccessKeyID == "" {
		log.Fatal("aws.access_key_id is required")
	}
	if Cfg.AWSAccessSecretKey == "" {
		log.Fatal("aws.secret_access_key is required")
	}
	if Cfg.AWSBucket == "" {
		log.Fatal("aws.bucket is required")
	}
	if Cfg.LocalDir == "" {
		log.Fatal("local_dir is required")
	}
	if Cfg.RemoteDir == "" {
		log.Fatal("remote_dir is required")
	}
}
