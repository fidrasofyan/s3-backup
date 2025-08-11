package config

import (
	"fmt"

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

func LoadConfig(configPath string) error {
	viper.SetConfigType("yaml")

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to load config file: %v", err)
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
		return fmt.Errorf("aws.endpoint is required")
	}
	if Cfg.AWSRegion == "" {
		return fmt.Errorf("aws.region is required")
	}
	if Cfg.AWSAccessKeyID == "" {
		return fmt.Errorf("aws.access_key_id is required")
	}
	if Cfg.AWSAccessSecretKey == "" {
		return fmt.Errorf("aws.secret_access_key is required")
	}
	if Cfg.AWSBucket == "" {
		return fmt.Errorf("aws.bucket is required")
	}
	if Cfg.LocalDir == "" {
		return fmt.Errorf("local_dir is required")
	}
	if Cfg.RemoteDir == "" {
		return fmt.Errorf("remote_dir is required")
	}

	return nil
}
