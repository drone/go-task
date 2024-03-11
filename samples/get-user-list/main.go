package main

import (
	"io"
	"net/http"
	"net/http/cgi"
)

func main() {
	http.HandleFunc("/", handle)
	cgi.Serve(http.DefaultServeMux)
}

func handle(w http.ResponseWriter, r *http.Request) {
	// request the user details
	res, err := http.Get("https://jsonplaceholder.typicode.com/users")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// write the status to the response
	w.WriteHeader(res.StatusCode)

	// copy the repository details to the cgi response
	io.Copy(w, res.Body)
}
