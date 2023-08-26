package main

func main() {
	server := newAPIServer(":8080") // Adjust the address as needed
	server.run()
}
