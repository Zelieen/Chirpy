package main

import (
	"net/http"
)

func main() {
	ServeMux := http.NewServeMux()
	Server := &http.Server{}
	Server.Handler = ServeMux
	Server.Addr = ":8080"
	Server.ListenAndServe()
}
