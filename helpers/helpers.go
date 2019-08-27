package helpers

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	configmanager "github.com/project-eria/eria-base/config-manager"
	"github.com/project-eria/logger"
	"github.com/project-eria/xaal-go/engine"
)

var _xAALConfigFile = "xaal.json"

// InitEngine load the xAAL config file and start the engine
func InitEngine() {
	var config engine.XaalConfig

	configManagerXAAL, err := configmanager.Init(_xAALConfigFile, &config)
	if err != nil {
		logger.Module("engine").WithError(err).WithField("filename", _xAALConfigFile).Fatal()
	}

	if err := configManagerXAAL.Load(); err != nil {
		logger.Module("engine").WithError(err).Fatal()
	}
	engine.Init(config)
}

// AddShowVersion add a flag to display app version
func AddShowVersion(version string) {
	_showVersion := flag.Bool("v", false, "Display the version")
	if !flag.Parsed() {
		flag.Parse()
	}

	// Show version (-v)
	if *_showVersion {
		fmt.Printf("%s\n", version)
		os.Exit(0)
	}
}

// WaitForExit Wait for any signal and runs all the defer
func WaitForExit() {
	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	// Block until keyboard interrupt is received.
	<-c
	runtime.Goexit()
}
