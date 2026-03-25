package main

import (
	"log"
	"net/http"

	"github.com/adk2004/auction_service/router"
)

func main() {
	r := router.SetupRouter()
	
	log.Printf("Starting server on port 5001...")
	if err := http.ListenAndServe(":5001", r); err != nil {
		panic(err)
	}
}
