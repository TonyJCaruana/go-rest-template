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
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
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

// Problem detail as defined in RFC7807 specification ( https://tools.ietf.org/html/rfc7807 )
type problemDetail struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
}

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

func performRequest(id string) (body string, status int, err error) {

	// TODO! - Write logic here to perform service function and return either result or error in response
	msg := "Service running!"

	if rand.Intn(10)%2 == 0 {
		// We have a problem so generate a problem detail ( N.B Depending on the issue you may return any 400 - 500 status to provide addtional information)
		problem := &problemDetail{Type: "http://example.org/error/500", Title: "The service is currently un-available", Status: http.StatusInternalServerError, Detail: "Unable to resolve DNS hostname MyService", Instance: "http://example.org/myservice/error/500"}
		document, _ := json.Marshal(problem)
		return string(document), http.StatusInternalServerError, errors.New("Service un-available!")
	}
	// All is OK so just return the response to the caller
	return "{ \"ID\" : \"" + id + "\", \"Message\" : \"" + msg + "\", \"Status\" : \"" + http.StatusText(http.StatusOK) + "\" }", http.StatusOK, nil

}

func readinessProbe(response http.ResponseWriter, request *http.Request) {

	// Tells container orchestrator such as Mesos/Marathon OR Kubernetes or discovery system such as Consul OR ZooKeeper
	// that we are avaialable to serve traffic, and that can communicate with downstream services such as databases or queues

	// TODO! - Write logic here to determine application readiness for your service
	writeStandardHeaders(response, http.StatusOK, "application/json")
}

func requestHandler(response http.ResponseWriter, request *http.Request) {

	id := mux.Vars(request)["id"]

	if body, status, err := performRequest(id); err != nil {
		writeStandardHeaders(response, status, "application/problem+json")
		fmt.Fprintf(response, body)
	} else {
		writeStandardHeaders(response, status, "application/json")
		fmt.Fprintf(response, body)
	}

}

func livenessProbe(response http.ResponseWriter, request *http.Request) {

	// Tells container orchestrator such as Mesos/Marathon OR Kubernetes or discovery system such as Consul OR ZooKeeper
	// that we are still alive, haven't crashed, and don't need to be re-started. Equivalant to a HTTP Ping
	writeStandardHeaders(response, http.StatusOK, "application/json")
}

func writeStandardHeaders(response http.ResponseWriter, status int, contentType string) {

	// set common headers and response code
	response.Header().Set("content-type", contentType+";charset=utf-8")
	response.Header().Set("Content-Language", "en")
	response.Header().Set("Cache-Control", "no-cache")
	response.Header().Set("Pragma", "no-cache")
	response.Header().Set("Expires", "-1")

	response.WriteHeader(status)
}
