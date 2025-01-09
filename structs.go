package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"slices"
	"time"
)

type WebhookPayload struct {
	CallbackURL string `json:"callback_url"`
	PushData PushData `json:"push_data"`
	Repository RepositoryData `json:"repository"`
}

type PushData struct {
	PushedAt int64  `json:"pushed_at"`
	Pusher   string `json:"pusher"`
	Tag      string `json:"tag"`
}

type RepositoryData struct {
	CommentCount    int    `json:"comment_count"`
	DateCreated     int64  `json:"date_created"`
	Description     string `json:"description"`
	Dockerfile      string `json:"dockerfile"`
	FullDescription string `json:"full_description"`
	IsOfficial      bool   `json:"is_official"`
	IsPrivate       bool   `json:"is_private"`
	IsTrusted       bool   `json:"is_trusted"`
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	Owner           string `json:"owner"`
	RepoName        string `json:"repo_name"`
	RepoURL         string `json:"repo_url"`
	StarCount       int    `json:"star_count"`
	Status          string `json:"status"`
}


// The actual command functionality
func (w *WebhookPayload) RunCmd() {
	cmdName := Config["ON_WEBHOOK_COMMAND"].GetParams()[0]
	cmdExec := Config["ON_WEBHOOK_COMMAND"].GetParams()[1:]
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
	if !slices.Contains(Config["WHITELISTED_REPOSITORIES"].Params, w.Repository.RepoName) {
		fmt.Fprintln(cmd.Stderr, "The received repository was not in the whitelist")
		return
	}
	// run the command
	fmt.Fprintln(cmd.Stdout, "++++Execution output++++")
	err = cmd.Run()
	if err != nil {
		fmt.Fprintln(cmd.Stderr, "Error running command")
		cmd.Stderr.Write([]byte(err.Error()))
		return
	}
	fmt.Fprintln(cmd.Stdout, "++++++++++++++++++++++++")

	// Success?
	fmt.Fprintf(cmd.Stdout, "The run for %s has completed successfully.", cmd.Args)
}