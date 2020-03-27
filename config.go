package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
)

// Config export
type Config struct {
	Debug        bool     `json:"debug"`
	SSL          bool     `json:"ssl"`
	SSLPem       string   `json:"ssl_pem"`
	SSLKey       string   `json:"ssl_key"`
	Port         string   `json:"port"`
	DbHost       *string  `json:"db_host"`
	HashSalt     *string  `json:"hash_salt"`
	RecaptchaKey *string  `json:"recaptcha_key"`
	AllowOrigin  []string `json:"allow_origins"`
}

// NewConfig export
func NewConfig(path *string) *Config {
	dbHost := "mongo:27107"
	hashSalt := "THIS_IS_A_VERY_COMPLICATED_HASH_SALT_FOR_SILVERFISH_BACKEND"

	jsonFile, err := os.Open(*path)
	if err != nil {
		logrus.Fatal(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	c := &Config{
		Debug:       false,
		SSL:         true,
		SSLPem:      "./server.pem",
		SSLKey:      "./server.key",
		Port:        "8080",
		DbHost:      &dbHost,
		HashSalt:    &hashSalt,
		AllowOrigin: []string{"https://silverfish.cc"},
	}
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	if c.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	json.Unmarshal(byteValue, c)

	if *c.HashSalt == "THIS_IS_A_VERY_COMPLICATED_HASH_SALT_FOR_SILVERFISH_BACKEND" {
		logrus.Println("You are using default `hash_salt`, maybe change one?")
	}
	if *c.RecaptchaKey == "" {
		logrus.Fatal("env recaptcha_key is needed.")
	}
	return c
}
