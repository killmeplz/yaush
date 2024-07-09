package main

import (
	"log"
	"net/http"
	"os"

	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"math/rand"
	"time"
)

var ctx = context.Background()

type App struct {
	Router *mux.Router
	Client *redis.Client
}

type URLRequest struct {
	URL string `json:"url"`
}

type URLResponse struct {
	ShortURL string `json:"short_url"`
}

func (a *App) Initialize() {
	a.Client = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	a.Router = mux.NewRouter()
	a.Router.HandleFunc("/shorten", a.shortenURL).Methods("POST")
	a.Router.HandleFunc("/{shortURL}", a.redirect).Methods("GET")
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) shortenURL(w http.ResponseWriter, r *http.Request) {
	var urlRequest URLRequest
	_ = json.NewDecoder(r.Body).Decode(&urlRequest)

	if urlRequest.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	shortURL := generateShortURL()
	err := a.Client.Set(ctx, shortURL, urlRequest.URL, 0).Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := URLResponse{ShortURL: shortURL}
	json.NewEncoder(w).Encode(response)
}

func (a *App) redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortURL := vars["shortURL"]

	url, err := a.Client.Get(ctx, shortURL).Result()
	if err == redis.Nil {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func generateShortURL() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func main() {
	app := &App{}
	app.Initialize()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	app.Run(":" + port)
}
