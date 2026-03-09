package config

import (
	"errors"
	"time"

	"github.com/spf13/viper"
)

type Conf struct {
	IpLimitRps         int           `mapstructure:"IP_LIMIT_RPS"`
	IpBlockDuration    time.Duration `mapstructure:"IP_BLOCK_DURATION"`
	TokenLimitRps      int           `mapstructure:"TOKEN_LIMIT_RPS"`
	TokenBlockDuration time.Duration `mapstructure:"TOKEN_BLOCK_DURATION"`

	RedisAddr     string `mapstructure:"REDIS_ADDR"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`
	ServerPort    string `mapstructure:"SERVER_PORT"`
}

func LoadConfig(path string) (*Conf, error) {
	var cfg *Conf

	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, errors.New("erro ao fazer Unmarshal: " + err.Error())
	}

	return cfg, nil
}
