package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/Suryansh-23/GoLang-URL-Shortener/cmd/api"
	"github.com/Suryansh-23/GoLang-URL-Shortener/cmd/cron"
)

func main() {

	fmt.Println("Inside main")

	db := api.URLMap{}
	signal := make(chan bool)
	var wg sync.WaitGroup

	go api.APInit(&db, &signal)
	wg.Add(1)

	fmt.Println("API Initialized")
	time.Sleep(10 * time.Second) // Buffer Time untill the db is loaded with data

	go cron.CronInit(&db, "1m", signal)
	wg.Add(1)

	wg.Wait()
}
