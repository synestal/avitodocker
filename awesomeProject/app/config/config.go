package config

import (
	"github.com/spf13/viper"
	"log"
	"time"
)

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	AppVersion string `json:"appVersion"`
	Host       string `json:"host" validate:"required"`
	Port       string `json:"port" validate:"required"`
	Timeout    time.Duration
}

type PostgresConfig struct {
	PgDriver string `json:"pgDriver"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"-"`
	DBName   string `json:"DBName"`
	SSLMode  string `json:"sslMode"`
}

type RedisConfig struct {
	Addr     string `json:"host"`
	Password string `json:"password"`
	DB       string `json:"db"`
}

func LoadConfig() (*viper.Viper, error) {
	viperInstance := viper.New()
	viperInstance.AddConfigPath("./config")
	viperInstance.SetConfigName("config")
	viperInstance.SetConfigType("yml")
	err := viperInstance.ReadInConfig()
	if err != nil {
		return nil, err
	}
	return viperInstance, nil
}

func ParseConfig(v *viper.Viper) (*Config, error) {
	var c Config
	err := v.Unmarshal(&c)
	if err != nil {
		log.Fatalf("unable to decode config into struct, %v", err)
		return nil, err
	}
	return &c, nil
}
