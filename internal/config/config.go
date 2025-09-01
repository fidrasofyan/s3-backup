package config

import (
	"log"
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

var Cfg Config

func MustLoadConfig(configPath string) {
	viper.SetConfigType("yaml")

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("failed to load config file: %v", err)
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		log.Fatalf("failed to decode config: %v", err)
	}

	// Validation
	if Cfg.AWS.Endpoint == "" {
		log.Fatalf("aws.endpoint is required")
	}
	if Cfg.AWS.Region == "" {
		log.Fatalf("aws.region is required")
	}
	if Cfg.AWS.AccessKeyID == "" {
		log.Fatalf("aws.access_key_id is required")
	}
	if Cfg.AWS.SecretAccessKey == "" {
		log.Fatalf("aws.secret_access_key is required")
	}
	if Cfg.AWS.Bucket == "" {
		log.Fatalf("aws.bucket is required")
	}
	if Cfg.LocalDir == "" {
		log.Fatalf("local_dir is required")
	} else {
		info, err := os.Stat(Cfg.LocalDir)
		if err != nil {
			log.Fatalf("local_dir %v is invalid", Cfg.LocalDir)
		}
		if !info.IsDir() {
			log.Fatalf("local_dir %v is not a directory", Cfg.LocalDir)
		}
	}
	if Cfg.RemoteDir == "" {
		log.Fatalf("remote_dir is required")
	} else {
		// Remove all leading slash if exists
		for strings.HasPrefix(Cfg.RemoteDir, "/") {
			Cfg.RemoteDir = strings.TrimPrefix(Cfg.RemoteDir, "/")
		}
	}
	if len(Cfg.BackupDB) > 0 {
		for _, db := range Cfg.BackupDB {
			if db.Type != "mysql" && db.Type != "mariadb" {
				log.Fatalf("backup_db.type is invalid")
			}
			if db.Host == "" {
				log.Fatalf("backup_db.host is required")
			}
			if db.Port == "" {
				log.Fatalf("backup_db.port is required")
			}
			if db.User == "" {
				log.Fatalf("backup_db.user is required")
			}
			if db.Password == "" {
				log.Fatalf("backup_db.password is required")
			}
			if db.DBName == "" {
				log.Fatalf("backup_db.dbname is required")
			}
		}
	}
}
