# Kill on File

*This program will daemonize itself. Refer to the log file to see status.*

This will periodically poll a given file and when it exists [or ceases to exist] send a specified signal to a specified PID(s).

A configurable delay is available between the file's presence changing and the signal being sent.

Optionally, after the first signal has been sent and a configurable delay, a `KILL` signal can be sent to the same PID(s)

### Command line flags

*Flag values can also be set using environment variables. The environment variable name is the uppercase flag name prefixed with `KILL_ON_FILE_`.*

```
  -delay int
    	Wait at least this number of seconds after file is detected before sending signal [KILL_ON_FILE_DELAY]
  -killfile string
    	File path to trigger signal [KILL_ON_FILE_KILLFILE] (default "./kill-on-file.trigger")
  -killgrace int
    	Send a KILL signal this number of seconds after initial signal. Zero disables this [KILL_ON_FILE_KILLGRACE]
  -logfile string
    	Log file path [KILL_ON_FILE_LOGFILE] (default "./kill-on-file.log")
  -notexist
    	Send signal if file is not present instead of when file is present [KILL_ON_FILE_NOTEXIST]
  -pidfile string
    	PID file path [KILL_ON_FILE_PIDFILE] (default "./kill-on-file.pid")
  -pollseconds int
    	File polling frequency in seconds [KILL_ON_FILE_POLLSECONDS] (default 5)
  -signal string
    	Name of the signal to send e.g TERM | USR1 | QUIT | KILL | ... [KILL_ON_FILE_SIGNAL] (default "TERM")

Provide one or more PIDs as arguments after the config flags

```

### Example

This will wait until file `./trigger-file` exists, then pause for at least 2 seconds before sending a `QUIT` signal to PIDs `112` & `446`. After sending the `QUIT` signals, it will wait `6` seconds, and then send a `KILL` signal to the same two PIDs (`112` & `446`)

```
./kill-on-file  -delay 2  -killfile ./trigger-file  -killgrace 6  -signal QUIT  112 446
```

### Known considerations

 * When using `-killgrace` to send `KILL` after the initial signal, there is no verification that the PID is actually the same process as before. If a process exists quickly after the initial signal, and the OS happens to re-use the same PID for some new process, then that new process will be inadvertently killed.