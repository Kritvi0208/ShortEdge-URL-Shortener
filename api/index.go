package handler

import (
	"net/http"
	"sync"
	"url-shortener/appcore"
)

var (
	serverOnce sync.Once
	serverErr  error
	server     *appcore.Server
)

func Handler(w http.ResponseWriter, r *http.Request) {
	serverOnce.Do(func() {
		server, serverErr = appcore.NewServer()
	})

	if serverErr != nil {
		http.Error(w, serverErr.Error(), http.StatusInternalServerError)
		return
	}

	server.DynamicHandler().ServeHTTP(w, r)
}
