package interf

import (
	"net/http"

	"github.com/gorilla/mux"
)

// IBlueprint export
type IBlueprint interface {
	RouteRegister(*mux.Router)
}

// IBlueprintAPI export
type IBlueprintAPI interface {
	IBlueprint
	GetVersion() string
	Remove(http.ResponseWriter, *http.Request)
}

// IBlueprintComic export
type IBlueprintComic interface {
	IBlueprint
	comic(http.ResponseWriter, *http.Request)
	listComic(http.ResponseWriter, *http.Request)
	chapter(http.ResponseWriter, *http.Request)
}

// IBlueprintNovel export
type IBlueprintNovel interface {
	IBlueprint
	novel(http.ResponseWriter, *http.Request)
	listNovel(http.ResponseWriter, *http.Request)
	chapter(http.ResponseWriter, *http.Request)
}
