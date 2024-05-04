package api

import (
	"net/http"

	"github.com/bongofriend/bookplayer/backend/lib/api/middleware"
	"github.com/bongofriend/bookplayer/backend/lib/config"
)

/*
Custom http.ServeMux with additional methods
TODO: Add additional methods for authentication and routes requiring a authenticated request
*/
type ServiceMux struct {
	http.ServeMux
}

func newServiceMux() *ServiceMux {
	return &ServiceMux{
		ServeMux: *http.NewServeMux(),
	}
}

func GetApiHandler(c config.Config) http.Handler {
	middlewareStack := middleware.CreateMiddlewareStack(middleware.Logging(c))
	mux := newServiceMux()

	return middlewareStack(mux)
}
