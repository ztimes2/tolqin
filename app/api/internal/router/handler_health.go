package router

import "net/http"

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	// TODO respond with a dedicated response body containing information about
	// the application's version.

	// Simply indicates if the server is up and running.
	w.WriteHeader(http.StatusOK)
}
