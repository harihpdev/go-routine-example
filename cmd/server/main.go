package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/harihpdev/go-routine-example/internal/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Your Mac Temp Dir:", os.TempDir())

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, api.NewRouter()); err != nil {
		log.Fatalf("server error: %v", err)
	}

}
