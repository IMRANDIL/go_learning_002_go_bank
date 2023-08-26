package main

import (
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
	server := newAPIServer(":8080", store) // Adjust the address as needed
	server.run()
}
