package main;

import (
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
	"crypto/sha1"
	"regexp"
	"bytes"
	"flag"
	"os"
	"strings"
	"encoding/hex"
	"github.com/gorilla/mux"
)

var (
	secret = ""
)

// POST /put
// body: secret \n data
// response: hash
func handlePut(w http.ResponseWriter, r *http.Request) {
	log.Printf("Method: %s, URI: %s", r.Method, r.RequestURI)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	sepPos := bytes.IndexByte(body, '\n')
	if sepPos == -1 {
		http.Error(w, "malformed request", http.StatusBadRequest)
		return
	}

	line := string(body[:sepPos])
	line = strings.Trim(line, " \t\f\n")
	if line != secret {
		http.Error(w, "invalid secret", http.StatusBadRequest)
		return
	}

	data := body[sepPos+1:]

	h := sha1.New()
	h.Write(data)
	hash := hex.EncodeToString(h.Sum(nil))

    err = ioutil.WriteFile(hash, data, 0644)
	if err != nil {
		http.Error(w, "unable to write to disk", http.StatusInternalServerError)
		return
	}

	log.Printf(":: PUT File %s, %d bytes", hash, len(data))
	fmt.Fprintf(w, "%s\n", hash)
}

// GET /get/<hash>
// response: data
func handleGet(w http.ResponseWriter, r *http.Request) {
	log.Printf("Method: %s, URI: %s", r.Method, r.RequestURI)

	vars := mux.Vars(r)

	hash := strings.ToLower(vars["hash"])
	ok, err := regexp.MatchString("^[a-z0-9]+$", hash)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if ! ok {
		http.Error(w, "bad hash", http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadFile(vars["hash"])
	if err != nil {
		http.Error(w, "error reading or finding file", http.StatusBadRequest)
		return
	}
	log.Printf(":: GET File %s, %d bytes", hash, len(data))
	fmt.Fprintf(w, "%s", data)
}

// POST /delete
// body: secret \n hash
// response: ok/error
func handleDelete(w http.ResponseWriter, r *http.Request) {
	log.Printf("Method: %s, URI: %s", r.Method, r.RequestURI)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	sepPos := bytes.IndexByte(body, '\n')
	if sepPos == -1 {
		http.Error(w, "malformed request", http.StatusBadRequest)
		return
	}

	line := string(body[:sepPos])
	line = strings.Trim(line, " \t\f\n")
	if line != secret {
		http.Error(w, "invalid secret", http.StatusBadRequest)
		return
	}

	hashBytes := body[sepPos+1:]

	hash := string(hashBytes[:])
	hash = strings.TrimRight(strings.ToLower(hash), " \t\f\n")

	ok, err := regexp.MatchString("^[a-z0-9]+$", hash)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if ! ok {
		http.Error(w, "bad hash", http.StatusBadRequest)
		return
	}

	err = os.Remove(hash)
	if err != nil {
		http.Error(w, "error deleting file", http.StatusBadRequest)
		return
	}

	log.Printf(":: DELETE File %s", hash)
	fmt.Fprintf(w, "%s\n", "ok")
}

func main() {
	addrPtr := flag.String("addr", ":7878", "binding address")
	dirPtr := flag.String("dir", ".", "data directory")
	secretPtr := flag.String("secret", "", "file containing the secret")
	prefixPtr := flag.String("prefix", "", "url prefix")
	certPtr := flag.String("cert", "", "path to ssl certificate")
	keyPtr := flag.String("key", "", "path to ssl private key")
	flag.Parse()

	data, err := ioutil.ReadFile(*secretPtr)
	if err != nil {
		fmt.Printf("unable to read secret file '%s'\n", *secretPtr);
		os.Exit(1)
	}
	secret = string(data[:])
	secret = strings.Trim(secret, " \t\f\n")

	err = os.Chdir(*dirPtr)
	if err != nil {
		fmt.Printf("unable to chdir to '%s'\n", *dirPtr);
		os.Exit(1)
	}

	r := mux.NewRouter()
	if *prefixPtr != "" {
		r = r.PathPrefix(*prefixPtr).Subrouter()
	}
	r.HandleFunc("/put", handlePut).Methods("POST");
	r.HandleFunc("/get/{hash}", handleGet).Methods("GET");
	r.HandleFunc("/delete", handleDelete).Methods("POST");

	if *certPtr != "" || *keyPtr != "" {
		http.ListenAndServeTLS(*addrPtr, *certPtr, *keyPtr, r)
	} else {
		http.ListenAndServe(*addrPtr, r)
	}
}
