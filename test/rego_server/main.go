package main

import (
	"io"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

func handleDocker(w http.ResponseWriter, r *http.Request) {
	log.Println("received request")

	f, err := os.Open("docker.rego")
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	b, err := io.Copy(w, f)
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else if b == 0 {
		log.Error("no bytes copied from docker.rego")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func handleTerraform(w http.ResponseWriter, r *http.Request) {
	log.Println("received request")

	f, err := os.Open("terraform.rego")
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	b, err := io.Copy(w, f)
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else if b == 0 {
		log.Error("no bytes copied from terraform.rego")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func handleYaml(w http.ResponseWriter, r *http.Request) {
	log.Println("received request")

	f, err := os.Open("yaml.rego")
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	b, err := io.Copy(w, f)
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else if b == 0 {
		log.Error("no bytes copied from yaml.rego")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func main() {
	mux := &http.ServeMux{}

	mux.HandleFunc("/docker", handleDocker)
	mux.HandleFunc("/terraform", handleTerraform)
	mux.HandleFunc("/yaml", handleYaml)

	srv := &http.Server{
		Addr:    ":5000",
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}
