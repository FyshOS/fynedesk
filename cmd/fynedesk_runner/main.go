package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

const runCmd = "fynedesk"

func main() {
	_ = os.Remove(logPath()) // remove old logs
	_ = os.Remove(runnerLogPath())
	log.SetOutput(openRunnerLogWriter())

	for {
		logFile := logPath()
		if _, err := os.Stat(logFile); err == nil {
			crashFile := crashLogPath()
			err = os.Rename(logFile, crashFile)
			if err != nil {
				log.Println("Could not save crash file", crashFile)
			}
		}

		exe := exec.Command(runCmd)
		exe.Env = append(os.Environ(), "FYNE_DESK_RUNNER=1")
		// logger will be closed at the end of this for loop
		logger := openLogWriter()
		exe.Stdout, exe.Stderr = logger, logger
		err := exe.Run()
		if err == nil {
			return
		}

		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			log.Println("Could not execute", runCmd, "command")
			return
		}

		if exit, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			status := exit.ExitStatus()
			if status == 0 {
				log.Println("Exiting Error 0")
				return
			} else if status == 512 { // X server unavailable
				log.Println("X server went away")
				return
			} else if status == 2 {
				log.Println("Failed to connect to X, retrying")
			} else {
				log.Println("Restart from status", status)
			}
		}

		// close before starting next run
		_ = logger.Close()
	}
}
