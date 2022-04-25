package router

import (
	"encoding/json"
	"errors"
	"net/http"
	interf "silverfish/router/interface"
	"silverfish/silverfish"
	"silverfish/silverfish/entity"

	"github.com/gorilla/mux"
)

// BlueprintAdmin export
type BlueprintAdmin struct {
	auth   *silverfish.Auth
	admin  *silverfish.Admin
	novel  *silverfish.Novel
	comic  *silverfish.Comic
	router interf.IRouter
	route  string
}

// NewBlueprintAdmin export
func NewBlueprintAdmin(
	auth *silverfish.Auth,
	admin *silverfish.Admin,
	novel *silverfish.Novel,
	comic *silverfish.Comic,
	router interf.IRouter,
) *BlueprintAdmin {
	bpa := new(BlueprintAdmin)
	bpa.auth = auth
	bpa.admin = admin
	bpa.novel = novel
	bpa.comic = comic
	bpa.route = "/admin"
	bpa.router = router
	return bpa
}

// RouteRegister export
func (bpa *BlueprintAdmin) RouteRegister(parentRouter *mux.Router) {
	router := parentRouter.PathPrefix(bpa.route).Subrouter()
	router.HandleFunc("/fetchers", bpa.fetcherList).Methods("GET")
}

// FetcherList export
func (bpa *BlueprintAdmin) fetcherList(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("Authorization")
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:

		session, err := bpa.auth.GetSession(&sessionToken)
		response := new(entity.APIResponse)
		if err != nil {
			response = entity.NewAPIResponse(nil, err)
		} else if isAdmin, _ := bpa.auth.IsAdmin(session.GetAccount()); isAdmin == false {
			response = entity.NewAPIResponse(nil, errors.New("Only Admin allowed"))
		} else {
			fetcherLists := map[string][]string{
				"novels": bpa.novel.GetFetcherNameLists(),
				"comics": bpa.comic.GetFetcherNameLists(),
			}
			response = entity.NewAPIResponse(map[string]interface{}{
				"fetchers": fetcherLists,
			}, nil)
		}
		js, _ := json.Marshal(response)
		w.Write(js)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
