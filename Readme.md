# docker-webhook-micro

This is an extremely small, dependency-less, and basic executable made to run a specified command on your system when the Docker Hub pushes a webhook to your url.

## Usage

- clone the repository
- run `go build .`
- run `./dockerhub-hook` in the same folder as your docker-compose file

On Docker Hub:

Find the "webhooks" tab on your Docker Hub repository

![Wehbooks in the nav bar](https://files.iqun.xyz/H06D14V5E1AX/wh1.png)

Set them up as follows (using your defined HTTP endpoint if you have set a custom one)

![Webhooks example](https://files.iqun.xyz/PES0COQOZYXF/wh2.png)

## Environment Variable Configuration:

These can be placed either in a .env file in the same folder as the executable
or they can be set on your filesystem/container

Please avoid using "=" in your variable names

### REQUIRED

`WHITELISTED_REPOSITORIES="user/repo user/repo2 ..."`
A **Space-separated** list of dockerhub repositories that will trigger the webhook.

### OPTIONAL

`ON_WEBHOOK_COMMAND="docker compose up -d"`

Place the command you wish to run when the webhook triggers here.
Defaults to `docker compose up -d`

`WEBHOOK_URL="/webhook"`

If you would like your webhook URL to be something other than "/webhook". Defaults to "/webhook"
This MUST start with a "/" Ex. /dockerhook/testhook

`WEB_PORT=8080`

The port you would like to run the web listener on. Defaults to 8080.

#### Note:

I largely made this for my own shoddy CD pipeline. It runs a raw exec.Command() on your system, so be careful using this!
