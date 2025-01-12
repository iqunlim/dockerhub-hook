package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"time"
)

var config map[string]*ConfigElement

func StartWebHandler(injectedConfig map[string]*ConfigElement) {
	config = injectedConfig
	port := config["WEB_PORT"].GetParams()[0]
	webEndpoint := config["WEBHOOK_URL"].GetParams()[0]

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
	go RunCmd(payload)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}
// The actual command functionality
func RunCmd(w *WebhookPayload) {
	cmdName := config["ON_WEBHOOK_COMMAND"].GetParams()[0]
	cmdExec := config["ON_WEBHOOK_COMMAND"].GetParams()[1:]
	cmd := exec.Command(cmdName, cmdExec...)

	// Set up the run log for this request
	formattedTime := time.Now().Format(time.RFC3339)
	formattedTimeStdout := fmt.Sprintf("./runlogs/run-%s.log", formattedTime)
	formattedTimeStderr := fmt.Sprintf("./runlogs/run-%s.err.log", formattedTime)
	runlogs, err := os.OpenFile(formattedTimeStdout, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0447)
	runlogsErr, err := os.OpenFile(formattedTimeStderr, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0447)
	if err != nil {
		log.Println("Error in opening log files")
		log.Println(err)
		return
	}
	defer runlogs.Close()
	defer runlogsErr.Close()

	// Set command stdout to these files
	cmd.Stdout = io.Writer(runlogs)
	cmd.Stderr = io.Writer(runlogsErr)


	// Check against whitelist
	if !slices.Contains(config["WHITELISTED_REPOSITORIES"].Params, w.Repository.RepoName) {
		fmt.Fprintln(cmd.Stderr, "The received repository was not in the whitelist")
		return
	}
	// run the command
	fmt.Fprintln(cmd.Stdout, "++++Execution output++++")
	err = cmd.Run()
	if err != nil {
		fmt.Fprintln(cmd.Stderr, "Error running command")
		fmt.Fprintln(cmd.Stderr, err.Error())
		return
	}
	fmt.Fprintln(cmd.Stdout, "++++++++++++++++++++++++")

	// Success?
	fmt.Fprintf(cmd.Stdout, "The run for %s has completed successfully.", cmd.Args)
}