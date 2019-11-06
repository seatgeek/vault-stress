package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/vault/api"
)

func main() {
	if path := os.Getenv("READ_PATH"); path == "" {
		log.Fatal("Missing READ_PATH environment - e.g. `db-tzanalytic/creds/read-only`")
	}

	parallel := 10
	if p := os.Getenv("READ_CONCURRENCY"); p != "" {
		var err error
		parallel, err = strconv.Atoi(p)
		if err != nil {
			log.Fatal("Invalid value for READ_CONCURRENCY, must be an integer")
		}
	}

	count := 1000
	if p := os.Getenv("READ_COUNT"); p != "" {
		var err error
		count, err = strconv.Atoi(p)
		if err != nil {
			log.Fatal("Invalid value for READ_COUNT, must be an integer")
		}
	}

	fmt.Println("Reading", count, "secrets from", os.Getenv("READ_PATH"), "with concurrency of", parallel)

	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	waiter := make(chan int, count)
	var wg sync.WaitGroup
	wg.Add(count)

	for i := 0; i < parallel; i++ {
		wg.Add(1)
		go thing(client, &wg, waiter)
	}

	for i := 0; i < count; i++ {
		waiter <- i
	}

	time.Sleep(1 * time.Second)

	start := time.Now()
	wg.Wait()
	close(waiter)

	stop := time.Now()

	fmt.Println(stop.Sub(start))
}

func thing(client *api.Client, wg *sync.WaitGroup, waiter <-chan int) {
	for {
		select {
		case _, ok := <-waiter:
			if !ok {
				return
			}

			start := time.Now()
			_, err := client.Logical().Read(os.Getenv("READ_PATH"))
			duration := time.Now().Sub(start)
			if err != nil {
				log.Println("Error:", err, "in", duration)
			} else {
				log.Println("Success in", duration)
			}

			wg.Done()
		}
	}
}
