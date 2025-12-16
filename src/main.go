package main

import (
	"fmt"
	"matheusflix/hls-streaming-server/src/application"
	"os"
)

func main() {
	fmt.Println("Starting HLS Streaming Server...")

	err := application.Run()

	if err != nil {
		fmt.Println("Error running application:", err)
		os.Exit(1)
	}
}
