package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var listen = os.Getenv("LISTEN")

func handler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	w.Header().Set("Content-Type", "text/plain")
	sleep := query.Get("sleep")
	if sleep != "" {
		d, err := time.ParseDuration(sleep)
		if err != nil {
			w.WriteHeader(400)
			fmt.Fprint(w, "invalid sleep duration\n")
			return
		}
		time.Sleep(d)
	}

	fmt.Fprintf(w, "%s\n", listen)
}

func main() {
	log.Printf("Listening %s\n", listen)
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(listen, nil))
}
