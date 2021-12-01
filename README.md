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
    hosts := []console.Host{
        {
            Host:          "192.168.1.10",
            Port:          22,
            TransportType: console.TransportSSH,
            Account: console.Account{
                Username: "admin",
                Password: "pass",
            },
        },
        {
            Host:          "192.168.1.20",
            Port:          22,
            TransportType: console.TransportTELNET,
            Account: console.Account{
                Username: "admin",
                Password: "pass",
            },
        },
    }

    for _, h := range hosts {
        c := console.New()
        if err := c.Open(&h); err != nil {
            log.Fatal(err)
        }
        defer c.Close()

        if err := c.Run("term le 0"); err != nil {
            log.Fatal(err)
        }

        out, err := c.Execute("sh ver")
        if err != nil {
            log.Fatal(err)
        }

        fmt.Println(out)

        c.Send("q")
    }

}
```
