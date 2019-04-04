package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/jackey8616/Silverfish-backend/silverfish"
	"github.com/jackey8616/Silverfish-backend/silverfish/entity"
	"github.com/rs/cors"
	"gopkg.in/mgo.v2"
)

func modeInit() (string, string, bool, string, bool, []string) {
	mode := os.Getenv("mode")
	port := os.Getenv("port")
	if port == "" {
		port = "8080"
	}

	if mode == "prod" {
		return mode, port, false, "mongo:27017", false,
			[]string{"https://jackey8616.github.io", "http://jackey8616.github.io", "https://*.clo5de.info", "http://*.clo5de.info"}
	}
	return mode, port, true, "127.0.0.1:27017", true, []string{"*"}
}

func dbInit(mongoHost string) *mgo.Session {
	session, _ := mgo.Dial(mongoHost)
	return session
}

func main() {
	mode, port, debug, dbHost, allowCredentials, allowOrigins := modeInit()
	log.Printf("Debug: %t, DbHost: %s", debug, dbHost)
	session := dbInit(dbHost)
	mgoInf := entity.NewMongoInf(session, session.DB("silverfish").C("novel"))
	silverfish := silverfish.New(mgoInf, []string{
		"http://www.77xsw.la/book/389/",
		"http://www.77xsw.la/book/11072/",
		"http://www.77xsw.la/book/11198/",
		"http://www.77xsw.la/book/13192/",
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/", helloWorld)
	mux.HandleFunc("/api/v1/novel", silverfish.Router.V1Novel)
	mux.HandleFunc("/api/v1/novels", silverfish.Router.V1Novels)
	mux.HandleFunc("/api/v1/chapter", silverfish.Router.V1Chapter)
	mux.HandleFunc("/api/v1/chapter_new", silverfish.Router.V1ChapterNew)
	mux.HandleFunc("/fetch_novel", silverfish.FetchNovel)
	mux.HandleFunc("/fetch_chapter", silverfish.FetchChapter)

	handler := cors.New(cors.Options{
		AllowedOrigins:   allowOrigins,
		AllowCredentials: allowCredentials,
		Debug:            debug,
	}).Handler(mux)

	log.Printf("Everything Inited! HooRay!~ Silverfish!")
	if mode == "prod" {
		log.Printf("Connect to https://localhost:%s/ backend", port)
		log.Fatal(http.ListenAndServeTLS(":"+port, "server.pem", "server.key", handler))
	}
	log.Printf("Connect to http://localhost:%s/ backend", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	js, _ := json.Marshal(map[string]bool{"Success": true})
	w.Write(js)
}
