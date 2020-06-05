package evo

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var signals = [...]string{
	1:  "Hangup",
	2:  "Interrupt",
	3:  "Quit",
	4:  "Illegal instruction",
	5:  "Trace/breakpoint trap",
	6:  "Aborted",
	7:  "Bus error",
	8:  "Floating point exception",
	9:  "Killed",
	10: "User defined signal 1",
	11: "Segmentation fault",
	12: "User defined signal 2",
	13: "Broken pipe",
	14: "Alarm clock",
	15: "Terminated",
}
var exit = false

// InterceptOSSignal TODO:Remove
func InterceptOSSignal() {
	return

	Events.On("app.signal", func(signal os.Signal) error {
		if signal.String() == "interrupt" {
			if !exit {
				exit = true
				go func() {
					time.Sleep(5 * time.Second)
					fmt.Println("User did not reply in time. go back to normal life ...")
					exit = false
				}()
				return fmt.Errorf("Are you sure to exit? press ctrl+c again")
			}
		}

		return nil
	})

	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGILL,
		syscall.SIGTRAP,
		syscall.SIGABRT,
		syscall.SIGBUS,
		syscall.SIGFPE,
		syscall.SIGKILL,
		syscall.SIGSEGV,
		syscall.SIGPIPE,
		syscall.SIGALRM,
		syscall.SIGTERM,
	)
	exit_chan := make(chan int)

	go func() {
		for {
			s := <-signal_chan
			fmt.Println(s.String())
			err := Events.Go("app.signal", s)
			if err != nil {
				fmt.Println(err)
				continue
			}
			switch s {
			// kill -SIGINT XXXX or Ctrl+c
			case syscall.SIGINT:
				fmt.Println("Ctrl+C")
				exit_chan <- 0

			default:
				exit_chan <- 1
			}
		}
	}()

	code := <-exit_chan
	Events.Go("app.exit")
	os.Exit(code)

}
