package main

import (
	"flag"
	"fmt"
	"golang.org/x/sys/unix"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jamiealquiza/envy"
	"github.com/sevlyar/go-daemon"
)

// This function is not fool proof (permissions or non-file types may cause issues)
// However, for this simple use case, it is enough
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func main() {

	// GET CONFIGURATION
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nProvide one or more PIDs as arguments after the config flags\n")
	}

	pidFileName := flag.String("pidfile", "./kill-on-file.pid", "PID file path")
	logFileName := flag.String("logfile", "./kill-on-file.log", "Log file path")
	killFileName := flag.String("killfile", "./kill-on-file.trigger", "File path to trigger signal")
	actionOnNotExist := flag.Bool("notexist", false, "Send signal if file is not present instead of when file is present")
	pollSeconds := flag.Int("pollseconds", 5, "File polling frequency in seconds")
	signalName := flag.String("signal", "TERM", "Name of the signal to send e.g TERM | USR1 | QUIT | KILL | ...")
	killGrace := flag.Int("killgrace", 0, "Send a KILL signal this number of seconds after initial signal. Zero disables this")
	triggerDelay := flag.Int("delay", 0, "Wait at least this number of seconds after file is detected before sending signal")

	envy.Parse("KILL_ON_FILE")
	flag.Parse()

	killPidsString := flag.Args()

	var killPids = []int{}
	for _, killPidArg := range killPidsString {
		killPidInt, err := strconv.Atoi(killPidArg)
		if err != nil {
			log.Printf("Ignoring: %v", killPidArg)
			continue
		}
		killPids = append(killPids, killPidInt)
	}

	if len(killPids) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	signalNum := unix.SignalNum(fmt.Sprintf("SIG%s", strings.ToUpper(*signalName)))

	// LOG A LINE TO SHOW WE ARE RUNNING
	log.Printf("== Will send %s if exists=%v for file %s to %v (polling every %ds)", strings.ToUpper(*signalName), !*actionOnNotExist, *killFileName, killPids, *pollSeconds)

	// DAEMONIZE
	cntxt := &daemon.Context{
		PidFileName: *pidFileName,
		PidFilePerm: 0644,
		LogFileName: *logFileName,
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        append([]string{"KOF"}, os.Args[1:]...),
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	// POLLING FILE LOOP
	for {
		if existsBool := fileExists(*killFileName); existsBool != *actionOnNotExist {
			log.Printf("File: %s, Exists: %v", *killFileName, existsBool)
			break
		}
		time.Sleep(time.Duration(*pollSeconds) * time.Second)
	}

	// PAUSE FOR TRIGGER DELAY
	time.Sleep(time.Duration(*triggerDelay) * time.Second)

	// SEND INITIAL SIGNALS
	for _, killPid := range killPids {
		if foundProcess, err := os.FindProcess(killPid); err == nil {
			sigErr := foundProcess.Signal(signalNum)
			log.Printf("Sent %d to %d (err: %v)", signalNum, killPid, sigErr)
		}

	}

	// SEND KILL SIGNALS
	if *killGrace > 0 {
		time.Sleep(time.Duration(*killGrace) * time.Second)
		for _, killPid := range killPids {
			if foundProcess, err := os.FindProcess(killPid); err == nil {
				sigKillNum := unix.SignalNum("SIGKILL")
				sigErr := foundProcess.Signal(sigKillNum)
				log.Printf("Sent %d to %d (err: %v)", sigKillNum, killPid, sigErr)
			}
		}

	}

}
