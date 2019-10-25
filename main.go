package main

import (
	"log"
	"net/http"
	"os"

	silverfish "silverfish/silverfish"
	entity "silverfish/silverfish/entity"

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
	log.Print("... MongoDB inited!")
	return session
}

func main() {
	configPath := os.Getenv("config")
	config := NewConfig(&configPath)
	log.Printf("ConfigPath: %s, Debug: %t", configPath, config.Debug)

	log.Printf("Initing MongoDB with host: %s ...", *config.DbHost)
	session := dbInit(config.DbHost)
	userInf := entity.NewMongoInf(session, session.DB("silverfish").C("user"))
	novelInf := entity.NewMongoInf(session, session.DB("silverfish").C("novel"))
	comicInf := entity.NewMongoInf(session, session.DB("silverfish").C("comic"))
	log.Printf("Collections inited.")
	silverfish := silverfish.New(config.HashSalt, userInf, novelInf, comicInf)
	log.Printf("Silverfish inited.")

	router := NewRouter(config.RecaptchaKey, silverfish)
	mux := http.NewServeMux()
	mux.HandleFunc("/", router.Root)
	mux.HandleFunc("/auth/status", router.AuthStatus)
	mux.HandleFunc("/auth/register", router.AuthRegister)
	mux.HandleFunc("/auth/login", router.AuthLogin)
	mux.HandleFunc("/auth/logout", router.AuthLogout)
	mux.HandleFunc("/auth/isAdmin", router.AuthIsAdmin)
	mux.HandleFunc("/user/bookmark", router.UserBookmark)
	mux.HandleFunc("/admin/fetchers", router.FetcherList)
	mux.HandleFunc("/api/v1/novels", router.V1Novels)
	mux.HandleFunc("/api/v1/novel", router.V1Novel)
	/* TODO: route should be /api/v1/novel/chapter
	This change will need to update Frontend's api calling. */
	mux.HandleFunc("/api/v1/chapter", router.V1NovelChapter)
	mux.HandleFunc("/api/v1/novel/chapter", router.V1NovelChapter)
	mux.HandleFunc("/api/v1/comics", router.V1Comics)
	mux.HandleFunc("/api/v1/comic", router.V1Comic)
	mux.HandleFunc("/api/v1/comic/chapter", router.V1ComicChapter)
	log.Printf("Http Route inited.")

	log.Println(config.AllowOrigin)
	handler := cors.New(cors.Options{
		AllowedOrigins:   config.AllowOrigin,
		AllowedHeaders:   []string{"Authorization"},
		AllowCredentials: false,
		Debug:            config.Debug,
	}).Handler(mux)
	log.Printf("CORS inited.")

	log.Printf("Everything Inited! HooRay!~ Silverfish!")
	log.Printf("Connect to https://localhost:%s/", config.Port)
	if config.SSL == true {
		log.Fatal(http.ListenAndServeTLS(":"+config.Port, config.SSLPem, config.SSLKey, handler))
	} else {
		log.Fatal(http.ListenAndServe(":"+config.Port, handler))
	}
}
