package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/component-to-devfile/oc_client"

	"github.com/component-to-devfile/che_client"
	"github.com/component-to-devfile/translate"
	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

// HandlerCreateWorkspace create workspace using provided component CR
func HandlerCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	ocClient := oc_client.OcClient{}
	cr, err := ocClient.GetComponentDesc(v["component_name"])

	t := translate.Translater{}
	t.Input = cr
	dF, err := t.Convert()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"response": "devfile creation failed " + err.Error(),
		})
		return
	}

	client := che_client.CheDefaultClient()
	resp, err := client.CreateWorkspace(dF.String())
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"response": "devfile creation failed " + err.Error(),
		})
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.Write(resp)
}

// HandlerGenerateDevfile generate che supported devfile out of component CR
func HandlerGenerateDevfile(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	ocClient := oc_client.OcClient{}
	cr, err := ocClient.GetComponentDesc(v["component_name"])

	t := translate.Translater{}
	t.Input = cr
	dF, err := t.Convert()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"response": "devfile creation failed " + err.Error(),
		})
		return
	}
	yaml.NewEncoder(w).Encode(dF)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/api/workspace/{component_name}", HandlerCreateWorkspace).Methods("PUT")
	r.HandleFunc("/api/devfile/{component_name}", HandlerGenerateDevfile).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", r))
}
