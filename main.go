package main

import (
	"context"
	"net/http"
	"os"
	"time"

	router "silverfish/router"
	silverfish "silverfish/silverfish"
	entity "silverfish/silverfish/entity"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func dbInit(mongoHost *string) *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(*mongoHost))
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "...while initing: "))
	}
	if err = client.Ping(ctx, nil); err != nil {
		logrus.Fatal(errors.Wrap(err, "... while pinging: "))
	}
	return client
}

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	configPath := os.Getenv("CONFIG_PATH")
	if len(configPath) > 0 {
		logrus.Printf("Detected config path: %s", configPath)
		err := godotenv.Load(configPath)
		if err != nil {
			logrus.Fatal(err)
		}
	}
	config := NewConfig()
	logrus.Printf("Debug: %t", config.Debug)

	logrus.Printf("-> Initing MongoDB")
	client := dbInit(&config.DbHost)
	logrus.Print("<- MongoDB inited!")
	logrus.Print("-> Initing Silverfish ...")
	db := client.Database("silverfish")
	userInf := entity.NewMongoInf(db.Collection("user"))
	novelInf := entity.NewMongoInf(db.Collection("novel"))
	comicInf := entity.NewMongoInf(db.Collection("comic"))
	userColCount, _ := userInf.CountDocuments()
	logrus.Printf("..... User Collection documents count: %d", userColCount)
	novelColCount, _ := novelInf.CountDocuments()
	logrus.Printf("..... Novel Collection documents count: %d", novelColCount)
	comicColCount, _ := comicInf.CountDocuments()
	logrus.Printf("..... Comic Collection documents count: %d", comicColCount)
	logrus.Print("... Collection Infrastructure inited.")

	silverfishInstance := silverfish.New(&config.HashSalt, config.CrawlDuration, userInf, novelInf, comicInf)
	muxRouter := mux.NewRouter()
	router := router.NewRouter(
		&config.RecaptchaKey,
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
	logrus.Print("... CORS Origins: ", config.AllowOrigin)
	logrus.Print("... CORS inited.")

	if userColCount == 0 {
		account := "admin"
		password := silverfish.RandomStr(12)
		silverfishInstance.Auth.Register(true, &account, password)
		logrus.Printf("Detected no first user, create with `admin:%s`", *password)
	}

	logrus.Print("<- Silverfish inited.")

	logrus.Print("Everything Inited! HooRay!~ Silverfish!")
	logrus.Printf("Connect to https://localhost:%s/", config.Port)
	if config.SSL == true {
		logrus.Fatal(http.ListenAndServeTLS(":"+config.Port, config.SSLPem, config.SSLKey, handler))
	} else {
		logrus.Fatal(http.ListenAndServe(":"+config.Port, handler))
	}
}
