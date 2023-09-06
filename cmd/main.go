package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"

	"github.com/jgivc/console"
	"github.com/jgivc/console/config"
	"github.com/jgivc/console/host"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

const (
	defaultLogDirPerm = 0755
)

type commands []string

func (c *commands) String() string {
	return "commands to execute"
}

func (c *commands) Set(value string) error {
	*c = append(*c, value)
	return nil
}

func main() {
	configFileName := flag.String("c", "", "Path to config")
	workers := flag.Int("w", 1, "Concurrency count")
	logDir := flag.String("l", "", "Log dir. Store output to logdir/host_addtess.log")
	ack := flag.Bool("a", false, "Ack username, password")
	ackEnable := flag.Bool("A", false, "Ack enable password. Works together with -a")
	dummy := flag.String("d", "", "Dummy transport config")
	printConfig := flag.Bool("p", false, "Print default console config and exit.")

	var commandFlags commands
	flag.Var(&commandFlags, "e", "Commands to execute. Multiple values accepted.")

	flag.Parse()

	if *printConfig {
		cfg := config.DefaultConsoleConfig()
		out, err := yaml.Marshal(cfg)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprint(os.Stdout, string(out))
		os.Exit(0)
	}

	flagsCfg := &config.FromFlags{
		Commands:    commandFlags,
		DummyConfig: *dummy,
	}

	if *ack {
		account, err := getAccount(*ackEnable)
		if err != nil {
			log.Fatal(err)
		}
		flagsCfg.Account = account
	}

	cfg, err := config.Load(*configFileName, flagsCfg)
	if err != nil {
		log.Fatal(err)
	}

	if *logDir != "" {
		if errCreateDir := os.Mkdir(*logDir, defaultLogDirPerm); errCreateDir != nil {
			panic(errCreateDir)
		}
	}

	var wg sync.WaitGroup
	logger := log.New(os.Stdout, "", log.LstdFlags)
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan *config.HostConfig)

	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer close(c)

	go func() {
		<-c
		cancel()
	}()

	for i := 0; i < *workers; i++ {
		w := worker{
			logDir: *logDir,
			logger: logger,
		}

		wg.Add(1)
		go w.Run(ctx, &wg, ch)
	}

	wg.Add(1)
	go func() {
		defer func() {
			close(ch)
			wg.Done()
		}()

		for i := range cfg.Hosts {
			select {
			case <-ctx.Done():
				return
			case ch <- &cfg.Hosts[i]:
			}
		}
	}()

	wg.Wait()
}

func getAccount(ackEnable bool) (*host.Account, error) {
	var account host.Account

	reader := bufio.NewReader(os.Stdin)

	envUser := os.Getenv("USER")
	if envUser != "" {
		fmt.Printf("Username (default: %s): ", envUser) //nolint: forbidigo //User input
	} else {
		fmt.Print("Username: ") //nolint: forbidigo //User input
	}

	enteredUsername, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("cannot read username: %w", err)
	}

	if strings.TrimSpace(enteredUsername) == "" {
		enteredUsername = envUser
	}

	if enteredUsername == "" {
		return nil, fmt.Errorf("no username defined")
	}

	account.Username = strings.TrimSpace(enteredUsername)

	fmt.Print("Enter Password: ") //nolint: forbidigo //User input
	bytePassword, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		return nil, fmt.Errorf("cannot read password: %w", err)
	}

	account.Password = strings.TrimSpace(string(bytePassword))

	if ackEnable {
		fmt.Print("Enter enable Password: ") //nolint: forbidigo //User input
		byteEnablePassword, err2 := term.ReadPassword(syscall.Stdin)
		if err2 != nil {
			return nil, fmt.Errorf("cannot read password: %w", err2)
		}

		account.EnablePassword = strings.TrimSpace(string(byteEnablePassword))
	}

	return &account, nil
}

type worker struct {
	logDir string
	logger *log.Logger
}

func (w *worker) Run(ctx context.Context, wg *sync.WaitGroup, ch chan *config.HostConfig) {
	defer wg.Done()
	for cfg := range ch {
		w.run(ctx, cfg)
	}
}

func (w *worker) run(ctx context.Context, cfg *config.HostConfig) {
	w.logger.Printf("Get host: %s", cfg.Host.Host)

	var (
		outFile *os.File
		err     error
	)

	if w.logDir == "" {
		outFile = os.Stdout
	} else {
		outFile, err = os.OpenFile(path.Join(w.logDir, fmt.Sprintf("%s.log", cfg.Host.Host)),
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			w.logger.Printf("Cannot openlog file for host %s, error: %v", cfg.Host.Host, err)
			return
		}
		defer outFile.Close()
	}

	c := console.NewWithConfig(&cfg.ConsoleConfig)
	if err2 := c.Open(ctx, &cfg.Host); err2 != nil {
		w.logger.Printf("Cannot open console to host %s, error: %v", cfg.Host.Host, err2)
		return
	}
	defer c.Close()

	for _, cmd := range cfg.InitialCommands {
		if err2 := c.Run(cmd); err2 != nil {
			w.logger.Printf("Cannot run command: %s on host %s, error: %v", cfg.Host.Host, cmd, err2)
		}
	}

	for _, cmd := range cfg.Commands {
		out, err3 := c.Execute(cmd)
		if err3 != nil {
			w.logger.Printf("Cannot execute command: %s to host %s, error: %v", cfg.Host.Host, cmd, err3)
			continue
		}

		_, errWrite := outFile.WriteString(out)
		if errWrite != nil {
			w.logger.Printf("Cannot write result to host out file. Host: %s, error: %v", cfg.Host.Host, errWrite)
		}
	}

	c.Sendln(cfg.ExitCommand)
}
