package mlog

import (
	"errors"
	"net/http"
	"regexp"
)

var (
	pattern = regexp.MustCompile(`/logs/(\S+)/level`)
)

func Route(mux *http.ServeMux) {
	mux.HandleFunc("/logs/", ServeHTTP)
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ln, err := parseLogName(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	l := Logger(ln)
	l.level.ServeHTTP(w, r)
}

func parseLogName(path string) (string, error) {
	ms := pattern.FindStringSubmatch(path)
	if len(ms) != 2 {
		return "", errors.New("invalid path")
	}
	return ms[1], nil
}
