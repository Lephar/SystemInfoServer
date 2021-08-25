package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Endpoint is the struct type that contains all the necessary
// info that is needed to add new endpoints to the server
type Endpoint struct {
	path     string
	info     string
	callback func(w http.ResponseWriter, r *http.Request)
}

var portString string
var initializationMessage string
var homepageMessageTemplate string

const applicationVersion = "v0.3.0"

// To add new endpoints to the server, simply add a new Endpoint
// instance at the end of this array and implement the corresponding
// callback function to reference from the new entry
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

	// Homepage message is modified depending on the URL that the user used to
	// connect to the server, e.g. if connected using localhost:8080, the URLs
	// will be seen as localhost:8080/version and localhost:8080/duration but
	// if it's time.server.edu the URLs will be time.server.edu/version and
	// time.server.edu/duration
	message := strings.ReplaceAll(homepageMessageTemplate, "${URL}", r.Host)
	sendResponse(w, message)
}

func versionCallback(w http.ResponseWriter, _ *http.Request) {
	log.Println("Version request")
	sendResponse(w, applicationVersion)
}

// systemd-analyze outputs loader and firmware times, which we don't need
// alongside kernel and userspace times. This method parses this outputs
// and calculates the sum of kernel and userspace times and return the
// formatted output to be sent to the client
func parseSystemdOutput(array []byte) (string, error) {
	var kernelTime time.Duration
	var userspaceTime time.Duration
	var err error

	tokens := strings.Split(string(array), " ")

	for i := 0; i < len(tokens); i++ {
		// Output is as follows: "... <kernelTime> (kernel) + <userspaceTime> (userspace) ..."
		// So we look for the kernel and userspace tokens, and the token before that is the
		// time value that we are looking for
		switch tokens[i] {
		case "(kernel)":
			kernelTime, err = time.ParseDuration(tokens[i-1])
		case "(userspace)":
			userspaceTime, err = time.ParseDuration(tokens[i-1])
		}

		if err != nil {
			return "", err
		}
	}

	kernelTimeSeconds := kernelTime.Seconds()
	userspaceTimeSeconds := userspaceTime.Seconds()
	totalTimeSeconds := kernelTimeSeconds + userspaceTimeSeconds

	return fmt.Sprintf("%Gs (kernel) + %Gs (userspace) = %Gs (total)", kernelTimeSeconds, userspaceTimeSeconds, totalTimeSeconds), nil
}

func durationCallback(w http.ResponseWriter, _ *http.Request) {
	log.Println("Duration request")
	// time is optional and the default command line argument of systemd-analyze,
	// so it is not necessary to add it but this makes it more future-proof in case
	// default parameters are changed in the next releases
	cmd := exec.Command("systemd-analyze", "time")

	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalln(err)
	} else {
		message, err := parseSystemdOutput(out)
		if err != nil {
			log.Fatalln(err)
		}
		sendResponse(w, message)
	}
}

func initialize() {
	// Default port number is 8080 if no port is specified while starting the server
	// application. Simply supply the desired port number as the first command line
	// argument if you wish to run the server application on another port
	if len(os.Args) == 1 {
		portString = "8080"
	} else {
		portString = os.Args[1]
	}

	// Server info and homepage messages are initialized here by iterating all the
	// endpoints in the endpoints array, so it's not necessary to make changes here
	// if you wish to add new endpoints to the server
	initializationMessage = "Server is ready at port " + portString + ", endpoints are:"
	homepageMessageTemplate = "Visit"

	for i := 0; i < len(endpoints); i++ {
		endpoint := &endpoints[i]

		if endpoint.path == "/" {
			continue
		}

		initializationMessage += " " + endpoint.path

		// ${URL} part will be replaced on every homepage request with the hostname
		// the client used to connect to the server
		homepageMessageTemplate += " ${URL}" + endpoint.path + " for " + endpoint.info

		if i < len(endpoints)-1 {
			homepageMessageTemplate += ","
		} else {
			homepageMessageTemplate += "."
		}
	}
}

// This function registers the corresponding callback functions of endpoint paths
// by iterating the endpoints array
// TODO: Add error handlers for the callbacks
func registerCallbacks() {
	for _, endpoint := range endpoints {
		http.HandleFunc(endpoint.path, endpoint.callback)
	}
}

// TODO: Add graceful shutdown via SIGTERM
// TODO: Use server multiplexer instead of default http server
func startServer() {
	log.Println(initializationMessage)
	log.Fatalln(http.ListenAndServe(":"+portString, nil))
}

func main() {
	initialize()
	registerCallbacks()
	startServer()
}
