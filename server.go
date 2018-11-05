/*This service is intended to be deployed in a container such as Docker and run on a server orchestration framework.
  Here’s what a potential real life request processing failure scenario might look like

   - Readiness probe fails
   - Kubernetes stops routing traffic to the pod.
   - Liveness probe fails.
   - Kubernetes restarts the failed container*.
   - Readiness probe succeeds.
   - Kubernetes starts routing traffic to the pod again.

  Author: Anthony Caruana
*/
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

var (
	addr = "0.0.0.0:50001"
	port = 50001
)

func main() {

	// subcribe to SIGINT/SIGKILL
	osSignalChannel := make(chan os.Signal)
	signal.Notify(osSignalChannel, os.Interrupt, os.Kill)

	// set up handler and ready/live probes for orchestration framework
	router := mux.NewRouter()
	router.HandleFunc("/{id}", requestHandler)
	router.HandleFunc("/live", livenessProbe)
	router.HandleFunc("/ready", readinessProbe)

	fmt.Printf("\n>> Server running on [%d]\n", port)
	fmt.Println("   Press <Ctr-C> to quit...")

	// configure timeouts/address/handler
	srv := &http.Server{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
		Addr:         addr,
		Handler:      router,
	}

	// launch http server
	go func() {
		srv.ListenAndServe()
	}()

	// listen for SIGINT/SIGKILL
	<-osSignalChannel
	fmt.Println("   Server shutting down...")

	// shut down server gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer srv.Shutdown(ctx)
	fmt.Println(">> Server stopped")
	fmt.Println("")
}

func performRequest(id string) (body string, err error) {

	// TODO! - Write logic here to perform service function and return either result or error in response
	msg := "Service running"
	return "{ \"ID\" : \"" + id + "\", \"Message\" : \"" + msg + "\", \"Status\" : \"" + http.StatusText(200) + "\" }", nil
}

func readinessProbe(response http.ResponseWriter, request *http.Request) {

	// Tells container orchestrator such as Mesos/Marathon OR Kubernetes or discovery system such as Consul OR ZooKeeper
	// that we are avaialable to serve traffic, and that can communicate with downstream services such as databases or queues

	// TODO! - Write logic here to determine application readiness for your service
	writeStandardHeaders(response, http.StatusOK)
}

func requestHandler(response http.ResponseWriter, request *http.Request) {

	id := mux.Vars(request)["id"]

	body, err := performRequest(id)
	if err != nil {
		writeStandardHeaders(response, http.StatusInternalServerError)
	} else {
		writeStandardHeaders(response, http.StatusOK)
	}

	fmt.Fprintf(response, body)

}

func livenessProbe(response http.ResponseWriter, request *http.Request) {

	// Tells container orchestrator such as Mesos/Marathon OR Kubernetes or discovery system such as Consul OR ZooKeeper
	// that we are still alive, haven't crashed, and don't need to be re-started. Equivalant to a HTTP Ping
	writeStandardHeaders(response, http.StatusOK)
}

func writeStandardHeaders(response http.ResponseWriter, status int) {

	// set common headers and response code
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Cache-Control", "no-cache")
	response.Header().Set("Pragma", "no-cache")
	response.Header().Set("Expires", "-1")
	response.WriteHeader(status)
}
