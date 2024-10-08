package main

import (
	"context"
	"fenfa/internal/config"
	"fenfa/internal/link"
	"fenfa/internal/store"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/sevlyar/go-daemon"
)

const (
	LogFileName      = "fenfa.log"
	PIDFileName      = "fenfa.pid"
	CommandStart     = "start"
	CommandStop      = "stop"
	CommandForceQuit = "force-quit"
	CommandLink      = "link"
	CommandList      = "list"
	CommandUnban     = "unban"
)

var (
	requests        int
	mu              sync.Mutex
	windowDuration  = time.Minute
	signalFlag      = new(string)
	httpServer      *http.Server
	binaryDirectory string
	cntxt           *daemon.Context
)

func main() {

	binaryPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable path: %v", err)
	}
	binaryDirectory = filepath.Dir(binaryPath)

	var logPath = filepath.Join(binaryDirectory, LogFileName)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	if len(os.Args) < 2 {
		fmt.Println("No command provided. Usage: fenfa [start|stop|force-quit|list [entries|ip_attempts]|link /path/to/file]")
		os.Exit(1)
	}

	config.Initialize(binaryDirectory)
	store.Initialize()

	command := os.Args[1]

	switch command {
	case CommandStart, CommandStop, CommandForceQuit:
		initializeDaemonContext()
	}

	switch command {
	case CommandStart:
		startServer(cntxt)
	case CommandStop:
		*signalFlag = CommandStop
		sendFlag(cntxt)
	case CommandForceQuit:
		*signalFlag = CommandForceQuit
		sendFlag(cntxt)
	case CommandLink:
		if len(os.Args) < 3 {
			fmt.Println("No path provided. Usage: fenfa link /path/to/file]")
			os.Exit(1)
		}
		link.GenerateFileLink(os.Args[2])
	case CommandList:
		if len(os.Args) < 3 {
			fmt.Println("No table provided. Usage: fenfa link [entries|ip_entries]")
			os.Exit(1)
		}
		store.List(os.Args[2])
	case CommandUnban:
		if len(os.Args) < 3 {
			fmt.Println("No IP provided. Usage: fenfa unban [IP Address]")
			os.Exit(1)
		}
		store.ResetFailedAttempts(os.Args[2])
	default:
		fmt.Println("Invalid command. Usage: fenfa [start|stop|list [entries|ip_attempts]|link /path/to/file]")
	}

	switch command {
	case CommandStart, CommandStop, CommandForceQuit:
		initializeDaemonContext()
	}
}

func startServer(cntxt *daemon.Context) {
	d, err := cntxt.Search()
	if err == nil && d != nil {
		log.Println("Server is already running.")
		return
	}

	d, err = cntxt.Reborn()
	if err != nil {
		log.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	log.Println("Started Daemon")
	defer cntxt.Release()
	go resetRateLimit()
	httpServer = &http.Server{
		Addr: fmt.Sprintf(":%d", config.Port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			if !rateLimit() {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			link.FileHandler(w, r)
		}),
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	err = daemon.ServeSignals()
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}

	log.Println("Daemon successfully terminated.")
}

func initializeDaemonContext() {
	daemon.AddCommand(daemon.StringFlag(signalFlag, CommandStop), syscall.SIGQUIT, signalHandler)
	daemon.AddCommand(daemon.StringFlag(signalFlag, CommandForceQuit), syscall.SIGTERM, signalHandler)

	cntxt = &daemon.Context{
		PidFileName: PIDFileName,
		PidFilePerm: 0644,
		LogFileName: LogFileName,
		LogFilePerm: 0640,
		WorkDir:     binaryDirectory, //Daemon process changes to the binary directory.
		Umask:       027,
	}
}

func sendFlag(cntxt *daemon.Context) {
	d, err := cntxt.Search()
	if err != nil {
		log.Fatalf("Unable to send signal to the daemon: %s", err.Error())
	}
	err = daemon.SendCommands(d)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func signalHandler(sig os.Signal) error {
	log.Printf("Received signal: %v, shutting down...", sig)
	if sig == syscall.SIGQUIT {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
	} else {
		if err := httpServer.Close(); err != nil {
			log.Printf("HTTP server Close: %v", err)
		}
	}

	return daemon.ErrStop
}

func rateLimit() bool {
	mu.Lock()
	defer mu.Unlock()

	if requests >= config.RateLimit {
		return false
	}
	requests++
	return true
}

func resetRateLimit() {
	for {
		time.Sleep(windowDuration)
		mu.Lock()
		requests = 0
		mu.Unlock()
	}
}
