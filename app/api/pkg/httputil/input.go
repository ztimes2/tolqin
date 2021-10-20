package httputil

import (
	"errors"
	"net/http"
	"strconv"
)

// QueryParam retrieves a query parameter from the given request by the given key.
// An empty string is returned if the query parameter is not found.
func QueryParam(r *http.Request, key string) string {
	if r.Form == nil {
		_ = r.ParseForm()
	}
	return r.FormValue(key)
}

// ErrParamNotFound is used when a parameter is not found.
var ErrParamNotFound = errors.New("parameter not found")

// QueryParamInt retrieves a query parameter from the given request by the given
// key and parses it as an integer number. ErrParamNotFound error is returned if
// the query parameter is not found.
func QueryParamInt(r *http.Request, key string) (int, error) {
	v := QueryParam(r, key)
	if v == "" {
		return 0, ErrParamNotFound
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}

	return i, nil
}

// QueryParamFloat retrieves a query parameter from the given request by the given
// key and parses it as a float number. ErrParamNotFound error is returned if the
// query parameter is not found.
func QueryParamFloat(r *http.Request, key string) (float64, error) {
	v := QueryParam(r, key)
	if v == "" {
		return 0, ErrParamNotFound
	}

	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, err
	}

	return f, nil
}
