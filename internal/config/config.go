package config

import (
	"fmt"

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

	if err := viper.Unmarshal(&Cfg); err != nil {
		return fmt.Errorf("failed to decode config: %v", err)
	}

	// Validation
	if Cfg.AWS.Endpoint == "" {
		return fmt.Errorf("aws.endpoint is required")
	}
	if Cfg.AWS.Region == "" {
		return fmt.Errorf("aws.region is required")
	}
	if Cfg.AWS.AccessKeyID == "" {
		return fmt.Errorf("aws.access_key_id is required")
	}
	if Cfg.AWS.SecretAccessKey == "" {
		return fmt.Errorf("aws.secret_access_key is required")
	}
	if Cfg.AWS.Bucket == "" {
		return fmt.Errorf("aws.bucket is required")
	}
	if Cfg.LocalDir == "" {
		return fmt.Errorf("local_dir is required")
	}
	if Cfg.RemoteDir == "" {
		return fmt.Errorf("remote_dir is required")
	}
	if len(Cfg.BackupDB) > 0 {
		for _, db := range Cfg.BackupDB {
			if db.Type != "mysql" && db.Type != "mariadb" {
				return fmt.Errorf("backup_db.type is invalid")
			}
			if db.Host == "" {
				return fmt.Errorf("backup_db.host is required")
			}
			if db.Port == "" {
				return fmt.Errorf("backup_db.port is required")
			}
			if db.User == "" {
				return fmt.Errorf("backup_db.user is required")
			}
			if db.Password == "" {
				return fmt.Errorf("backup_db.password is required")
			}
			if db.DBName == "" {
				return fmt.Errorf("backup_db.dbname is required")
			}
		}
	}

	return nil
}
