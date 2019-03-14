// NOTE: This file was generated and should be modified by you!

package appserver

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Handle{{.HandlerName}} takes a request with a name and responds with
//
// {"greeting": "Hello, <NAME>!"}
//
// or
//
// {"greeting": "Hallo, <NAME>!"}
//
// depending on what language was given in the route path.
//
// Try this by calling
//
//     curl -v -X POST -d '{"name": "John"}' 0.0.0.0:8080/{{.Path}}/english
//     curl -v -X POST -d '{"name": "John"}' 0.0.0.0:8080/{{.Path}}/german
//
// Now, using the route to parameterize the response isn't particularly elegant
// but I wanted to show you how you can have a variable inside your route path.
func (srv *AppServer) Handle{{.HandlerName}}() http.HandlerFunc {
	request := struct {
		Name string `json:"name"`
	}{}
	response := struct {
		Greeting string `json:"greeting"`
	}{}

	return func(w http.ResponseWriter, r *http.Request) {
		language := mux.Vars(r)["language"]

		// read request
		requestBytes, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, "failed to read input data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		err = json.Unmarshal(requestBytes, &request)
		if err != nil {
			http.Error(w, "failed to parse input data: "+err.Error(), http.StatusBadRequest)
			return
		}

		if strings.TrimSpace(request.Name) == "" {
			http.Error(w, "name property must not be empty in: "+string(requestBytes), http.StatusBadRequest)
			return
		}

		// construct response dependent on URL
		switch language {
		case "english":
			response.Greeting = "Hello " + request.Name + "!"
		case "german":
			response.Greeting = "Hallo " + request.Name + "!"
		}

		bytes, err := json.Marshal(&response)
		if err != nil {
			http.Error(w, "failed to marshal response: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set(http.CanonicalHeaderKey("Content-Type"), "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}
}