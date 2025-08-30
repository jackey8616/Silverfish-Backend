package main

import (
	"net/http"
	"os"

	router "silverfish/router"
	silverfish "silverfish/silverfish"
	entity "silverfish/silverfish/entity"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

func dbInit(mongoHost *string) *mgo.Session {
	session, err := mgo.Dial(*mongoHost)
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "...while initing: "))
	}
	err = session.Ping()
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "... while pinging: "))
	}
	return session
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

	logrus.Printf("-> Initing MongoDB with host: %s ...", config.DbHost)
	session := dbInit(&config.DbHost)
	logrus.Print("<- MongoDB inited!")
	logrus.Print("-> Initing Silverfish ...")
	userCol := session.DB("silverfish").C("user")
	userColCount, _ := userCol.Count()
	logrus.Printf("..... User Collection documents count: %d", userColCount)
	novelCol := session.DB("silverfish").C("novel")
	novelColCount, _ := novelCol.Count()
	logrus.Printf("..... Novel Collection documents count: %d", novelColCount)
	comicCol := session.DB("silverfish").C("comic")
	comicColCount, _ := comicCol.Count()
	logrus.Printf("..... Comic Collection documents count: %d", comicColCount)
	logrus.Print("... Collection inited.")
	userInf := entity.NewMongoInf(session, userCol)
	novelInf := entity.NewMongoInf(session, novelCol)
	comicInf := entity.NewMongoInf(session, comicCol)
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
