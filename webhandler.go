package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

var Config map[string]*ConfigElement

func StartWebHandler(config map[string]*ConfigElement) {
	Config = config
	port := Config["WEB_PORT"].GetParams()[0]
	webEndpoint := Config["WEBHOOK_URL"].GetParams()[0]

	http.HandleFunc(webEndpoint, WebHookHandler)
	fmt.Printf("Starting webserver on port: %s...\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		panic(err)
	}
}

func WebHookHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Read body data
	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error Reading Request Body. Try again.", http.StatusBadRequest)
		log.Printf("Error Reading Request Body. Read body = %s", string(bodyData))
		return
	}

	// Marshall to the WebhookPayload struct
	var payload *WebhookPayload
	if err := json.Unmarshal(bodyData, &payload); err != nil {
		http.Error(w, "Error Reading Request Body. Try again.", http.StatusBadRequest)
		log.Printf("Error Unmarshalling json from the Request Body. Read body = %s", string(bodyData))
		return
	}

	log.Printf("Webhook Received from repository \"%v\"", payload.Repository.RepoName)
	// Run the payloads command 
	go payload.RunCmd()

	fmt.Fprintf(w, "OK")
}