package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
)

const applicationVersion = "v0.1.0"

func homepageCallback(w http.ResponseWriter, r *http.Request) {
	log.Println("Homepage request")
	fmt.Fprintln(w, "Visit /version for server version info and /duration for boot-up time")
}

func versionCallback(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Version request")
	fmt.Fprintln(w, applicationVersion)
}

//TODO: Parse the output
func durationCallback(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Duration request")
	cmd := exec.Command("systemd-analyze", "time")
	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(w, string(out))
}

func registerCallbacks() {
	http.HandleFunc("/", homepageCallback)
	http.HandleFunc("/version", versionCallback)
	http.HandleFunc("/duration", durationCallback)
}

//TODO: Add time info to logs
func startServer() {
	log.Println("Server ready, endpoints: /version and /duration")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	registerCallbacks()
	startServer()
}
