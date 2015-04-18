package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"

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
	var found bool
	var emacsConfigRepo github.Repository
	for _, r := range rs {
		if *r.Name == ".emacs.d" {
			emacsConfigRepo = r
			found = true
			break
		}
	}
	if !found {
		io.WriteString(w, "<p>Could not find .emacs.d repo</p>")
		return nil
	}
	code, err := getCode(*emacsConfigRepo.CloneURL)
	if err != nil {
		return err
	}
	io.WriteString(w, "<pre>")
	io.WriteString(w, code)
	io.WriteString(w, "</pre>")
	return nil
}

var storeDir = filepath.Join(os.TempDir(), "emacs-code-store")

// just get init.el for now.
func getCode(cloneURL string) (string, error) {
	u, err := url.Parse(cloneURL)
	if err != nil {
		return "", err
	}
	p := filepath.Join(storeDir, u.Host, u.Path)
	if _, err := os.Stat(p); err != nil {
		// Repo doesn't exist, create it.
		if err := os.MkdirAll(p, os.ModePerm); err != nil {
			return "", err
		}
		c := exec.Command("git", "init")
		c.Dir = p
		if o, err := c.CombinedOutput(); err != nil {
			return "", fmt.Errorf("%s, err: %s", o, err)
		}
		c = exec.Command("git", "remote", "add", "origin", cloneURL)
		c.Dir = p

		if o, err := c.CombinedOutput(); err != nil {
			return "", fmt.Errorf("%s, err: %s", o, err)
		}
	}
	c := exec.Command("git", "pull", "origin", "master")
	c.Dir = p
	if o, err := c.CombinedOutput(); err != nil {
		return "", fmt.Errorf("%s, err: %s", o, err)
	}
	f, err := ioutil.ReadFile(filepath.Join(p, "init.el"))
	if err != nil {
		return "", err
	}
	return string(f), nil
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
