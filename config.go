package main

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

// Config export
type Config struct {
	Debug         bool
	SSL           bool
	SSLPem        string
	SSLKey        string
	Port          string
	DbHost        string
	HashSalt      string
	RecaptchaKey  string
	AllowOrigin   []string
	CrawlDuration int
}

func getEnvWithDefault[T int | float64 | bool | string](key string, fallback T) T {
	value, exist := os.LookupEnv(key)
	if !exist {
		return fallback
	}

	switch any(fallback).(type) {
	case int:
		parsedValue, err := strconv.Atoi(value)
		if err != nil {
			return fallback
		}
		return any(parsedValue).(T)
	case float64:
		parsedValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fallback
		}
		return any(parsedValue).(T)
	case bool:
		parsedValue, err := strconv.ParseBool(value)
		if err != nil {
			return fallback
		}
		return any(parsedValue).(T)
	case string:
		return any(value).(T)
	default:
		return fallback
	}
}

// NewConfig export
func NewConfig() *Config {
	c := &Config{
		Debug:         getEnvWithDefault("DEBUG", false),
		SSL:           getEnvWithDefault("SSL", true),
		SSLPem:        getEnvWithDefault("SSL_PEM", "./server.pem"),
		SSLKey:        getEnvWithDefault("SSL_KEY", "./server.key"),
		Port:          getEnvWithDefault("PORT", "8080"),
		DbHost:        getEnvWithDefault("DB_HOST", "mongo:27017"),
		HashSalt:      getEnvWithDefault("HASH_SALT", "THIS_IS_A_VERY_COMPLICATED_HASH_SALT_FOR_SILVERFISH_BACKEND"),
		RecaptchaKey:  getEnvWithDefault("RECAPTCHA_KEY", ""),
		AllowOrigin:   []string{"https://silverfish.cc"},
		CrawlDuration: 60,
	}
	if c.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	if c.HashSalt == "THIS_IS_A_VERY_COMPLICATED_HASH_SALT_FOR_SILVERFISH_BACKEND" {
		logrus.Println("You are using default `hash_salt`, maybe change one?")
	}
	if c.RecaptchaKey == "" {
		logrus.Fatal("env recaptcha_key is needed.")
	}
	return c
}
