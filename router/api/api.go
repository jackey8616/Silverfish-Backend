package api

import (
	"encoding/json"
	"net/http"
	v1 "silverfish/router/api/v1"
	interf "silverfish/router/interface"
	silverfish "silverfish/silverfish"

	"github.com/gorilla/mux"
)

// BlueprintAPI export
type BlueprintAPI struct {
	silverfish     *silverfish.Silverfish
	auth           *silverfish.Auth
	sessionUsecase *silverfish.SessionUsecase
	route          string
	v1             *v1.BlueprintAPIv1
}

// NewBlueprintAPI export
func NewBlueprintAPI(silverfish *silverfish.Silverfish, router interf.IRouter, sessionUsecase *silverfish.SessionUsecase) *BlueprintAPI {
	ba := new(BlueprintAPI)
	ba.silverfish = silverfish
	ba.auth = silverfish.Auth
	ba.sessionUsecase = sessionUsecase
	ba.route = "/api"
	ba.v1 = v1.NewBlueprintAPIv1(silverfish, sessionUsecase)
	return ba
}

// RouteRegister export
func (ba *BlueprintAPI) RouteRegister(parentRouter *mux.Router) {
	router := parentRouter.PathPrefix(ba.route).Subrouter()
	router.HandleFunc("", ba.Root)
	router.HandleFunc("/", ba.Root)

	ba.v1.RouteRegister(router)
}

// Root export
func (ba *BlueprintAPI) Root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	js, _ := json.Marshal(map[string]bool{
		"v1":      true,
		"Success": true,
	})
	w.Write(js)
}
