package main

import (
	"context"
	"net/http"
	"os"
	"silverfish/router"
	"silverfish/silverfish"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

func CreateLocalClient() *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolver(aws.EndpointResolverFunc(
			func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{URL: "http://localhost:8000"}, nil
			})),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: "dummy", SecretAccessKey: "dummy", SessionToken: "dummy",
				Source: "Hard-coded credentials; values are irrelevant for local DynamoDB",
			},
		}),
	)
	if err != nil {
		panic(err)
	}

	return dynamodb.NewFromConfig(cfg)
}

func main() {
	configPath := os.Getenv("config")
	config := NewConfig(&configPath)
	logrus.Printf("ConfigPath: %s, Debug: %t", configPath, config.Debug)

	logrus.Printf("-> Initing MongoDB with host: %s ...", *config.DbHost)
	session := CreateLocalClient()
	logrus.Print("<- MongoDB inited!")
	logrus.Print("-> Initing Silverfish ...")

	silverfishInstance := silverfish.New(config.HashSalt, config.CrawlDuration, session)
	muxRouter := mux.NewRouter()
	router := router.NewRouter(
		config.RecaptchaKey,
		silverfishInstance.Auth,
		silverfishInstance.Admin,
		silverfishInstance.User,
		silverfishInstance.Novel,
		silverfishInstance.Comic,
	)
	logrus.Print("... Http Router inited.")
	router.RouteRegister(muxRouter)
	logrus.Print("... Http Router registered.")

	handler := cors.New(cors.Options{
		AllowedOrigins:   config.AllowOrigin,
		AllowedHeaders:   []string{"Authorization"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowCredentials: false,
		Debug:            config.Debug,
	}).Handler(muxRouter)
	logrus.Print("... CORS inited.")

	// if userColCount == 0 {
	// 	account := "admin"
	// 	password := silverfish.RandomStr(12)
	// 	silverfishInstance.Auth.Register(true, &account, password)
	// 	logrus.Printf("Detected no first user, create with `admin:%s`", *password)
	// }

	logrus.Print("<- Silverfish inited.")

	logrus.Print("Everything Inited! HooRay!~ Silverfish!")
	logrus.Printf("Connect to https://localhost:%s/", config.Port)
	if config.SSL == true {
		logrus.Fatal(http.ListenAndServeTLS(":"+config.Port, config.SSLPem, config.SSLKey, handler))
	} else {
		logrus.Fatal(http.ListenAndServe(":"+config.Port, handler))
	}
}
