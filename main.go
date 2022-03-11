package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"
)

//Creating struct to hold the work request
type WorkRequest struct {
	Name  string
	Delay time.Duration
}

// A buffered channel that we can send the work request on.
var WorkQueue = make(chan WorkRequest, 100)

// A collector function recieves client request for work, build a work request that a worker can understand and finally pushes the work onto the end of the work queue
func Collector(w http.ResponseWriter, r *http.Request) {
	//Make sure we can only be called with an HTTP POST request.
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//Parse the delay
	delay, err := time.ParseDuration(r.FormValue("delay"))
	if err != nil {
		http.Error(w, "Bad delay value: "+err.Error(), http.StatusBadRequest)
		return
	}

	//Check to make sure the delay is anywhere from 1 to 10 seconds.
	if delay.Seconds() < 1 || delay.Seconds() > 10 {
		http.Error(w, "The delay must be between 1 and 10 seconds, inclusively.", http.StatusBadRequest)
	}

	//Retrieve the person's name from the request
	name := r.FormValue("name")

	//create a work request with the name and delay field.
	work := WorkRequest{Name: name, Delay: delay}

	//Push the work onto the queue
	WorkQueue <- work

	//Letting the user that the resource has been created.
	w.WriteHeader(http.StatusCreated)
	return
}

var (
	Nworkers = flag.Int("n", 4, "The number of workers to start")
	HTTPAddr = flag.String("http", "127.0.0.1:8080", "Address to listen for HTTP requests on")
)

func main() {

	//Parse the command-line flags
	flag.Parse()

	//Start the dispatcher
	StartDispatcher(*Nworkers)

	//Register our collector as an HTTP handler function.
	fmt.Println("Registering the collector")
	http.HandleFunc("/work", Collector)

	//Start the HTTP Server
	fmt.Println("HTTP server listening on", *HTTPAddr)
	if err := http.ListenAndServe(*HTTPAddr, nil); err != nil {
		fmt.Println(err.Error())
	}

}
