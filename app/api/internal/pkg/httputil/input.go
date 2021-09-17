package httputil

import (
	"errors"
	"net/http"
	"strconv"
)

func QueryParam(r *http.Request, key string) string {
	if r.Form == nil {
		_ = r.ParseForm()
	}
	return r.FormValue(key)
}

var ErrEmptyParam = errors.New("empty parameter")

func QueryParamInt(r *http.Request, key string) (int, error) {
	v := QueryParam(r, key)
	if v == "" {
		return 0, ErrEmptyParam
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func QueryParamFloat(r *http.Request, key string) (float64, error) {
	v := QueryParam(r, key)
	if v == "" {
		return 0, ErrEmptyParam
	}

	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, err
	}

	return f, nil
}
