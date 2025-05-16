package configs

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type RateLimiterConfig struct {
	Default struct {
		Capacity   int     `mapstructure:"capacity"`
		RefillRate float64 `mapstructure:"refill_rate"`
	} `mapstructure:"default"`
	IPBased struct {
		Capacity   int     `mapstructure:"capacity"`
		RefillRate float64 `mapstructure:"refill_rate"`
	} `mapstructure:"ip_based"`
	SpecialClients []struct {
		ID         string  `mapstructure:"id"`
		Capacity   int     `mapstructure:"capacity"`
		RefillRate float64 `mapstructure:"refill_rate"`
	} `mapstructure:"special_clients"`
}

type Config struct {
	ProxyPort   string
	BackendURLs string
	RateLimiter RateLimiterConfig
}

func Load() *Config {
	cfg := &Config{}
	if err := initConfig(); err != nil {
		log.Println("error while init config")
	}

	cfg.ProxyPort = viper.Get("proxy_port").(string)
	if cfg.ProxyPort == "" {
		if envPort := os.Getenv("PROXY_PORT"); envPort != "" {
			cfg.ProxyPort = envPort
		} else {
			log.Fatal("empty proxy port")
		}
	}

	cfg.BackendURLs = viper.Get("backend_urls").(string)
	if cfg.BackendURLs == "" {
		if envBackends := os.Getenv("BACKEND_URLS"); envBackends != "" {
			cfg.BackendURLs = envBackends
		} else {
			log.Fatal("empty backend urls")
		}
	}

	if err := viper.UnmarshalKey("rate_limiter", &cfg.RateLimiter); err != nil {
		log.Fatal("failed to load rate limiter config: ", err)
	}

	return cfg
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
