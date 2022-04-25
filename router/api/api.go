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
	auth  *silverfish.Auth
	route string
	v1    *v1.BlueprintAPIv1
}

// NewBlueprintAPI export
func NewBlueprintAPI(
	auth *silverfish.Auth,
	user *silverfish.User,
	novel *silverfish.Novel,
	comic *silverfish.Comic,
	router interf.IRouter,
) *BlueprintAPI {
	ba := new(BlueprintAPI)
	ba.auth = auth
	ba.route = "/api"
	ba.v1 = v1.NewBlueprintAPIv1(auth, user, novel, comic)
	return ba
}

// RouteRegister export
func (ba *BlueprintAPI) RouteRegister(parentRouter *mux.Router) {
	router := parentRouter.PathPrefix(ba.route).Subrouter()
	router.HandleFunc("", ba.root)
	router.HandleFunc("/", ba.root)

	ba.v1.RouteRegister(router)
}

func (ba *BlueprintAPI) root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	js, _ := json.Marshal(map[string]bool{
		"v1":      true,
		"Success": true,
	})
	w.Write(js)
}
