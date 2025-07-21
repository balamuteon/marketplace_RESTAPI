package config

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Env        string     `mapstructure:"ENV"`
	HTTPServer HTTPServer `mapstructure:"HTTP_SERVER"`
	Database   Database   `mapstructure:"DB"`
	Auth       Auth       `mapstructure:"AUTH"`
}

type HTTPServer struct {
	Port        string        `mapstructure:"PORT"`
	Timeout     time.Duration `mapstructure:"TIMEOUT"`
	IdleTimeout time.Duration `mapstructure:"IDLE_TIMEOUT"`
}

type Database struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type Auth struct {
	JWTSecret string        `mapstructure:"jwtsecret"`
	TokenTTL  time.Duration `mapstructure:"token_ttl"`
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	viper.AutomaticEnv()

	_ = viper.BindEnv("db.host", "DB_HOST")
	_ = viper.BindEnv("db.user", "DB_USER")
	_ = viper.BindEnv("db.password", "DB_PASSWORD")
	_ = viper.BindEnv("db.dbname", "DB_NAME")

	_ = viper.BindEnv("auth.jwtsecret", "AUTH_JWT_SECRET")

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	if port := os.Getenv("PORT"); port != "" {
		cfg.HTTPServer.Port = port
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation error: %s", err)
	}

	return &cfg
}

func (c *Config) Validate() error {
	if c.Auth.JWTSecret == "" {
		return errors.New("auth.jwt_secret is not set")
	}
	if c.Auth.TokenTTL <= 0 {
		return errors.New("auth.token_ttl must be a positive duration")
	}
	if c.HTTPServer.Port == "" {
		return errors.New("http_server.port is not set")
	}
	return nil
}
