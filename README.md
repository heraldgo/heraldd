# Herald Daemon

The herald daemon builds the [herald](https://github.com/heraldgo/herald)
workflow from a YAML configuration file. Herald daemon also provides
some useful herald components.


## Installation

First install [Go](https://golang.org/) and setup the workspace,
then use the following command to install herald daemon.

```shell
$ go get -u github.com/heraldgo/heraldd
```

Run the herald daemon:

```shell
$ heraldd -config config.yml
```


## Configuration

The workflow is defined in a single [YAML](https://yaml.org/) file.

```yaml
trigger:
  every2s:
    type: tick
    interval: 2

router:
  print_param_every2s:
    trigger: every2s
    selector: all
    job:
      print_param: print
```


## Trigger

### tick

A trigger activated periodically. The unit for the interval is second.

```yaml
trigger:
  every2s:
    type: tick
    interval: 2
```


### cron

"cron" builds a trigger with cron syntax.
It uses the [cron](https://github.com/robfig/cron) library.

```yaml
trigger:
  cron:
    cron: '30 6 * * *'
```

You can add second field if option `with_seconds` is true.

```yaml
trigger:
  cron_every2s:
    type: cron
    cron: '*/2 * * * * *'
    with_seconds: true
```


### http

"http" is trigger which will create a http server.
The trigger will be activated when it receives proper http request.

```yaml
trigger:
  manual:
    type: http
    host: 127.0.0.1
    port: 8181
```

You must `POST` with a json body to the server to activate the trigger:

```shell
$ curl -i -H "Content-Type:application/json" -X POST -d '{"clean":"old_files"}' localhost:8181
```

The json body will be parsed as the "trigger param".

This trigger is suitable for doing some manual actions.

"http" trigger can also listen on unix socket, which could use
nginx as the reverse proxy.

```yaml
trigger:
  http:
    unix_socket: /var/run/heraldd/http.sock
```


## Extend components with plugin

Herald daemon has provided some common triggers, selectors and executors.
You can also define your own ones in the form of plugin
to meet your requirements.

The extended components should be implemented as a
[Go plugin](https://golang.org/pkg/plugin/)
which is built with:

```shell
$ go build --buildmode=plugin
```

Take trigger as example, there must be one function `CreateTrigger` exported:

```go
type triggerExample struct {}

func (tgr *triggerExample) Run(ctx context.Context, sendParam func(map[string]interface{})) {
	...
}

func CreateTrigger(typeName string, param map[string]interface{}) (interface{}, error) {
	if typeName == "trigger_example" {
		return &triggerExample{}, nil
	}
	return nil, fmt.Errorf(`Trigger "%s" is not in this plugin`, typeName)
}
```

> `CreateTrigger` returns `interface{}` instead of `Herald.Trigger`
> in order not to introduce extra import in the plugin.
> So it is possible that the plugin does not import `herald` package,
> which may reduce the possibility of version inconsistency
> between plugin and herald daemon.

Define a type name for each trigger, which will be used in the
configuration.
`CreateTrigger` function should return a trigger instance with
the trigger type and initialize it with the param argument.
If it is not able to create a corresponding trigger,
it should return an error. The returned trigger instance must implement the
[`Herald.Trigger`](https://github.com/heraldgo/herald#trigger) interface.

It is similar to [selector](https://github.com/heraldgo/herald#selector)
and [executor](https://github.com/heraldgo/herald#executor), which need
to export `CreateSelector` or `CreateExecutor` function.
A single plugin could include all or only part of the three kinds of
components (trigger, selector and executor).

The plugin files are specified in the configuration file.
More than one plugins could be added.

```yaml
plugin:
  - /usr/lib/heraldd/plugin/herald-gogshook.so
  - /usr/lib/heraldd/plugin/herald-plugin.so
```

Herald daemon will try to find a type of component first in the order of
plugin list and then from the internal ones.
Once the specified component is found, it will stop further searching.


### Optional function

There is one optional method for each components, `SetLogger`.

If you would like to share the logger with herald daemon, you can
implement this function:

```go
func (c *component) SetLogger(logger interface{}) {
	c.logger = logger
}
```

The `logger` could be considered as a `Herald.Logger` interface.
