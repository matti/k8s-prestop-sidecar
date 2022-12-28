package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func root(w http.ResponseWriter, r *http.Request) {
	if playing {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "k8s-prestop-sidecar: playing\n")
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		io.WriteString(w, "k8s-prestop-sidecar: stopped\n")
	}
}

func stop(w http.ResponseWriter, r *http.Request) {
	playing = false
	time.Sleep(31 * time.Second)
}

func play(w http.ResponseWriter, r *http.Request) {
	playing = true
}

var playing = true

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	mux := http.NewServeMux()
	mux.HandleFunc("/", root)
	mux.HandleFunc("/stop", stop)
	mux.HandleFunc("/play", play)

	server := &http.Server{Addr: ":8080", Handler: mux}
	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	<-sigs

	if err := server.Shutdown(ctx); err != nil {
		panic(err)
	}
}
