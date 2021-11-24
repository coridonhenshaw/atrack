package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type ScriptEntry struct {
	Exec    string
	Timeout time.Duration
}

type TriggeredScript struct {
	Trigger chan bool
	Script  []ScriptEntry
}

type RuntimeConfigStruct struct {
	Mutex                 sync.RWMutex
	ConfigurationFilename string
	PIDFilename           string

	IPv4Script TriggeredScript
	IPv6Script TriggeredScript

	ListenAddr string
	UserID     string
	Password   string

	TLSFullchain  string
	TLSPrivateKey string

	IPv4 net.IP
	IPv6 net.IP
}

func main() {
	fmt.Println("Atrack version 0.")
	RuntimeConfig, err := LoadConfig("atrack.xml")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Loaded configuration file", RuntimeConfig.ConfigurationFilename)

	go ExecuteQueue(RuntimeConfig, &RuntimeConfig.IPv4Script)
	go ExecuteQueue(RuntimeConfig, &RuntimeConfig.IPv6Script)

	writePidFile(RuntimeConfig.PIDFilename)

	err = HTTPServer(RuntimeConfig)
}

func writePidFile(pidFile string) error {
	// Read in the pid file as a slice of bytes.
	if piddata, err := ioutil.ReadFile(pidFile); err == nil {
		// Convert the file contents to an integer.
		if pid, err := strconv.Atoi(string(piddata)); err == nil {
			// Look for the pid in the process list.
			if process, err := os.FindProcess(pid); err == nil {
				// Send the process a signal zero kill.
				if err := process.Signal(syscall.Signal(0)); err == nil {
					// We only get an error if the pid isn't running, or it's not ours.
					return fmt.Errorf("pid already running: %d", pid)
				}
			}
		}
	}
	// If we get here, then the pidfile didn't exist,
	// or the pid in it doesn't belong to the user running this app.
	return ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0664)
}
