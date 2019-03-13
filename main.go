package main

import (
	"encoding/json"
	"fmt"
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
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"response": "devfile creation failed " + err.Error(),
		})
		return
	}

	client := che_client.CheDefaultClient()
	resp, err := client.CreateWorkspace(dF.String())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
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
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"response": "devfile creation failed " + err.Error(),
		})
		return
	}
	yaml.NewEncoder(w).Encode(dF)
}

func main() {
	port := ":8000"
	r := mux.NewRouter()
	r.HandleFunc("/api/component/{component_name}/workspace", HandlerCreateWorkspace).Methods("PUT")
	r.HandleFunc("/api/component/{component_name}/devfile", HandlerGenerateDevfile).Methods("GET")
	fmt.Println("Started App Service at " + port)
	log.Fatal(http.ListenAndServe(port, r))
}
