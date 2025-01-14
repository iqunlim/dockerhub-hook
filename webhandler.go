package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"slices"
)

var config map[string]*ConfigElement

func StartWebHandler(injectedConfig map[string]*ConfigElement) {
	config = injectedConfig

	for key, val := range config {
		logger.Debug("Configuration", "Key", key, "Val", val)
	}
	port := config["WEB_PORT"].First()
	webEndpoint := config["WEBHOOK_URL"].First()

	http.HandleFunc(webEndpoint, WebHookHandler)
	logger.Info("Starting webserver...", "port", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		logger.Error("Fatal Error", "source", "StartWebHandler.http.ListenAndServe", "err", err)
		os.Exit(1)
	}
}

func WebHookHandler(w http.ResponseWriter, r *http.Request) {

	// Read body data
	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error Reading Request Body. Try again.", http.StatusBadRequest)
		logger.Error("Error Reading Request Body.", "source", "WebHookHandler.io.ReadAll", "body", string(bodyData))
		return
	}

	// Marshall to the WebhookPayload struct
	var payload *WebhookPayload
	if err := json.Unmarshal(bodyData, &payload); err != nil {
		http.Error(w, "Error Reading Request Body. Try again.", http.StatusBadRequest)
		logger.Error("Error Unmarshalling json from the Request Body", "source", "WebHookHandler.io.ReaAll", "body", string(bodyData))
		return
	}

	logger.Info("Webhook Received", "repositoryName", payload.Repository.RepoName)
	// Run the payloads command 
	go RunCmd(payload)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
	logger.Info("WebHandler Returned OK")
}
// The actual command functionality
func RunCmd(w *WebhookPayload) {
	cmdName := config["ON_WEBHOOK_COMMAND"].First()
	cmdExec := config["ON_WEBHOOK_COMMAND"].After()
	cmd := exec.Command(cmdName, cmdExec...)

	// Set up the run log for this request

	// Set command stdout to these files
	//cmd.Stdout = io.Writer(runlogs)
	//cmd.Stderr = io.Writer(runlogsErr)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("Failed to create stdout pipe", "source", "RunCmd.cmd.stdoutPipe", "err", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		logger.Error("Failed to create stderr pipe", "source", "RunCmd.cmd.stderrPipe", "err", err)
	}

	// Check against whitelist
	if !slices.Contains(config["WHITELISTED_REPOSITORIES"].Params, w.Repository.RepoName) {
		logger.Warn("The received repository was not in the whitelist", "repository", w.Repository.RepoName)
		return
	}
	// run the command
	err = cmd.Start()
	if err != nil {
		logger.Error("Error starting command", "command", cmdName, "args", cmdExec, "err", err, "source", "RunCmd.cmd.Start")
		return
	}

	go pipeToLogger(stdoutPipe, logger, slog.LevelInfo)
	go pipeToLogger(stderrPipe, logger, slog.LevelError)

	if err := cmd.Wait(); err != nil {
		logger.Error("Command execution failed", "source", "RunCmd.cmd.Wait", "err", err, "command", cmdName, "args", cmdExec)
		return
	} else {
		logger.Info("The command has completed successfully.")
	}
}