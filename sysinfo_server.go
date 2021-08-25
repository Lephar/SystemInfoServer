package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/Lephar/SystemInfoServer/timeutil"
)

type Endpoint struct {
	path     string
	info     string
	callback func(w http.ResponseWriter, r *http.Request)
}

var portString string
var initializationMessage string
var homepageMessageTemplate string

const applicationVersion = "v0.3.0"

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

func sendResponse(w http.ResponseWriter, message string) {
	if _, err := fmt.Fprintln(w, message); err != nil {
		log.Fatalln(err)
	}
}

func homepageCallback(w http.ResponseWriter, r *http.Request) {
	log.Println("Homepage request")
	message := strings.ReplaceAll(homepageMessageTemplate, "${URL}", r.Host)
	sendResponse(w, message)
}

func versionCallback(w http.ResponseWriter, _ *http.Request) {
	log.Println("Version request")
	sendResponse(w, applicationVersion)
}

func parseSystemdOutput(array []byte) string {
	var kernelTimeValue float64
	var userspaceTimeValue float64

	tokens := strings.Split(string(array), " ")

	for i := 0; i < len(tokens); i++ {
		if tokens[i] == "(kernel)" {
			kernelTimeValue = timeutil.ParseDuration(tokens[i-1])
		} else if tokens[i] == "(userspace)" {
			userspaceTimeValue = timeutil.ParseDuration(tokens[i-1])
		}
	}

	kernelTime := timeutil.FormatDuration(kernelTimeValue)
	userspaceTime := timeutil.FormatDuration(userspaceTimeValue)
	totalTime := timeutil.FormatDuration(kernelTimeValue + userspaceTimeValue)

	return fmt.Sprintf("%s (kernel) + %s (userspace) = %s (total)", kernelTime, userspaceTime, totalTime)
}

func durationCallback(w http.ResponseWriter, _ *http.Request) {
	log.Println("Duration request")
	cmd := exec.Command("systemd-analyze", "time")

	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalln(err)
	} else {
		message := parseSystemdOutput(out)
		sendResponse(w, message)
	}
}

func initialize() {
	if len(os.Args) == 1 {
		portString = "8080"
	} else {
		portString = os.Args[1]
	}

	initializationMessage = "Server is ready at port " + portString + ", endpoints are:"
	homepageMessageTemplate = "Visit"

	for i := 0; i < len(endpoints); i++ {
		endpoint := &endpoints[i]

		if endpoint.path == "/" {
			continue
		}

		initializationMessage += " " + endpoint.path
		homepageMessageTemplate += " ${URL}" + endpoint.path + " for " + endpoint.info

		if i < len(endpoints)-1 {
			homepageMessageTemplate += ","
		} else {
			homepageMessageTemplate += "."
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
	log.Fatalln(http.ListenAndServe(":"+portString, nil))
}

//TODO: Add shutdown
func main() {
	initialize()
	registerCallbacks()
	startServer()
}
