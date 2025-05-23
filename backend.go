package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	port := flag.Int("port", 8081, "Port to serve on")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hostname, _ := os.Hostname()

		// Add a delay to simulate processing time
		time.Sleep(100 * time.Millisecond)

		fmt.Fprintf(w, "Host: %s, port: %d, path: %s\n", hostname, *port, r.URL.Path)
	})

	log.Printf("Started at port: %d\n", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
