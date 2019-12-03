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

Take trigger as example, there must be one function exported:

```go
func CreateTrigger(typeName string, param map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf(`Trigger "%s" not supported`, typeName)
}
```

> Here returns `interface{}` instead of `Herald.Trigger` in order not to
> introduce extra import in the plugin. So it is possible that the
> plugin does not import `herald` package,
> which may reduce the possibility of version inconsistency
> between plugin and herald daemon.

Define a type name for each trigger, which will be used in the
configuration.
`CreateTrigger` function will try to return a trigger instance with
the trigger type and initialized by the param argument.
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


### Optional functions

There is one optional method for each components, `SetLogger`.

If you would like to share the logger with herald daemon, you can
implement this function:

```go
func (c *component) SetLogger(logger interface{}) {
	c.logger = logger
}
```

The `logger` could be considered as a `Herald.Logger` interface.
