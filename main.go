package main

import (
	"log"
	"net/http"
	"os"

	router "silverfish/router"
	silverfish "silverfish/silverfish"
	entity "silverfish/silverfish/entity"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"gopkg.in/mgo.v2"
)

func dbInit(mongoHost *string) *mgo.Session {
	session, err := mgo.Dial(*mongoHost)
	if err != nil {
		log.Fatal(errors.Wrap(err, "...while initing: "))
	}
	err = session.Ping()
	if err != nil {
		log.Fatal(errors.Wrap(err, "... while pinging: "))
	}
	return session
}

func main() {
	configPath := os.Getenv("config")
	config := NewConfig(&configPath)
	log.Printf("ConfigPath: %s, Debug: %t", configPath, config.Debug)

	log.Printf("-> Initing MongoDB with host: %s ...", *config.DbHost)
	session := dbInit(config.DbHost)
	log.Print("<- MongoDB inited!")
	log.Print("-> Initing Silverfish ...")
	userCol := session.DB("silverfish").C("user")
	userColCount, _ := userCol.Count()
	log.Printf("..... User Collection documents count: %d", userColCount)
	novelCol := session.DB("silverfish").C("novel")
	novelColCount, _ := novelCol.Count()
	log.Printf("..... Novel Collection documents count: %d", novelColCount)
	comicCol := session.DB("silverfish").C("comic")
	comicColCount, _ := comicCol.Count()
	log.Printf("..... Comic Collection documents count: %d", comicColCount)
	log.Print("... Collection inited.")
	userInf := entity.NewMongoInf(session, userCol)
	novelInf := entity.NewMongoInf(session, novelCol)
	comicInf := entity.NewMongoInf(session, comicCol)
	log.Print("... Collection Infrastructure inited.")

	silverfishInstance := silverfish.New(config.HashSalt, userInf, novelInf, comicInf)
	sessionUsecase := silverfish.NewSessionUsecase()
	muxRouter := mux.NewRouter()
	router := router.NewRouter(config.RecaptchaKey, silverfishInstance, sessionUsecase)
	log.Print("... Http Router inited.")
	router.RouteRegister(muxRouter)
	log.Print("... Http Router registered.")

	handler := cors.New(cors.Options{
		AllowedOrigins:   config.AllowOrigin,
		AllowedHeaders:   []string{"Authorization"},
		AllowCredentials: false,
		Debug:            config.Debug,
	}).Handler(muxRouter)
	log.Print("... CORS inited.")

	if userColCount == 0 {
		account := "admin"
		password := silverfish.RandomStr(12)
		silverfishInstance.Auth.Register(true, &account, password)
		log.Printf("Detected no first user, create with `admin:%s`", *password)
	}

	log.Print("<- Silverfish inited.")

	log.Print("Everything Inited! HooRay!~ Silverfish!")
	log.Printf("Connect to https://localhost:%s/", config.Port)
	if config.SSL == true {
		log.Fatal(http.ListenAndServeTLS(":"+config.Port, config.SSLPem, config.SSLKey, handler))
	} else {
		log.Fatal(http.ListenAndServe(":"+config.Port, handler))
	}
}
