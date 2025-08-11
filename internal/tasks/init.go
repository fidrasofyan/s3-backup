package tasks

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AWS       AWSConfig `yaml:"aws"`
	LocalDir  string    `yaml:"local_dir"`
	RemoteDir string    `yaml:"remote_dir"`
}

type AWSConfig struct {
	Endpoint        string `yaml:"endpoint"`
	Region          string `yaml:"region"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Bucket          string `yaml:"bucket"`
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
