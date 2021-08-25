package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

type Endpoint struct {
	path     string
	info     string
	callback func(w http.ResponseWriter, r *http.Request)
}

var initializationMessage string
var homepageMessage string

const applicationVersion = "v0.1.0"

var endpoints = []Endpoint{
	{
		"/",
		"homepage",
		homepageCallback,
	}, {
		"/version",
		"application version info",
		versionCallback,
	}, {
		"/duration",
		"system boot-up duration",
		durationCallback,
	},
}

func respondRequest(w http.ResponseWriter, message string) {
	if _, err := fmt.Fprintln(w, message); err != nil {
		log.Fatalln(err)
	}
}

func homepageCallback(w http.ResponseWriter, r *http.Request) {
	log.Println("Homepage request")
	message := strings.ReplaceAll(homepageMessage, "${URL}", r.Host)
	respondRequest(w, message)
}

func versionCallback(w http.ResponseWriter, _ *http.Request) {
	log.Println("Version request")
	respondRequest(w, applicationVersion)
}

//TODO: Parse the output
func durationCallback(w http.ResponseWriter, _ *http.Request) {
	log.Println("Duration request")
	cmd := exec.Command("systemd-analyze", "time")

	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalln(err)
	} else {
		respondRequest(w, string(out))
	}
}

func initialize() {
	initializationMessage = "Server ready, endpoints:"
	homepageMessage = "Visit"

	for i := 0; i < len(endpoints); i++ {
		endpoint := &endpoints[i]

		if endpoint.path == "/" {
			continue
		}

		initializationMessage += " " + endpoint.path
		homepageMessage += " ${URL}" + endpoint.path + " for " + endpoint.info

		if i < len(endpoints)-1 {
			homepageMessage += ","
		} else {
			homepageMessage += "."
		}
	}
}

func registerCallbacks() {
	for _, endpoint := range endpoints {
		http.HandleFunc(endpoint.path, endpoint.callback)
	}
}

func startServer() {
	log.Println(initializationMessage)
	log.Fatalln(http.ListenAndServe(":8080", nil))
}

//TODO: Add shutdown
func main() {
	initialize()
	registerCallbacks()
	startServer()
}
