package configs

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env      string         `yaml:"env" env-default:"dev"`
	Port     int            `yaml:"port" env-required:"true"`
	Postgres PostgresConfig `yaml:"postgres" env-required:"true"`
	Hosts    []string       `yaml:"hosts" env-required:"true"`
}

type PostgresConfig struct {
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true" env:"POSTGRES_PASSWORD"`
	Host     string `yaml:"host" env-required:"true"`
	Port     int    `yaml:"port" env-required:"true"`
	DB       string `yaml:"db" env-required:"true"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config file path is required")
	}

	return loadByPath(configPath)
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	res = os.Getenv("CONFIG_PATH")

	log.Printf("config path fetched %s", res)

	return res
}

func loadByPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
