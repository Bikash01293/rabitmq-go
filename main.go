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

//Creating struct of Worker
type Worker struct {
	ID int
	Work chan WorkRequest
	WorkerQueue chan chan WorkerRequest
	QuitChan chan bool

}

// NewWorker creates, and returns a new Worker object
func NewWorker(id int, workerQueue chan chan WorkerRequest) Worker {
	//Create and return the worker
	worker := Worker {
		ID: id,
		Work: make(chan WorkRequest),
		workerQueue: workerQueue,
		QuitChan: make(chan bool)
	}

	return worker
}

// This function "starts" the worker by starting a goroutine, that is an infinite "for-select" loop.
func (w *Worker) Start() {
	go func() {
		
		for {
			//Add ourselves into the worker queue.
			w.WorkerQueue <- w.Work

			select {
			case work := <- w.Work:
				//Recieve a work request.
				fmt.Printf("worker%d: Recieved work request, delaying for %f seconds", w.Id, work.delay.Seconds())

				time.Sleep(work.Delay)
				fmt.Printf("worker%d: Hello, %s!\n", w.ID, work.Name)

			case <-w.QuitChan:
				//We have been asked to stop
				fmt.Printf("worker%d stopping\n", w.ID)
				return
			}
		}
	}()
}

//Stop tells the worker to stop listening for work requests.
//Note that the worker will only stop "after" it has finished its work.
func (w *Worker) Stop() {
	go func() {
		w.QuitChan <- true
	}
}()

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
