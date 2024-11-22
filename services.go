package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

// Service represents the configuration for each service
type Service struct {
	Path        string
	Port        string
	Description string
	Cmd         *exec.Cmd // Keep track of the command to enable graceful shutdown
	BinaryName  string    // Name of the compiled binary
}

// services is a slice containing all the services to run
var services = []Service{
	{Path: "../wf-ceo/", Port: "8081", Description: "Workflow General Manager", BinaryName: "wf-ceo"},
	// {Path: "../wf-gen/", Port: "8082", Description: "Workflow Generator Service", BinaryName: "wf-gen"},
	// {Path: "../wf-dba/", Port: "8083", Description: "Workflow Database Service", BinaryName: "wf-dba"},
	// {Path: "../wf-mgr/", Port: "8084", Description: "Workflow Manager Service", BinaryName: "wf-mgr"},
}

// startService builds and starts a service in a new goroutine
func startService(ctx context.Context, wg *sync.WaitGroup, service *Service, moduleCache, goPath string) error {
	// Build the service binary
	buildCmd := exec.Command("go", "build", "-o", service.BinaryName)
	buildCmd.Dir = service.Path

	buildCmd.Env = append(os.Environ(),
		"GOCACHE=/tmp/gocache",
		"GOMODCACHE="+moduleCache,
		"GOPATH="+goPath,
	)

	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("failed to build service %s: %w", service.Path, err)
	}

	// Prepare to remove the binary after execution
	binaryPath := service.Path + "/" + service.BinaryName
	defer os.Remove(binaryPath)

	cmd := exec.CommandContext(ctx, "./"+service.BinaryName)
	cmd.Dir = service.Path

	// Set environment variables
	cmd.Env = append(os.Environ(),
		"PORT="+service.Port,
		"GOCACHE=/tmp/gocache",
		"GOMODCACHE="+moduleCache,
		"GOPATH="+goPath,
	)

	// Redirect stdout and stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set process group ID so that we can send signals to the process group
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Start the service
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start service %s: %w", service.Path, err)
	}

	service.Cmd = cmd // Keep track of the command

	log.Printf("Started %s on port %s", service.Path, service.Port)

	// Wait for the service to finish
	go func() {
		defer wg.Done()
		if err := cmd.Wait(); err != nil && ctx.Err() == nil {
			// Only log error if context was not canceled
			log.Printf("Service %s exited with error: %v", service.Path, err)
		} else {
			log.Printf("Service %s stopped", service.Path)
		}
	}()

	return nil
}
