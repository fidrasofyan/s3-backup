package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type AWSConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	Region          string `mapstructure:"region"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Bucket          string `mapstructure:"bucket"`
}

type BackupDBConfig struct {
	Type     string `mapstructure:"type"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

type Config struct {
	AWS       AWSConfig        `mapstructure:"aws"`
	BackupDB  []BackupDBConfig `mapstructure:"backup_db"`
	LocalDir  string           `mapstructure:"local_dir"`
	RemoteDir string           `mapstructure:"remote_dir"`
}

func New(configPath string) (*Config, error) {
	viper.SetConfigType("yaml")

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	// Validation
	if cfg.AWS.Endpoint == "" {
		return nil, errors.New("aws.endpoint is required")
	}
	if cfg.AWS.Region == "" {
		return nil, errors.New("aws.region is required")
	}
	if cfg.AWS.AccessKeyID == "" {
		return nil, errors.New("aws.access_key_id is required")
	}
	if cfg.AWS.SecretAccessKey == "" {
		return nil, errors.New("aws.secret_access_key is required")
	}
	if cfg.AWS.Bucket == "" {
		return nil, errors.New("aws.bucket is required")
	}
	if cfg.LocalDir == "" {
		return nil, errors.New("local_dir is required")
	}
	info, err := os.Stat(cfg.LocalDir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(cfg.LocalDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create local_dir %v: %w", cfg.LocalDir, err)
			}
			// Update info after creation
			info, err = os.Stat(cfg.LocalDir)
			if err != nil {
				return nil, fmt.Errorf("local_dir %v is invalid after creation: %w", cfg.LocalDir, err)
			}
		} else {
			return nil, fmt.Errorf("local_dir %v is invalid: %w", cfg.LocalDir, err)
		}
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("local_dir %v is not a directory", cfg.LocalDir)
	}
	if cfg.RemoteDir == "" {
		return nil, errors.New("remote_dir is required")
	}

	for i, db := range cfg.BackupDB {
		if db.Type != "mysql" && db.Type != "mariadb" {
			return nil, fmt.Errorf("backup_db[%d].type is invalid", i)
		}
		if db.Host == "" {
			return nil, fmt.Errorf("backup_db[%d].host is required", i)
		}
		if db.Port == "" {
			return nil, fmt.Errorf("backup_db[%d].port is required", i)
		}
		if db.User == "" {
			return nil, fmt.Errorf("backup_db[%d].user is required", i)
		}
		if db.Password == "" {
			return nil, fmt.Errorf("backup_db[%d].password is required", i)
		}
		if db.DBName == "" {
			return nil, fmt.Errorf("backup_db[%d].dbname is required", i)
		}
	}

	// Normalize
	cfg.AWS.Endpoint = strings.TrimRight(cfg.AWS.Endpoint, "/")
	cfg.RemoteDir = strings.TrimLeft(cfg.RemoteDir, "/")

	return &cfg, nil
}
