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
	auth    *silverfish.Auth
	version string
	route   string
	comic   *BlueprintComicv1
	novel   *BlueprintNovelv1
}

// NewBlueprintAPIv1 export
func NewBlueprintAPIv1(
	auth *silverfish.Auth,
	user *silverfish.User,
	novel *silverfish.Novel,
	comic *silverfish.Comic,
) *BlueprintAPIv1 {
	ba1 := new(BlueprintAPIv1)
	ba1.auth = auth
	ba1.version = "v1"
	ba1.route = "/" + ba1.version
	ba1.novel = NewBlueprintNovelv1(auth, user, novel)
	ba1.comic = NewBlueprintComicv1(auth, user, comic)
	return ba1
}

// GetVersion export
func (ba1 *BlueprintAPIv1) GetVersion() string {
	return ba1.version
}

// RouteRegister export
func (ba1 *BlueprintAPIv1) RouteRegister(parentRouter *mux.Router) {
	router := parentRouter.PathPrefix(ba1.route).Subrouter()
	router.HandleFunc("", ba1.root)
	router.HandleFunc("/", ba1.root)

	ba1.novel.RouteRegister(router)
	ba1.comic.RouteRegister(router)
}

func (ba1 *BlueprintAPIv1) root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	js, _ := json.Marshal(map[string]string{
		"version": ba1.GetVersion(),
		"Success": strconv.FormatBool(true),
	})
	w.Write(js)
}
