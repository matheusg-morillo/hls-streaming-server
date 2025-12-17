package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"matheusflix/hls-streaming-server/src/application"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file", err)
		os.Exit(1)
	}

	fmt.Println("Starting HLS Streaming Server...")

	err = application.Run()
	if err != nil {
		fmt.Println("Error running application:", err)
		os.Exit(1)
	}
}
