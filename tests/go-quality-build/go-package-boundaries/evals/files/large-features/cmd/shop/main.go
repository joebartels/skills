package main

import (
	"example.com/commerce/internal/catalog"
	"log"
	"net/http"
)

func main() {
	s := &catalog.Service{Items: map[string]catalog.Item{"book": {SKU: "book", Price: 250}}}
	log.Fatal(http.ListenAndServe(":8080", s))
}
