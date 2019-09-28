package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"golang.org/x/sys/unix"
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

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nProvide one or more PIDs as arguments after the config flags\n")
	}

	pidFileName := flag.String("pidfile", "/var/run/kill-on-file.pid", "PID file path")
	logFileName := flag.String("logfile", "/var/log/kill-on-file.log", "Log file path")
	killFileName := flag.String("killfile", "/var/log/kill-on-file.log", "Log file path")
	actionOnNotExist := flag.Bool("notexist", false, "Send signal if file is not present instead of when file is present")
	pollSeconds := flag.Int("pollseconds", 5, "File polling frequency in seconds")
	signalName := flag.String("signal", "TERM", "Name of the signal to send e.g TERM | USR1 | QUIT | KILL | ...")
	killGrace := flag.Int("killgrace", 0, "Send a KILL signal this number of seconds after initial signal. Zero disables this")
	killPidsString := flag.Args()
	envy.Parse("KILL_ON_FILE")
	flag.Parse()

	var killPids = []int{}
	for _, killPidArg := range killPidsString {
		killPidInt, err := strconv.Atoi(killPidArg)
		if err != nil {
			continue
		}
		killPids = append(killPids, killPidInt)
	}

	if len(killPids) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	signalNum := unix.SignalNum(fmt.Sprintf("SIG%s", strings.ToUpper(signalName)))

	cntxt := &daemon.Context{
		PidFileName: pidFileName,
		PidFilePerm: 0644,
		LogFileName: logFileName,
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{fmt.Sprintf("[kill-on-file: %s]", killFile)},
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	log.Printf("== Will send %s if exists=%v for file %s (polling every %ds)", strings.ToUpper(signalName), !actionOnNotExist, killFileName, pollSeconds)

	for {

		if existsBool := fileExists(killFileName); existsBool != actionOnNotExist {
			for killPid := range killPids {
				if foundProcess, err := os.FindProcess(killPid); err == nil {
					sigErr := foundProcess.Signal(signalNum)
					log.Printf("Sent %s to %d (err: %v)", strings.ToUpper(signalName), killPid, sigErr)
				}
			}
			break
		}
		time.Sleep(time.Duration(pollSeconds) * time.Second)

	}
	if killGrace > 0 {
		time.Sleep(time.Duration(killGrace) * time.Second)
		for killPid := range killPids {
			if foundProcess, err := os.FindProcess(killPid); err == nil {
				sigErr := foundProcess.Signal(unix.SignalNum("SIGKILL"))
				log.Printf("Sent %s to %d (err: %v)", "KILL", killPid, sigErr)
			}
		}

	}

}
