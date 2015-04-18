package main

import (
	"io"
	"log"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
)

func serveHome(w http.ResponseWriter, r *http.Request) error {
	io.WriteString(w, "<p>hello emacs</p")
	return nil
}

func serveProfile(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	io.WriteString(w, "<p>Your username is: "+vars["username"]+"</p>")
	c := github.NewClient(nil)
	rs, _, err := c.Repositories.List(vars["username"], nil)
	if err != nil {
		return err
	}
	var repos []github.Repository
	for _, r := range rs {
		if r.Language != nil && *r.Language == "Emacs Lisp" {
			repos = append(repos, r)
		}
	}
	for _, r := range repos {
		io.WriteString(w, "<p>Elisp repo: "+*r.Name+"</p>")
	}
	return nil
}

type handler func(http.ResponseWriter, *http.Request) error

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	return
}

func main() {
	r := mux.NewRouter()
	r.Handle("/", handler(serveHome))
	r.Handle("/{username}", handler(serveProfile))
	url := ":5000"
	log.Printf("Listening on %s", url)
	log.Fatal(http.ListenAndServe(":5000", r))
}
