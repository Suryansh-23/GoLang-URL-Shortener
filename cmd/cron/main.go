package cron

import (
	"fmt"
	"time"

	"github.com/Suryansh-23/GoLang-URL-Shortener/cmd/api"
)

func Cron(db *api.URLMap, timeDuration time.Duration) {
	for k, v := range *db {
		t, _ := time.Parse(time.RFC822Z, v.Timestamp)
		if time.Since(t) >= timeDuration {
			delete(*db, k)
		}
	}
}

func CronInit(db *api.URLMap, interval string, signal chan bool) {
	timeDuration, _ := time.ParseDuration(interval)

	func() {
		for {
			select {
			case <-signal:
				fmt.Println("-----------:ENDING SERVICES:-----------")
				return
			default:
				Cron(db, timeDuration)
				time.Sleep(timeDuration)
			}
		}
	}()
}
