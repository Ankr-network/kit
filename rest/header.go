package rest

import "net/http"

func addJsonHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json;charset=UFT-8")
}
