package webhook

import "github.com/gorilla/mux"

// Handler is the interface that webhook proxies should implement so that they can be hooked into the server
type Handler interface {
	Name() string
	Route() string
	Register(*mux.Router)
	Usage() string
}
