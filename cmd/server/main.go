package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/luis13005/ratelimiter/internal/config"
	"github.com/luis13005/ratelimiter/internal/limiter"
	"github.com/luis13005/ratelimiter/internal/middleware"
	"github.com/luis13005/ratelimiter/internal/storage"
)

func main() {
	cfg, err := config.LoadConfig("../../example.env")
	if err != nil {
		panic(err)
	}

	redisClient := storage.NewClientRedis(cfg)
	rl := limiter.NewLimiter(redisClient, cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	handler := middleware.RateLimiter(rl)(mux)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Servidor rodando em %s", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
