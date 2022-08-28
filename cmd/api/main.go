package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

type URLMap map[string]URLItem

type URLItem struct {
	ToURL     string `json:"ToURL"`
	Timestamp string `json:"Timestamp"`
	Link      string `json:"Link"`
	Redirects int    `json:"Redirects"`
}

func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
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
	fmt.Println(availChars)
	return link
}

func homePage(w http.ResponseWriter, r *http.Request, urlMap URLMap) {
	ip := GetIP(r)
	vars := mux.Vars(r)
	link := vars["link"]
	redirect := urlMap[link].ToURL

	if link != "" && redirect != "" {
		fmt.Printf("Redirecting to %s...\n", redirect)
		http.Redirect(w, r, redirect, http.StatusSeeOther)
		return
	}

	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Printf("[%s]: homePage\n", ip)
}

func handleRequests(urlMap URLMap) {
	muxRouter := mux.NewRouter().StrictSlash(true)

	muxRouter.HandleFunc("/{link}", func(w http.ResponseWriter, r *http.Request) {
		homePage(w, r, urlMap)
	})

	// http.HandleFunc("/")
	log.Fatal(http.ListenAndServe("192.168.0.197:8080", muxRouter))
}

func main() {

	db := URLMap{
		"3141592": URLItem{
			ToURL: "https://www.wikipedia.org",
		},
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
	handleRequests(db)
}
