package routes

import (
	"net/http"

	contractHTTP "github.com/donnigundala/dgcore/contracts/http"
)

// Register registers the web routes.
func Register(router contractHTTP.Router) {
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from DG Framework!"))
	})
}
