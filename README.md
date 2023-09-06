# Console

[![Go Report Card](https://goreportcard.com/badge/github.com/jgivc/console)](https://goreportcard.com/report/github.com/jgivc/console)

This package provides a generic interface for accessing network equipment via telnet/ssh. It is used, for example, to collect some information from the equipment, for example, a running configuration or to execute some commands. It can be used as a library and as an application.

## Usage

### Sample config file

```yaml
# Default initial command: term le 0
# Default exit command: q
commands:
  - sh run
default_account:
  username: admin
  password: password
hosts:
  - 10.0.0.1
  - 10.0.0.2

```

Hosts may be defined as:

```
ssh://user:pass:enablepass@host:port
telnet://user:pass:enablepass@host:port
user:pass:enablepass@host:port
user:pass:enablepass@host
user:pass@host
host
```

Sample config can be found in [example](example/) folder.

### Running

```
  -A		Ack enable password. Work with -a
  -a		Ack username, password
  -c string Path to config
  -d string Dummy transport config
  -e string Commands to execute. Multiple values accepted.
  -l string	Log dir. Store output to logdir/host_address.log
  -p		Print default console config and exit.
  -w int	Concurrency count (default 1)

```

Example usage for store host running configs:

config.yml

```yaml
hosts:
  - 10.0.0.1
  - 10.0.0.2
  # ...
  - 10.0.0.50

```
and run

```shell
./console -a -c config.yml -l out -w 5 -e "sh run" 
```


### Check config

You can check your configuration with dummy transport. With it you can describe the received data and timeout using a xml file. Sample config can be seen in [example](example/) folder. Specify your configuration file with -d flag.


## Usage as library

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jgivc/console"
	"github.com/jgivc/console/host"
	"github.com/jgivc/console/util"
)

func main() {
	defaultAccount := host.Account{Username: "admin", Password: "password"}
	hostFactory := util.NewHostFactory(defaultAccount)

	for _, uri := range []string{"host1", "ssh://host2", "user:password@host3"} {
		host, err := hostFactory.GetHost(uri)
		if err != nil {
			log.Println("cannot convert uri to host")
			continue
		}

		if err := workOnHost(context.Background(), host); err != nil {
			log.Print(err)
		}
	}
}

func workOnHost(ctx context.Context, host *host.Host) (err error) {
	c := console.New()
	if err = c.Open(ctx, host); err != nil {
		return
	}
	defer c.Close()

	fmt.Printf("Connect to host %s\n", host.GetHostPort())

	if err = c.Run("term le 0"); err != nil {
		return
	}

	out, err := c.Execute("sh ver")
	if err != nil {
		return
	}

	fmt.Println(out)

	c.Sendln("q")

	return
}
```