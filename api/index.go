package handler

import (
	"net/http"
	"sync"
	"url-shortener/internal/app"
)

var (
	serverOnce sync.Once
	serverErr  error
	server     *app.Server
)

func Handler(w http.ResponseWriter, r *http.Request) {
	serverOnce.Do(func() {
		server, serverErr = app.NewServer()
	})

	if serverErr != nil {
		http.Error(w, serverErr.Error(), http.StatusInternalServerError)
		return
	}

	server.DynamicHandler().ServeHTTP(w, r)
}
