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

	parallel := 250
	if p := os.Getenv("READ_CONCURRENCY"); p != "" {
		var err error
		parallel, err = strconv.Atoi(p)
		if err != nil {
			log.Fatal("Invalid value for READ_CONCURRENCY, must be an integer")
		}
	}

	fmt.Println("Reading", os.Getenv("READ_PATH"), "with concurrency of", parallel)

	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	waiter := make(chan interface{})
	var wg sync.WaitGroup

	for i := 0; i < parallel; i++ {
		wg.Add(1)
		go thing(client, &wg, waiter)
	}

	time.Sleep(1 * time.Second)

	start := time.Now()
	close(waiter)

	wg.Wait()
	stop := time.Now()

	fmt.Println(stop.Sub(start))
}

func thing(client *api.Client, wg *sync.WaitGroup, waiter <-chan interface{}) {
	<-waiter

	defer wg.Done()

	{
		start := time.Now()
		_, err := client.Logical().Read(os.Getenv("READ_PATH"))
		duration := time.Now().Sub(start)
		if err != nil {
			log.Println("Error:", err, "in", duration)
			return
		}
		log.Println("Success in", duration)
	}
	// {
	// 	_, err := client.Logical().Read("secret/api/API_BASE_URL")
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	// }
}
