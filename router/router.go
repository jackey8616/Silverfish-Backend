package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	api "silverfish/router/api"
	interf "silverfish/router/interface"
	silverfish "silverfish/silverfish"
	entity "silverfish/silverfish/entity"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Router export
type Router struct {
	recaptchaPrivateKey *string
	auth                *BlueprintAuth
	admin               *BlueprintAdmin
	user                *BlueprintUser
	api                 interf.IBlueprint
}

// NewRouter export
func NewRouter(
	recaptchaPrivateKey *string,
	auth *silverfish.Auth,
	admin *silverfish.Admin,
	user *silverfish.User,
	novel *silverfish.Novel,
	comic *silverfish.Comic,
) *Router {
	rr := new(Router)
	rr.recaptchaPrivateKey = recaptchaPrivateKey
	rr.auth = NewBlueprintAuth(auth, rr)
	rr.admin = NewBlueprintAdmin(auth, admin, novel, comic, rr)
	rr.user = NewBlueprintUser(auth, user, rr)
	rr.api = api.NewBlueprintAPI(auth, user, novel, comic, rr)
	return rr
}

// RouteRegister export
func (rr *Router) RouteRegister(parentRouter *mux.Router) {
	router := parentRouter
	router.HandleFunc("/", rr.root)

	rr.auth.RouterRegiter(router)
	rr.admin.RouteRegister(router)
	rr.user.RouteRegister(router)
	rr.api.RouteRegister(router)
}

// VerifyRecaptcha export
func (rr *Router) VerifyRecaptcha(token *string) (bool, error) {
	url := fmt.Sprintf(
		"https://www.google.com/recaptcha/api/siteverify?secret=%s&response=%s",
		*rr.recaptchaPrivateKey,
		*token)
	res, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	result := new(entity.RecaptchaResponse)
	json.NewDecoder(res.Body).Decode(&result)
	return result.Success, errors.New(strings.Join(result.ErrorCodes, " , "))
}

func (rr *Router) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.Header().Set("Content-Type", "application/json")
		js, _ := json.Marshal(map[string]bool{"Success": true})
		w.Write(js)
	}
}
