# Herald Daemon

Herald Daemon is designed as a lightweight task dispatcher.
It could be used to arrange the server maintenance jobs,
which is able to control the tasks for a single server and remote servers.
Job scripts could be located on any git server.

The Herald Daemon builds the [Herald](https://github.com/heraldgo/herald)
workflow from a YAML configuration file. It also provides
some common Herald components.


## Installation

Download binary file from the
[release page](https://github.com/heraldgo/heraldd/releases).

If you would like to build from source,
first install [Go](https://golang.org/) and setup the workspace,
then use the following command to install Herald Daemon.

```shell
$ go get -u github.com/heraldgo/heraldd
```

Write a configuration file and run the Herald Daemon:

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
    match_key: router
    match_value: uptime_wednesday_morning
    print_key: trigger_param/result
```

`exe_done` trigger could be used to get the job execution result.
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
    match_key: router
    match_value: run_every5s
    print_key: [trigger_param/result/exit_code, trigger_param/result/output]
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

Trigger defines when the job workflow should start.
Herald Daemon provides the following triggers.


### exe_done

This is an internal trigger which is activated after any job is done.
Do **NOT** define a trigger with the same name.

The "trigger param" for `exe_done` is the result of last job, which
looks like:

```json
{
  "id": "F60CFC6A-2FDE-248D-6C35-C3EFD484014F",
  "trigger_id": "A8D875BC-5875-3BA7-EECB-F829A341F78E",
  "router": "router_name",
  "trigger": "trigger_name",
  "selector": "selector_name",
  "job": "job_name",
  "trigger_param": {},
  "job_param": {},
  "result": {},
}
```

The "trigger param" above is just the job common
information plus the job execution `result`.
The content of `result` depends on the execution,
and varies among executors.

`exe_done` can be used to check the status of a job exeuction.

```yaml
router:
  run_every5s:
    trigger: every5s
    selector: all
    job:
      hostname: local_command
    cmd: hostname
  print_result:
    trigger: exe_done
    selector: match_map
    job:
      print_result: print
    match_key: router
    match_value: run_every5s
```

With `exe_done` you can also build a job chain with proper selector.

```yaml
router:
  step1:
    trigger: every_morning
    selector: all
    job:
      step1: print
  step2:
    trigger: exe_done
    selector: match_map
    job:
      step2: print
    match_key: router
    match_value: step1
  step3:
    trigger: exe_done
    selector: match_map
    job:
      step3: print
    match_key: router
    match_value: step2
```

Do **NOT** use `all` selector with `exe_done` trigger, which will lead to
a dead loop.


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
    port: 8123
```

You must `POST` with a json body to the server to activate the trigger:

```shell
$ curl -i -H "Content-Type:application/json" -X POST -d '{"clean":"old_files"}' localhost:8123
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

The selector check the "trigger param" and "job param" to determine
whether or not to proceed with the job execution.

### all

Pass all the situation.

```yaml
router:
  print_param_every2s:
    trigger: every2s
    selector: all
    job:
      print_param: print
```


### match_map

Only pass when specified key and value match in trigger param.
Nested keys are seperated by "/".

```yaml
router:
  print_result:
    trigger: exe_done
    selector: match_map
    job:
      print_result: print
    match_key: router
    match_value: uptime_wednesday_morning
```

If `match_value` is absent, it will only check the existence of
the `match_key`.


### except_map

`except_map` is the opposite of `match_map`. It will **NOT** pass
if specified key and value are matched.

```yaml
router:
  print_result:
    trigger: exe_done
    selector: except_map
    job:
      print_result: print
    except_key: router
    except_value: print_result
```

If `except_value` is absent, it will fail when `except_key` exists.


### external

In case the selection is complex and no internal selector is available,
`external` selector provides a way to write your own program as selector.
It will call an external program which accepts json format
of "trigger param" and "job param" as arguments.
The selector will pass if the exit code is 0.

```yaml
selector:
  xxx:
    type: external
    program: /selector/xxx.py

router:
  print_result:
    trigger: exe_done
    selector: xxx
    job:
      print_result: print
    key: value
```

This is an example of program written in python:

```python
#!/usr/bin/env python

import sys
import json

trigger_param = json.loads(sys.argv[1])
job_param = json.loads(sys.argv[2])

if trigger_param.get('key') != job_param.get('key'):
    sys.exit(1)  # Do not pass

sys.exit(0)
```


## Executor

This is what the job param looks like.

```json
{
  "id": "F60CFC6A-2FDE-248D-6C35-C3EFD484014F",
  "trigger_id": "A8D875BC-5875-3BA7-EECB-F829A341F78E",
  "router": "router_name",
  "trigger": "trigger_name",
  "selector": "selector_name",
  "job": "job_name",
  "trigger_param": {},
  "job_param": {}
}
```

`trigger_param` comes from the trigger. `job_param` is the combination
of "router param" and "job specific param".


### none

Do nothing. Could be used for debug purpose.


### print

Print the job param to log.

```yaml
router:
  print:
    trigger: ttt
    selector: all
    job:
      print_it: print
    print_key: [trigger, trigger_param/result]
```

If the option `print_key` is set as job param,
the `print` executor will only print specified keys.


### local

Run command on the local server.
Make sure `work_dir` is set properly, which will keep the git repo
and used as the command current work directory.

```yaml
executor:
  local_command:
    type: local
    work_dir: /var/lib/heraldd/work

router:
  run_cmd:
    trigger: ttt
    selector: all
    job:
      run_cmd: local_command
    cmd: uptime
  run_git:
    trigger: ttt
    selector: all
    job:
      run_git: local_command
    script_repo: https://github.com/heraldgo/herald-script.git
    cmd: run/doit.sh
    param_arg: true
  print_result:
    trigger: exe_done
    selector: match_map
    job:
      print_result: print
    match_key: executor
    match_value: local_command
    print_key: [trigger_param/result]
```

If `param_arg` is set to true, the job param will be passed as the
last command argument in json form.

If `script_repo` is set, `local` executor will try to load it as
a git repo and then run the `cmd` from it.

Only use `script_repo` which you can trust.

The result of the `local` executor is like:

```json
{
  "exit_code": 0,
  "output": "",
  "file": {
    "file1": "/full/path/of/file1.dat",
    "file2": "/full/path/of/file2.dat"
  },
  "key1": "value1",
  "key2": "value2"
}
```

If the standard output of the command could be converted to json,
it will be merged into the result, or it will be directly put in `output`.

If you would like to get the result, add a router triggered by
`exe_done` and check the `trigger_param`.


### http_remote

`http_remote` provide the way to execute job on a remote server.
It must be used together with the
[Herald Runner](https://github.com/heraldgo/herald-runner).

`data_dir` is used to keep output files from the remote execution.
`secret` must be exactly the same with Herald Runner or
the request will be rejected.
`secret` is used for SHA256 HMAC signature of the request body.

```yaml
executor:
  remote_command:
    type: http_remote
    host: https://example.com/
    secret: yyyyyyyyyyyyyyyy
    data_dir: /var/lib/heraldd/data

router:
  run_cmd:
    trigger: ttt
    selector: all
    job:
      run_cmd: remote_command
    cmd: hostname
  run_git:
    trigger: ttt
    selector: all
    job:
      run_git: remote_command
    script_repo: https://github.com/heraldgo/herald-script.git
    cmd: run/doit.sh
  print_result:
    trigger: exe_done
    selector: match_map
    job:
      print_result: print
    match_key: executor
    match_value: remote_command
    print_key: trigger_param/result
```

The job param for `http_remote` is exactly the same as `local`, so you
can run the same job with both `local` and `http_remote`.

If the job need output files, the output json of the command
must include `file` part. These files will be validated by SHA256
checksum.

```json
{
  "file": {
    "file1": "/full/path/of/file1.dat",
    "file2": "/full/path/of/file2.dat"
  },
  "key1": "value1",
  "key2": "value2"
}
```

Then these files will be transferred back to the Herald Daemon server
and kept in `data_dir`.
The final result will also include these files with local path.

```json
{
  "file": {
    "file1": "/data_dir/job_id/file1/file1.dat",
    "file2": "/data_dir/job_id/file2/file2.dat"
  },
  "key1": "value1",
  "key2": "value2"
}
```


## Extend components with plugin

Herald Daemon has provided some internal triggers, selectors and executors.
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
> between plugin and Herald Daemon.

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

Herald Daemon will try to find a type of component first in the order of
plugin list and then from the internal ones.
Once the specified component is found, it will stop further searching.


### Optional function

There is one optional method for each components, `SetLogger`.

If you would like to share the logger with Herald Daemon, you can
implement this function:

```go
func (c *component) SetLogger(logger interface{}) {
	c.logger = logger
}
```

The `logger` could be considered as a `Herald.Logger` interface.
