package v1

import (
	"encoding/json"
	"net/http"
	silverfish "silverfish/silverfish"
	"strconv"

	"github.com/gorilla/mux"
)

// BlueprintAPIv1 export
type BlueprintAPIv1 struct {
	silverfish     *silverfish.Silverfish
	auth           *silverfish.Auth
	sessionUsecase *silverfish.SessionUsecase
	version        string
	route          string
	comic          *BlueprintComicv1
	novel          *BlueprintNovelv1
}

// NewBlueprintAPIv1 export
func NewBlueprintAPIv1(silverfish *silverfish.Silverfish, sessionUsecase *silverfish.SessionUsecase) *BlueprintAPIv1 {
	ba1 := new(BlueprintAPIv1)
	ba1.silverfish = silverfish
	ba1.auth = silverfish.Auth
	ba1.sessionUsecase = sessionUsecase
	ba1.version = "v1"
	ba1.route = "/" + ba1.version
	ba1.novel = NewBlueprintNovelv1(silverfish, sessionUsecase)
	ba1.comic = NewBlueprintComicv1(silverfish, sessionUsecase)
	return ba1
}

// GetVersion export
func (ba1 *BlueprintAPIv1) GetVersion() string {
	return ba1.version
}

// RouteRegister export
func (ba1 *BlueprintAPIv1) RouteRegister(parentRouter *mux.Router) {
	router := parentRouter.PathPrefix(ba1.route).Subrouter()
	router.HandleFunc("", ba1.Root)
	router.HandleFunc("/", ba1.Root)

	ba1.novel.RouteRegister(router)
	ba1.comic.RouteRegister(router)
}

// Root export
func (ba1 *BlueprintAPIv1) Root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	js, _ := json.Marshal(map[string]string{
		"version": ba1.GetVersion(),
		"Success": strconv.FormatBool(true),
	})
	w.Write(js)
}
