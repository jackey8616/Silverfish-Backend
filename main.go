package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/jackey8616/Silverfish-backend/silverfish"
	cors "github.com/rs/cors"
	"gopkg.in/mgo.v2"
)

func modeInit() (bool, string, bool, []string) {
	mode := os.Getenv("mode")
	if mode == "prod" {
		return false, "mongo:27017", false,
			[]string{"https://jackey8616.github.io", "http://jackey8616.github.io", "https://*.clo5de.info", "http://*.clo5de.info"}
	}
	return true, "127.0.0.1:27017", true, []string{"*"}
}

func dbInit(mongoHost string) *mgo.Session {
	session, _ := mgo.Dial(mongoHost)
	return session
}

func main() {
	debug, dbHost, allowCredentials, allowOrigins := modeInit()
	log.Printf("Debug: %t, DbHost: %s", debug, dbHost)
	session := dbInit(dbHost)
	mgoInf := silverfish.NewMongoInf(session, session.DB("silverfish").C("novel"))
	silverfish := silverfish.New(mgoInf)

	mux := http.NewServeMux()
	mux.HandleFunc("/", helloWorld)
	mux.HandleFunc("/proxy", silverfish.Proxy)

	handler := cors.New(cors.Options{
		AllowedOrigins:   allowOrigins,
		AllowCredentials: allowCredentials,
		Debug:            debug,
	}).Handler(mux)

	log.Printf("Everything Inited! HooRay!~ Silverfish!")
	log.Printf("Connect to http://localhost:%d/ backend", 8080)
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	js, _ := json.Marshal(map[string]bool{"Success": true})
	w.Write(js)
}
