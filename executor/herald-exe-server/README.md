# Herald Exe Server

The herald exe server is a http server which is used to
cooperate with the Herald Daemon executor `http_remote`.
The job from `http_remote` will be sent to the herald exe server
and `http_remote` get result from the http response.


## Installation

First install [Go](https://golang.org/) and setup the workspace,
then use the following command to install herald daemon.

```shell
$ go get -u github.com/heraldgo/heraldd/executor/herald-exe-server
```


## Configuration

```yaml
log_level: INFO
log_output: /var/log/herald-exe-server/herald-exe-server.log

host: 127.0.0.1
port: 8124
#unix_socket: /var/run/herald-exe-server/herald-exe-server.sock

secret: the_secret_should_be_strong_enough

work_dir: /var/lib/herald-exe-server/work
```

The secret must be exactly the same as the one in `http_remote`
executor.
The `work_dir` is similar to the `local` executor.


## Run the server

Run the herald exe server:

```shell
$ herald-exe-server -config config.yml
```

Press `Ctrl+C` to exit.


## HTTPS with nginx

If you would like to use https to secure the job request, you may use
reverse proxy of nginx and setup certificates there.
