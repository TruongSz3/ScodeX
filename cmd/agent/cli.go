package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/sz3/scodex/internal/client"
)

const defaultDaemonURL = "http://127.0.0.1:7777"

func run(args []string) int {
	if len(args) == 0 {
		printHelp()
		return 0
	}

	switch args[0] {
	case "doctor":
		return runDoctor(args[1:])
	case "ping":
		return runPing(args[1:])
	case "help", "-h", "--help":
		printHelp()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		printHelp()
		return 2
	}
}

func runDoctor(args []string) int {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	daemonURL := fs.String("daemon-url", defaultDaemonURL, "daemon HTTP URL")
	authToken := fs.String("auth-token", "", "daemon auth token (optional; defaults from AGENT_AUTH_TOKEN)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cli := client.New(client.WithBaseURL(*daemonURL), client.WithAuthToken(*authToken))
	health, err := cli.CheckHealth(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "doctor failed: %v\n", err)
		return 1
	}

	fmt.Printf("daemon healthy: %s\n", health.Status)
	return 0
}

func runPing(args []string) int {
	fs := flag.NewFlagSet("ping", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	daemonURL := fs.String("daemon-url", defaultDaemonURL, "daemon HTTP URL")
	authToken := fs.String("auth-token", "", "daemon auth token (optional; defaults from AGENT_AUTH_TOKEN)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cli := client.New(client.WithBaseURL(*daemonURL), client.WithAuthToken(*authToken))
	if err := cli.PingHandshake(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ping failed: %v\n", err)
		return 1
	}

	fmt.Println("handshake ok")
	return 0
}

func printHelp() {
	fmt.Println("agent - local CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  agent <command> [flags]")
	fmt.Println()
	fmt.Println("Available Commands:")
	fmt.Println("  doctor   Check daemon health endpoint")
	fmt.Println("  ping     Run initialize/initialized handshake")
	fmt.Println("  help     Show help")
}
