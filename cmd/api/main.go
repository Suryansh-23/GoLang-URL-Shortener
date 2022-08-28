package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

const hostURL string = "localhost:8080"

type URLMap map[string]URLItem

type URLItem struct {
	ToURL     string `json:"ToURL"`
	Timestamp string `json:"Timestamp"`
	Link      string `json:"Link"`
	Redirects int    `json:"Redirects"`
}

type ShortenRequest struct {
	URL string `json:"URL"`
}

type ShortenResponse struct {
	URL string `json:"URL"`
}

func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func GetTimestamp() string {
	return time.Now().Format(time.RFC822Z)
}

func Cleanup(db URLMap) {
	dbByte, err := json.Marshal(db)
	if err != nil {
		panic(err)
	}

	os.WriteFile("cmd/db/db.json", dbByte, fs.FileMode(os.O_TRUNC)|fs.FileMode(os.O_WRONLY))
}

func SetupCloseHandler(db URLMap) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		Cleanup(db)
		os.Exit(0)
	}()
}

func getRandomLink() string {
	rand.Seed(time.Now().UnixNano()) // Important Step: Otherwise would generate the same sequence of keys
	availChars := []rune{}
	link := ""

	// 0-9
	for i := 48; i < 58; i++ {
		availChars = append(availChars, rune(i))
	}
	// A-Z
	for i := 65; i < 91; i++ {
		availChars = append(availChars, rune(i))
	}
	// a-z
	for i := 97; i < 123; i++ {
		availChars = append(availChars, rune(i))
	}

	for i := 0; i < 7; i++ {
		link += string(availChars[rand.Intn(len(availChars))]) // Selects a random Index and then the array gives the character at that index
	}
	// fmt.Println(availChars)
	return link
}

func homePage(w http.ResponseWriter, r *http.Request, urlMap URLMap) {
	vars := mux.Vars(r)
	link := vars["link"]
	redirect := urlMap[link].ToURL
	urlItem := urlMap[link]

	if link != "" && redirect != "" {
		urlMap[link] = URLItem{ToURL: urlItem.ToURL, Timestamp: urlItem.Timestamp, Link: urlItem.Link, Redirects: urlItem.Redirects + 1}
		fmt.Printf("[%s]: %s\n", GetIP(r), r.URL)
		fmt.Printf("Redirecting to %s...\n", redirect)
		http.Redirect(w, r, redirect, http.StatusSeeOther)
		return
	}

	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Printf("[%s]: %s\n", GetIP(r), r.URL)
}

func shorten(w http.ResponseWriter, r *http.Request, urlMap URLMap) {
	respBody, _ := io.ReadAll(r.Body)
	var body ShortenRequest

	err := json.Unmarshal(respBody, &body)
	if err != nil || body.URL == "" {
		http.Error(w, "Request Body Invalid", http.StatusBadRequest)
	}

	shortURL := getRandomLink()
	timestamp := GetTimestamp()
	u, err := url.Parse(body.URL)
	if err != nil {
		http.Error(w, "Request Body Invalid", http.StatusBadRequest)
	}
	if !u.IsAbs() {
		u.Scheme = "https"
	}

	urlMap[shortURL] = URLItem{ToURL: u.String(), Timestamp: timestamp, Link: hostURL + "/" + shortURL, Redirects: 0}
	json.NewEncoder(w).Encode(ShortenResponse{URL: hostURL + "/" + shortURL})
	fmt.Printf("[%s]: %s\n", GetIP(r), r.URL)
}

func handleRequests(urlMap URLMap) {
	muxRouter := mux.NewRouter().StrictSlash(true)

	muxRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to the HomePage!")
		fmt.Printf("[%s]: %s\n", GetIP(r), r.URL)
	})

	muxRouter.HandleFunc("/{link}", func(w http.ResponseWriter, r *http.Request) {
		homePage(w, r, urlMap)
	}).Methods("GET")

	muxRouter.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Inside Shorten")
		shorten(w, r, urlMap)
	}).Methods("POST")
	log.Fatal(http.ListenAndServe(hostURL, muxRouter))
}

func main() {

	db := URLMap{}

	_, err := os.Stat("cmd/db/db.json")
	if errors.Is(err, os.ErrNotExist) {
		os.Create("cmd/db/db.json")
	}

	file, err := os.ReadFile("cmd/db/db.json")
	if err != nil {
		panic(err)
	}

	if string(file) != "" {
		err = json.Unmarshal(file, &db)
		if err != nil {
			panic(err)
		}
	}

	// fmt.Println(getRandomLink())
	// fmt.Println(getRandomLink())
	SetupCloseHandler(db)

	handleRequests(db)
}
