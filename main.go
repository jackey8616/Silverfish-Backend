package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/axgle/mahonia"
	cors "github.com/rs/cors"
)

func main() {
	allowOrigins, allowCredentials, debug :=
		[]string{"*"}, true, true
	mux := http.NewServeMux()

	mux.HandleFunc("/", helloWorld)
	mux.HandleFunc("/proxy", proxy)

	handler := cors.New(cors.Options{
		AllowedOrigins:   allowOrigins,
		AllowCredentials: allowCredentials,
		Debug:            debug,
	}).Handler(mux)
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	js, _ := json.Marshal(map[string]bool{"Success": true})
	w.Write(js)
}

func proxy(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	data := map[string]string{}
	decoder.Decode(&data)
	res, err := http.Get(data["proxy_url"])
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	resbody := mahonia.NewDecoder("gbk").NewReader(res.Body)
	rawHTML, err := ioutil.ReadAll(resbody)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	js, _ := json.Marshal(map[string]string{"Rtn": string(rawHTML)})
	w.Write(js)
}
