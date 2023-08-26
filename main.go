package main

import (
	"fmt"
	"log"
)

func main() {
	store, err := newPostgesStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("%+v\n", store)
	fmt.Println("Before creating APIServer")
	server := newAPIServer(":8080", store)
	fmt.Println("After creating APIServer")
	server.setupRoutes()
	server.run()
}
