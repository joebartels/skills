package main

import (
	"log"
	"net/http"
	"os"

	"example.com/fielddesk"
)

func main() {
	root := os.Getenv("FIELDDESK_ROOT")
	if root == "" {
		root = "./data"
	}
	log.Fatal(http.ListenAndServe(":8080", fielddesk.New(root).Handler()))
}
