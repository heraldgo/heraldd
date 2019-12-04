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

Write a configuration file and run the herald daemon:

```shell
$ heraldd -config config.yml
```

Press `Ctrl+C` to exit.


## Configuration

The workflow is defined in a single [YAML](https://yaml.org/) file.

The configuration consists of following sections:

1. log
2. plugin
3. trigger
4. selector
5. executor
6. job
7. router


### Log to file

If no `output` specified, the log will go to stderr.

```yaml
log:
  level: DEBUG
  output: /var/log/heraldd/heraldd.log
```


### Run periodically

This is an example which print the param every 2 seconds.

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


### Run command with cron

This is an example which run `uptime` command on wednesday morning.

```yaml
trigger:
  wednesday_morning:
    type: cron
    cron: '30 6 * * 3'

executor:
  local_command:
    type: local
    work_dir: /var/lib/heraldd/work

router:
  uptime_wednesday_morning:
    trigger: wednesday_morning
    selector: all
    job:
      run_local: local_command
    cmd: uptime
  print_result:
    trigger: exe_done
    selector: match_map
    job:
      print_result: print
    match_key: info/router
    match_value: uptime_wednesday_morning
    print_key: [trigger_param/exit_code, trigger_param/output]
```

If you would like to check the status of a job execution,
you can add a router with `exe_done` trigger which is activated
after any job is done. You also need a selector to filter the result
you would like to see, or all the job results will be printed,
including `print_result` itself, where it will lead to a dead loop.
The `match_map` selector here only accepts previous job which
comes from `uptime_wednesday_morning` router.


### Run with job specific param

You can put job specific param in the `job` section, which will
overwrite param in router.

```yaml
trigger:
  every5s:
    type: tick
    interval: 5

executor:
  local_command:
    type: local
    work_dir: /var/lib/heraldd/work

job:
  hostname:
    cmd: hostname
  df:
    cmd: df
    arg: [-hT]
  uptime:
    cmd: uptime

router:
  run_every5s:
    trigger: every5s
    selector: all
    job:
      hostname: local_command
      df: local_command
      uptime: local_command
  print_result:
    trigger: exe_done
    selector: match_map
    job:
      print_result: print
    match_key: info/router
    match_value: run_every5s
    print_key: [trigger_param/exit_code, trigger_param/output]
```


### Run with complex workflow

You can combine different triggers, executors and selectors
in the routers.

```yaml
trigger:
  every2s:
    type: tick
    interval: 2
  wednesday_morning:
    type: cron
    cron: '30 6 * * 3'
  every_evening:
    type: cron
    cron: '0 18 * * *'

executor:
  local_command:
    type: local
    work_dir: /var/lib/heraldd/work
  remote_command:
    type: http_remote
    host: https://example.com/
    secret: yyyyyyyyyyyyyyyy
    data_dir: /var/lib/heraldd/data

job:
  hostname:
    cmd: hostname
  df:
    cmd: df
    arg: [-hT]
  uptime:
    cmd: uptime

router:
  print_param_every2s:
    trigger: every2s
    selector: all
    job:
      print_param: print
  uptime_wednesday_morning:
    trigger: wednesday_morning
    selector: all
    job:
      run_local_ls: local_command
    cmd: ls
    arg: /
  run_every_evening:
    trigger: every_evening
    selector: all
    job:
      hostname: remote_command
      df: local_command
      df: remote_command
      uptime: local_command
      print_param: print
    cmd: hostname
  doit_remote_every_evening:
    trigger: every_evening
    selector: all
    job:
      doit: remote_command
    script_repo: https://github.com/heraldgo/herald-script
    cmd: doit.sh
```


## Trigger

Herald daemon provides the following triggers.

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

There is no authority control for this trigger,
so it is not a good idea to open it globally.


## Selector


## Extend components with plugin

Herald daemon has provided some internal triggers, selectors and executors.
If you are not satisfied with them, you can also define your own ones
in the form of plugin to meet your requirements.

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
> So it is OK that the plugin does not import `herald` package,
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
