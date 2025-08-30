package tasks

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type AWSConfig struct {
	Endpoint        string `yaml:"endpoint"`
	Region          string `yaml:"region"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Bucket          string `yaml:"bucket"`
}

type BackupDBConfig struct {
	Type     string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type Config struct {
	AWS       AWSConfig        `yaml:"aws"`
	BackupDB  []BackupDBConfig `yaml:"backup_db"`
	LocalDir  string           `yaml:"local_dir"`
	RemoteDir string           `yaml:"remote_dir"`
}

func InitializeConfig() error {
	if _, err := os.Stat("config.yaml"); err == nil {
		return fmt.Errorf("config.yaml already exists")
	}

	fmt.Println("Initializing config.yaml...")
	cfg := Config{
		AWS: AWSConfig{
			Endpoint:        "https://abc123.r2.cloudflarestorage.com",
			Region:          "auto",
			AccessKeyID:     "abc123",
			SecretAccessKey: "abc123",
			Bucket:          "my-bucket",
		},
		BackupDB: []BackupDBConfig{
			{
				Type:     "mysql",
				Host:     "localhost",
				Port:     "3306",
				User:     "root",
				Password: "password",
				DBName:   "mydb",
			},
		},
		LocalDir:  "/home/user/backup",
		RemoteDir: "backup",
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to generate config: %v", err)
	}

	err = os.WriteFile("config.yaml", data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}

	fmt.Println("Config initialized successfully")

	return nil
}
