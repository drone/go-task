package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cgi"
)

func main() {
	http.HandleFunc("/", handle)
	cgi.Serve(http.DefaultServeMux)
}

type input struct {
	ID int `json:"id"`
}

func handle(w http.ResponseWriter, r *http.Request) {
	// unmarshal the input
	in := new(input)
	json.NewDecoder(r.Body).Decode(in)

	// request the user details
	res, err := http.Get(fmt.Sprintf("https://jsonplaceholder.typicode.com/users/%d", in.ID))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// write the status to the response
	w.WriteHeader(res.StatusCode)

	// copy the repository details to the cgi response
	io.Copy(w, res.Body)
}
