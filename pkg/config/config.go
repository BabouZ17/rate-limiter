package config

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

type RedisConfig struct {
	Address  string `json:"address"`
	Database int    `json:"database"`
}

type RateLimiterConfig struct {
	Capacity          int32  `json:"capacity"`
	Expiration        int32  `json:"expiration"`
	TokensRefreshTime string `json:"tokens_refresh_time"`
	FlushBucketsTime  string `json:"flush_buckets_time"`
}

type Config struct {
	RedisConfig       RedisConfig       `json:"redis"`
	RateLimiterConfig RateLimiterConfig `json:"rate_limiter"`
}

func NewConfig() Config {
	path := os.Getenv("CONFIG_PATH")
	content, err := os.Open(path)
	if err != nil {
		log.Fatalf("Could not open %s", path)
	}
	defer content.Close()

	byteValue, _ := io.ReadAll(content)

	var config Config
	json.Unmarshal(byteValue, &config)

	return config
}

func (c Config) GetEnvValue(name string) string {
	return os.Getenv(name)
}
