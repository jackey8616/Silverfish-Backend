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
	sf                  *silverfish.Silverfish
	sessionUsecase      *silverfish.SessionUsecase
	api                 interf.IBlueprint
	user                *BlueprintUser
}

// NewRouter export
func NewRouter(recaptchaPrivateKey *string, sf *silverfish.Silverfish, sessionUsecase *silverfish.SessionUsecase) *Router {
	rr := new(Router)
	rr.recaptchaPrivateKey = recaptchaPrivateKey
	rr.sf = sf
	rr.sessionUsecase = sessionUsecase
	rr.user = NewBlueprintUser(sf, rr, sessionUsecase)
	rr.api = api.NewBlueprintAPI(sf, rr, sessionUsecase)
	return rr
}

// RouteRegister export
func (rr *Router) RouteRegister(parentRouter *mux.Router) {
	router := parentRouter
	router.HandleFunc("/", rr.Root)
	router.HandleFunc("/admin/fetchers", rr.FetcherList)

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

// Root export
func (rr *Router) Root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.Header().Set("Content-Type", "application/json")
		js, _ := json.Marshal(map[string]bool{"Success": true})
		w.Write(js)
	}
}

// FetcherList export
func (rr *Router) FetcherList(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		sessionToken := r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")

		session, err := rr.sessionUsecase.GetSession(&sessionToken)
		response := new(entity.APIResponse)
		if err != nil {
			response = entity.NewAPIResponse(nil, err)
		} else if isAdmin, _ := rr.sf.Auth.IsAdmin(session.GetAccount()); isAdmin == false {
			response = entity.NewAPIResponse(nil, errors.New("Only Admin allowed"))
		} else {
			fetcherLists := rr.sf.GetLists()
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
