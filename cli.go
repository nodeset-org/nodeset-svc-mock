package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/nodeset-org/nodeset-svc-mock/server"
	"github.com/urfave/cli/v2"
)

const (
	Version string = "0.1.0"
)

// Run
func main() {
	// Initialise application
	app := cli.NewApp()

	// Set application info
	app.Name = "nodeset-svc-mock"
	app.Usage = "Mock of the nodeset.io service, useful for testing"
	app.Version = Version
	app.Authors = []*cli.Author{
		{
			Name:  "Nodeset",
			Email: "info@nodeset.io",
		},
	}
	app.Copyright = "(C) 2024 NodeSet LLC"

	ipFlag := &cli.StringFlag{
		Name:    "ip",
		Aliases: []string{"i"},
		Usage:   "The IP address to bind the API server to",
		Value:   "127.0.0.1",
	}
	portFlag := &cli.UintFlag{
		Name:    "port",
		Aliases: []string{"p"},
		Usage:   "The port to bind the API server to",
		Value:   49537,
	}

	app.Flags = []cli.Flag{
		ipFlag,
		portFlag,
	}
	app.Action = func(c *cli.Context) error {
		logger := slog.Default()

		// Create the server
		var err error
		ip := c.String(ipFlag.Name)
		port := uint16(c.Uint(portFlag.Name))
		server, err := server.NewNodeSetMockServer(logger, ip, port)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating server: %v", err)
			os.Exit(1)
		}

		// Start it
		wg := &sync.WaitGroup{}
		err = server.Start(wg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error starting server: %v", err)
			os.Exit(1)
		}
		port = server.GetPort()

		// Handle process closures
		termListener := make(chan os.Signal, 1)
		signal.Notify(termListener, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-termListener
			fmt.Println("Shutting down...")
			server.Stop()
		}()

		// Run the daemon until closed
		logger.Info(fmt.Sprintf("Started nodeset.io mock server on %s:%d", ip, port))
		wg.Wait()
		fmt.Println("Server stopped.")
		return nil
	}

	// Run application
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
