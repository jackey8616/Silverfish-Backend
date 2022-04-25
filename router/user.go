package router

import (
	"encoding/json"
	"net/http"

	interf "silverfish/router/interface"
	silverfish "silverfish/silverfish"
	entity "silverfish/silverfish/entity"

	"github.com/gorilla/mux"
)

// BlueprintUser export
type BlueprintUser struct {
	auth   *silverfish.Auth
	user   *silverfish.User
	router interf.IRouter
	route  string
}

// NewBlueprintUser export
func NewBlueprintUser(
	auth *silverfish.Auth,
	user *silverfish.User,
	router interf.IRouter,
) *BlueprintUser {
	bpu := new(BlueprintUser)
	bpu.auth = auth
	bpu.user = user
	bpu.route = "/user"
	bpu.router = router
	return bpu
}

// RouteRegister export
func (bpu *BlueprintUser) RouteRegister(parentRouter *mux.Router) {
	router := parentRouter.PathPrefix(bpu.route).Subrouter()
	router.HandleFunc("", bpu.root)
	router.HandleFunc("/", bpu.root)
	router.HandleFunc("/bookmark", bpu.bookmark).Methods("GET")
	router.HandleFunc("s/bookmark", bpu.bookmark).Methods("GET")
}

func (bpu *BlueprintUser) root(w http.ResponseWriter, r *http.Request) {}

func (bpu *BlueprintUser) bookmark(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("Authorization")
	w.Header().Set("Content-Type", "application/json")

	session, err := bpu.auth.GetSession(&sessionToken)
	response := new(entity.APIResponse)
	if err != nil {
		response = entity.NewAPIResponse(nil, err)
	} else {
		result, err := bpu.user.GetUserBookmark(session.GetAccount())
		response = entity.NewAPIResponse(result, err)
	}
	js, _ := json.Marshal(response)
	w.Write(js)
}
