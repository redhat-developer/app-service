// Package appserver provides a Server structure that let's you package things you
// need all around your service
package appserver

import (
	"encoding/json"
	"net/http"

	"github.com/ghodss/yaml"
	"github.com/gorilla/mux"
)

// HandleStatus returns the handler function for the /status endpoint
func (srv *AppServer) HandleStatus() http.HandlerFunc {
	response := struct {
		Commit    string `json:"commit"`
		BuildTime string `json:"build_time"`
		StartTime string `json:"start_time"`
	}{
		BuildTime: BuildTime,
		StartTime: StartTime,
		Commit:    Commit,
	}
	return func(w http.ResponseWriter, r *http.Request) {
		format := mux.Vars(r)["format"]
		var err error
		var bytes []byte
		switch format {
		case "yaml":
			bytes, err = yaml.Marshal(&response)
		case "json":
			fallthrough
		default:
			bytes, err = json.Marshal(&response)
			format = "json"
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Set(http.CanonicalHeaderKey("Content-Type"), "application/"+format)
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}
}
