# jgivc/console

[![Go Report Card](https://goreportcard.com/badge/github.com/jgivc/console)](https://goreportcard.com/report/github.com/jgivc/console)

This package provides a uniform interface for interacting with network hardware via telnet/ssh
This package uses part of [reiver/go-telnet package](https://github.com/reiver/go-telnet) for handle telnet connection.

## Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/jgivc/console"
)

func main() {
    defaultAccount := console.Account{Username: "user", Password: "password"}
    hostFactory := console.NewHostFactory(defaultAccount)

    for _, uri := range []string{"host1", "ssh://host2", "user1:password1@host3"} {
        host, err := hostFactory.GetHost(uri)
        if err != nil {
            log.Println("cannot convert uri to host")
            continue
        }

        if err := workOnHost(host); err != nil {
            log.Print(err)
        }

    }
}

func workOnHost(host *console.Host) (err error) {
    c := console.New()
    if err = c.Open(host); err != nil {
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
