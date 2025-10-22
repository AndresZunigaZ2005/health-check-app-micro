package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/example/monitor/internal/api"
	"github.com/example/monitor/internal/checker"
	"github.com/example/monitor/internal/notifier"
	"github.com/example/monitor/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	pg := os.Getenv("POSTGRES_URL")
	if pg == "" {
		log.Fatal("POSTGRES_URL not set. Example: postgres://user:pass@host:5432/dbname?sslmode=disable")
	}

	st, err := store.New(pg)
	if err != nil {
		log.Fatalf("store init: %v", err)
	}
	defer st.Close()

	n := notifier.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	checker.Start(ctx, st, n)

	h := api.NewAPI(st)
	mux := http.NewServeMux()
	mux.HandleFunc("/register", h.RegisterHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || r.URL.Path == "/health/" {
			h.GetAllHealth(w, r)
			return
		}
		h.GetOneHealth(w, r)
	})

	srv := &http.Server{Addr: ":" + port, Handler: mux}
	go func() {
		log.Println("monitor listening on :" + port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("shutting down")
	ctxSh, cancelSh := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelSh()
	srv.Shutdown(ctxSh)
}
