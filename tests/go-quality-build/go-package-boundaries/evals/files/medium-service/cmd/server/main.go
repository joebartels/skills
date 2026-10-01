package main

import (
	"example.com/dispatch"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	s := &dispatch.Service{Dir: os.Getenv("RECORD_DIR"), NotifyURL: os.Getenv("NOTIFY_URL"), Client: &http.Client{Timeout: 5 * time.Second}}
	log.Fatal(http.ListenAndServe(":8080", s))
}
