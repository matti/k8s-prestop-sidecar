package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/paulbellamy/ratecounter"
)

type Handler struct{}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/readyz" {
		log.Println("/readyz", "from", r.RemoteAddr, "current rate", rate.Rate(), "requests in", interval)
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.URL.Path == "/waitz" {
		log.Println("/waitz", "from", r.RemoteAddr, "current rate", rate.Rate(), "requests in", interval)
		for {
			if completed {
				break
			}
			time.Sleep(time.Second)
		}

		w.WriteHeader(http.StatusGone)
		return
	}

	rate.Incr(1)

	status := http.StatusOK
	if shutdown {
		status = http.StatusServiceUnavailable
	}

	w.WriteHeader(status)
	body := fmt.Sprintf("%d\n", status)
	w.Write([]byte(body))

	log.Println("hit", r.URL.Path, "from", r.RemoteAddr, "status", status, "current rate", rate.Rate(), "requests in", interval)
}

var shutdown = false
var completed = false

var rate *ratecounter.RateCounter
var interval time.Duration

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	if os.Getenv("LOG") != "yes" {
		log.SetOutput(ioutil.Discard)
	}

	interval = 15 * time.Second
	if s := os.Getenv("INTERVAL"); s != "" {
		if d, err := time.ParseDuration(s); err != nil {
			panic(err)
		} else {
			interval = d
		}
	}
	rate = ratecounter.NewRateCounter(interval)

	var port = "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	handler := &Handler{}
	server := &http.Server{Addr: ":" + port, Handler: handler}
	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	log.Println("k8s-prestop-sidecar started")
	s := <-sigs
	log.Println("received signal", s)
	shutdown = true
	for {
		r := rate.Rate()
		log.Println("shutdown", "current rate", r, "requests in", interval)
		if r == 0 {
			break
		}
		time.Sleep(time.Second)
	}

	log.Println("completed")
	completed = true
	time.Sleep(time.Second)

	log.Println("exiting")
	if err := server.Shutdown(ctx); err != nil {
		panic(err)
	}

	log.Println("bye")
}
