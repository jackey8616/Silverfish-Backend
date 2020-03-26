package interf

import "net/http"

// IRouter export
type IRouter interface {
	VerifyRecaptcha(token *string) (bool, error)
	FetcherList(w http.ResponseWriter, r *http.Request)
}
