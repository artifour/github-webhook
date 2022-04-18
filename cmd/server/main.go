package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/artifour/github-webhook/internal/config"
	"github.com/artifour/github-webhook/internal/middleware"
)

type configuration struct {
	Port         string            `json:"port"`
	Secret       string            `json:"secret"`
	Repositories map[string]string `json:"repositories"`
}

func main() {
	config.Init("conf.json", configuration{})

	runHttpServer()
}

func runHttpServer() {
	mux := http.NewServeMux()
	mux.Handle("/", middleware.GitHubMiddleware(middleware.DefaultMiddleware()))

	addr := fmt.Sprintf(":%s", config.Get("port"))
	log.Printf("Listening on %s...\n", addr)
	err := http.ListenAndServe(addr, mux)
	log.Fatal(err)
}
