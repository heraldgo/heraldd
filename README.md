# Herald Daemon

[![Go Report Card](https://goreportcard.com/badge/github.com/heraldgo/heraldd)](https://goreportcard.com/report/github.com/heraldgo/heraldd)

Herald Daemon is designed as a lightweight task dispatcher.
It could be used to arrange the server maintenance jobs,
which is able to control the tasks for a single server and remote servers.
Job scripts could be located on any git server.

The Herald Daemon builds the [Herald](https://github.com/heraldgo/herald)
workflow from a YAML configuration file. It also provides
some common Herald components.


## Contents

* [Installation](#installation)
* [Configuration](#configuration)
  * [Log to file](#log-to-file)
  * [Structure for trigger, selector and executor section](#structure-for-trigger-selector-and-executor-section)
  * [Preset section](#preset-section)
  * [Router section](#router-section)
* [Examples](#examples)
  * [Run periodically](#run-periodically)
  * [Run command with cron](#run-command-with-cron)
  * [Run with preset param](#run-with-preset-param)
  * [Run with complex workflow](#run-with-complex-workflow)
* [Trigger](#trigger)
  * [exe_done](#exe_done)
  * [tick](#tick)
  * [cron](#cron)
  * [http](#http)
* [Selector](#selector)
  * [all](#all)
  * [match_map](#match_map)
  * [except_map](#except_map)
  * [external](#external)
* [Executor](#executor)
  * [none](#none)
  * [print](#print)
  * [local](#local)
  * [http_remote](#http_remote)
* [Extend components with plugin](#extend-components-with-plugin)
  * [Optional function](#optional-function)


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
6. preset
7. router


### Log to file

If no `output` specified, the log will go to stderr.

```yaml
log:
  level: DEBUG
  output: /var/log/heraldd/heraldd.log
```


### Structure for trigger, selector and executor section

The configuration structure for trigger, selector and executor are quite
similar. Take selector as an example:

```yaml
selector:
  selector1_name:
    type: selector_type

    param1: value1
    param2: value2

  selector2_name:
    type: selector_type
```

The name for the same component must be unique.
Each component should have its type,
which could be omitted if it is the same as the component name.
All remaining parameters are passed to component.
The parameters vary among different component types.

It is not necessary to write configuration for all components. They
could be specified by the type name in router directly, which will use
their default parameters.


### Preset section

The `preset` section includes a map which includes some common params
which could be used in router.

```yaml
preset:
  preset1:
    key1: value1
    key2: value2
  preset2:
    key3: value3
```


### Router section

```yaml
router:
  router1_name:
    trigger: trigger1_name
    selector: selector1_name
    task:
      task1_name: executor1_name
      task2_name:
        executor: executor2_name
        select_param:
          preset: [preset1, preset2]
          key1: value1
        job_param: preset2
    select_param:
      key2: value2
    job_param:
      preset: preset3
      key3: value3
  router2_name:
    trigger: trigger1_name
    selector: selector2_name
    task:
      task3_name: executor2_name
    job_param: [preset1, preset3, preset4]
```

`select_param` will be passed to the selector and `job_param` will be
added to the execution param of executor. Both params could be specified
in the route level and the task level. Each param could specify the
preset value which could be a string or a list of strings.

The final param for each task is the combination of preset and
inline params from both router level and task level.
In case there are conflicts, the priority is:

```
Task inline > Task preset[0] > Task preset[1] > ... > Router inline > Router preset[0] > Router preset[1] > ...
```

The name `preset` is reserved and should not be used as param name.
If no task specific params are needed, the executor name could be used
directly as a string.
If inline params are absent, the preset name could be specified as
a string or slice of strings directly.


## Examples


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
    task:
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
    task:
      run_local: local_command
    job_param:
      cmd: uptime
  print_result:
    trigger: exe_done
    selector: match_map
    task:
      print_result: print
    select_param:
      match_key: router
      match_value: uptime_wednesday_morning
    job_param:
      print_key: trigger_param/result
```

`exe_done` trigger could be used to get the job execution result.
The `match_map` selector here only accepts previous job which
comes from `uptime_wednesday_morning` router.


### Run with preset param

You can put common params in the `preset` section, which could
be referenced in router.

```yaml
trigger:
  every5s:
    type: tick
    interval: 5

executor:
  local_command:
    type: local
    work_dir: /var/lib/heraldd/work

preset:
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
    task:
      hostname:
        executor: local_command
        job_param:
          preset: hostname
      df:
        executor: local_command
        job_param:
          preset: df
      uptime:
        executor: local_command
        job_param:
          preset: uptime
  print_result:
    trigger: exe_done
    selector: match_map
    task:
      print_result: print
    select_param:
      match_key: router
      match_value: run_every5s
    job_param:
      print_key: [trigger_param/task, trigger_param/result/exit_code, trigger_param/result/output]
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

preset:
  common_script_repo:
    git_repo: https://github.com/heraldgo/herald-script
    git_username: user
    git_password: pass

router:
  print_param_every2s:
    trigger: every2s
    selector: all
    task:
      print_param: print
  ls_wednesday_morning:
    trigger: wednesday_morning
    selector: all
    task:
      run_local_ls: local_command
    job_param:
      cmd: ls
      arg: /
  run_every_evening:
    trigger: every_evening
    selector: all
    task:
      hostname:
        executor: remote_command
        job_param:
          cmd: hostname
      df:
        executor: local_command
        job_param:
          cmd: df
          arg: [-hT]
      uptime:
        executor: local_command
        job_param:
          cmd: uptime
      print_param: print
  doit_remote_every_evening:
    trigger: every_evening
    selector: all
    task:
      doit_locally: local_command
      doit_remotely: remote_command
    job_param:
      preset: common_script_repo
      cmd: doit.sh
```


## Trigger

Trigger defines when the job workflow should start.
Herald Daemon provides the following triggers.


### exe_done

This is an internal trigger name, not a type.
It is automatically activated after any job is done.
Do **NOT** define a trigger with the same name.
You can use `exe_done` trigger directly in the router.

The "trigger param" for `exe_done` is the result of last job, which
looks like:

```json
{
  "id": "F60CFC6A-2FDE-248D-6C35-C3EFD484014F",
  "trigger_id": "A8D875BC-5875-3BA7-EECB-F829A341F78E",
  "router": "router_name",
  "trigger": "trigger_name",
  "selector": "selector_name",
  "task": "task_name",
  "executor": "executor_name",
  "trigger_param": {},
  "select_param": {},
  "job_param": {},

  "success": true,
  "error": "",
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
    task:
      hostname: local_command
    job_param:
      cmd: hostname
  print_result:
    trigger: exe_done
    selector: match_map
    task:
      print_result: print
    select_param:
      match_key: router
      match_value: run_every5s
```

With `exe_done` you can also build a task chain with proper selector.

```yaml
router:
  step1:
    trigger: every_morning
    selector: all
    task:
      step1: print
  step2:
    trigger: exe_done
    selector: match_map
    task:
      step2: print
    select_param:
      match_key: router
      match_value: step1
  step3:
    trigger: exe_done
    selector: match_map
    task:
      step3: print
    select_param:
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
    task:
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
    task:
      print_result: print
    select_param:
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
    task:
      print_result: print
    select_param:
      except_key: router
      except_value: print_result
```

If `except_value` is absent, it will fail when `except_key` exists.


### external

In case the selection is complex and no internal selector is available,
`external` selector provides a way to write your own program as selector.
It will call an external program which sets json format
of "trigger param" and "job param" as environment variables.
The default variable names are `HERALD_TRIGGER_PARAM` and
`HERALD_SELECT_PARAM`, which could be configured in selector options
`trigger_param_env` and `select_param_env` individually.
The selector will pass if the exit code is 0.

```yaml
selector:
  xxx:
    type: external
    program: /selector/xxx.py
    #trigger_param_env: HERALD_TRIGGER_PARAM
    #select_param_env: HERALD_SELECT_PARAM

router:
  print_result:
    trigger: exe_done
    selector: xxx
    task:
      print_result: print
    select_param:
      key: value
```

This is an example of program written in python:

```python
#!/usr/bin/env python

import sys
import json

trigger_param = json.loads(os.environ['HERALD_TRIGGER_PARAM'])
select_param = json.loads(os.environ['HERALD_SELECT_PARAM'])

if trigger_param.get('key') != select_param.get('key'):
    sys.exit(1)  # Do not pass

sys.exit(0)
```


## Executor

This is what the execution param looks like.

```json
{
  "id": "F60CFC6A-2FDE-248D-6C35-C3EFD484014F",
  "trigger_id": "A8D875BC-5875-3BA7-EECB-F829A341F78E",
  "router": "router_name",
  "trigger": "trigger_name",
  "selector": "selector_name",
  "task": "task_name",
  "executor": "executor_name",
  "trigger_param": {},
  "select_param": {},
  "job_param": {}
}
```

`trigger_param` comes from the trigger. `job_param` is combined
from router and task.


### none

Do nothing. Could be used for debug purpose.


### print

Print the job param to log.

```yaml
router:
  print:
    trigger: ttt
    selector: all
    task:
      print_it: print
    job_param:
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
    task:
      run_cmd: local_command
    job_param:
      cmd: uptime
  run_git:
    trigger: ttt
    selector: all
    task:
      run_git: local_command
    job_param:
      git_repo: https://github.com/heraldgo/herald-script.git
      cmd: run/doit.sh
  check_env:
    trigger: ttt
    selector: all
    task:
      run_git: local_command
    job_param:
      cmd: printenv
      arg: ['TEST_SET_ENV', 'TEST_ANOTHER_ENV']
      env:
        TEST_SET_ENV: 'Herald daemon'
        TEST_ANOTHER_ENV: 'This is another env'
  print_result:
    trigger: exe_done
    selector: match_map
    task:
      print_result: print
    select_param:
      match_key: executor
      match_value: local_command
    job_param:
      print_key: [trigger_param/result]
```

Here are the params used by `local` executor.

* `cmd`. The command to be executed. If `git_repo` is set,
  the `cmd` will be relative to the `git_repo`.
* `arg`. Argument(s) which will be passed to the command.
  The `arg` could be a list of strings.
  If it is a string, it will be used as a single argument.
  Do **NOT** write multiple arguments in the same string.
* `env`. This is a map which will be set as environment
  variables for the command.
* `param_env_name`. The name of the environment variable
  which includes `json` format of all execution parameters.
  The default name is `HERALD_EXECUTE_PARAM`.
* `ignore_param_env`. Set to `true` if you do not want to
  set the `HERALD_EXECUTE_PARAM` environment variable.
* `background`. If set to `true`, the command will run
  in background and return immediately.
  You are not able to get the result of the command anymore.
* `git_repo`. The executor will try to load the git repo
  and run `cmd` from it.
  **Only use `git_repo` which you can trust.**
* `git_username`. The username for authentication.
* `git_password`. The password for authentication.
  If `git_password` is not empty, `git_ssh_key` and `git_ssh_key_file`
  will be ignored.
* `git_ssh_key`. The string of ssh key if you are using ssh protocol.
  if `git_ssh_key` is not empty, `git_ssh_key_file` will be ignored.
* `git_ssh_key_file`. The file path of ssh key.
  If it is empty, it will try to find one under `~/.ssh`.
* `git_ssh_key_password`. Provide the password in case
  the ssh key is encrypted.
* `git_branch`. Remote branch for the git repo.

All the trigger and job params could be found in
`HERALD_EXECUTE_PARAM` environment variable.

The multiline `git_ssh_key` could be written like this:

```yaml
router:
  router1:
    trigger: trigger1
    selector: all
    task:
      run_cmd: local_command
    job_param:
      cmd: test/run_script.sh
      git_repo: git@github.com:heraldd/herald-script.git
      git_ssh_key: |
        -----BEGIN RSA PRIVATE KEY-----
        MIIEpAIBAAKCAQEAn5aGCdbBNBOawhMJ2/SoKVoAAL5tRN5MJzrJGob09p6MC/dc
        AZoMH5YOQEpoBaZOg8smh9GlqGSE2LKmWreNrqZk/+0w5XUnNhcQ/I6MY2u2l5fb
        iYx5FgclExsNH+Y3EUyv1LVfRuIRRLg7WH1snoKYsmteAVwVIZGtFgBs4AzhTsn9
        k3mW9FZF90DuEbsrRu6up8SiDobF4t2IDZKU/wIDAQABAoIBAQCVh1YkFeKFRvE0
        ct5EB+Mgi8GA8Ow1IQy9nSkc/+K6ySdzdtvwbER7u/+yYYVB9eePOWPq0pajRzvq
        Rsn0KhRI1oPAAKBV/wU0ezxhR7dm2GAHfjQnl0VFTICCfFA52V0zimUdqquRIPUJ
        bubkD58K2S6DuVOJ1DB3VNRI4qvCGu8D1N+iS9o0l07NtKzSITFNlRQvdk5OSPZJ
        fWtxuU3SfXAw8Y2cPM1j1SECgYEAovzjGRhZv+lXh51YBU+7XBM7iUMswslzNiNh
        eV36dfDzhwqQKF3AGtX0nMWkSruS8s8AzqBuNfmF5/O7H0nrvIShMA73gJ8eHLPe
        8aFzMaPX4uvt7ZAFJDfXU1Eqbb4T/W0oBtQJopI9n5r7Ry9/eghCqZBMabRpsVvp
        9v9dVW0CgYAl87ZqIwr/JwiqnDuxvS2E+3hYK3pWjFAl9/OKi1JbS94/ZmyLO3l5
        +bnhEXkB/wUHF59yPVu4M9JOg67ugNmW8gRAQ7SbRoGtdfUdC2zV4JstGqf+PJlM
        5xiJ2wxJsq+hct1OpbzjmzXBRrDUOevASvCsTGdfhG+Neqnm1IJUNA==
        -----END RSA PRIVATE KEY-----
```

The RFC4716-format ssh key is not supported currently. If you are encounting
problems with encrypted key, try to generate key in old PEM format with
`-m PEM`:

```shell
$ ssh-keygen -m PEM
```

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
    task:
      run_cmd: remote_command
    job_param:
      cmd: hostname
  run_git:
    trigger: ttt
    selector: all
    task:
      run_git: remote_command
    job_param:
      git_repo: https://github.com/heraldgo/herald-script.git
      cmd: run/doit.sh
  print_result:
    trigger: exe_done
    selector: match_map
    task:
      print_result: print
    select_param:
      match_key: executor
      match_value: remote_command
    job_param:
      print_key: trigger_param/result
```

The job param for `http_remote` is exactly the same as `local`, so you
can run the same task with both `local` and `http_remote`.

> If `git_ssh_key_file` is specified, it will try to load the ssh key file
> from the Herald Runner server, not the Herald Daemon server.
> If you would like to use the ssh key from the Herald Daemon server,
> write the content in `git_ssh_key`.

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
