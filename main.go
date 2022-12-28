package main

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func root(w http.ResponseWriter, r *http.Request) {
	log.Println("/", playing)

	if playing {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "k8s-prestop-sidecar: playing\n")
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		io.WriteString(w, "k8s-prestop-sidecar: stopped\n")
	}
}

func stop(w http.ResponseWriter, r *http.Request) {
	log.Println("/stop", playing, "start")
	playing = false
	time.Sleep(31 * time.Second)
	log.Println("/stop", playing, "done")
}

func play(w http.ResponseWriter, r *http.Request) {
	log.Println("/play", playing)
	playing = true
}

var playing = true

func main() {
	if os.Getenv("LOG") != "yes" {
		log.SetOutput(ioutil.Discard)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	var port = "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", root)
	mux.HandleFunc("/stop", stop)
	mux.HandleFunc("/play", play)

	server := &http.Server{Addr: ":" + port, Handler: mux}
	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	log.Println("k8s-prestop-sidecar started")
	<-sigs

	if err := server.Shutdown(ctx); err != nil {
		panic(err)
	}
}
