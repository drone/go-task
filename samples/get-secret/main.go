package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cgi"
)

// 1. run $ docker-compose up
// 2. open http://localhost:8200/ui/vault/secrets/secret/list
// 3. login with token 'root'
// 4. create secret at path 'docker'
// 5. create secret data with key 'username'
// 6. create secret data with key 'password'

func main() {
	http.HandleFunc("/", handle)
	cgi.Serve(http.DefaultServeMux)
}

type input struct {
	Path      string `json:"path"`
	Key       string `json:"key"`
	Namespace string `json:"namespace"`
	Token     string `json:"token"`
}

type output struct {
	Value string `json:"value"`
}

type vault struct {
	Data struct {
		Data map[string]string `json:"data"`
	} `json:"data"`
}

func handle(w http.ResponseWriter, r *http.Request) {
	// unmarshal the input
	in := new(input)
	json.NewDecoder(r.Body).Decode(in)

	// construct the secret url
	url := fmt.Sprintf("http://127.0.0.1:8200/v1/secret/data/%s", in.Path)

	// construct the http request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if in.Namespace != "" {
		req.Header.Add("X-Vault-Namespace", in.Namespace)
	}
	if in.Token != "" {
		req.Header.Add("X-Vault-Token", in.Token)
	}

	// make the http request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// return the error back to the caller
	if res.StatusCode > 299 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// decode the response
	out := new(vault)
	err = json.NewDecoder(res.Body).Decode(out)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// extract the secret value
	value, ok := out.Data.Data[in.Key]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// write the key to th response
	json.NewEncoder(w).Encode(&output{
		Value: value,
	})
}
