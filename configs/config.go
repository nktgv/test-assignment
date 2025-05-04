package configs

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env      string         `yaml:"env"      env-default:"dev"`
	Port     int            `yaml:"port"                       env-required:"true"`
	Postgres PostgresConfig `yaml:"postgres"                   env-required:"true"`
	Hosts    []string       `yaml:"hosts"                      env-required:"true"`
}

type PostgresConfig struct {
	User     string `yaml:"user"     env-required:"true"  env:"POSTGRES_USER"`
	Password string `yaml:"password" env-required:"true"  env:"POSTGRES_PASSWORD"`
	Host     string `yaml:"host"     env-required:"true"`
	Port     int    `yaml:"port"     env-required:"true"`
	DB       string `yaml:"db"       env-required:"true"  env:"POSTGRES_DBNAME"`
	Email    string `yaml:"email"    env-required:"false" env:"PGADMIN_EMAIL"`
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
		slog.Error("Error loading .env file", err)
	}

	res = os.Getenv("CONFIG_PATH")

	slog.Info("config path fetched", res)

	return res
}

func loadByPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		slog.Error("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		slog.Error("cannot read config: %s", err)
	}

	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	db := os.Getenv("POSTGRES_DBNAME")
	email := os.Getenv("PGADMIN_EMAIL")
	cfg.Postgres.User = user
	cfg.Postgres.Password = pass
	cfg.Postgres.DB = db
	cfg.Postgres.Email = email
	fmt.Println(cfg.Postgres)

	return &cfg
}
