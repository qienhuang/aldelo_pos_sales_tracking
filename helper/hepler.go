package helper

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sales_monitor/config"
)

var (
	DevMode = config.Config.GetBool("DevMode")
	handler OnClosingHandler
)

type OnClosingHandler struct {
	callList []func()
}

func OnClosing(fn func()) {
	handler.callList = append(handler.callList, fn)
}

func GetToday() string {
	if DevMode {
		return config.Config.GetString("Today") //for testing
	} else {
		return time.Now().Format("01/02/2006")
	}
}

func GetYesterday() string {
	if DevMode {
		return config.Config.GetString("Yesterday") //for testing
	} else {
		return time.Now().Add((time.Hour * -24)).Format("01/02/2006")
	}
}

func GetThisHour() string {
	if DevMode {
		return config.Config.GetString("ThisHour")
	} else {
		return time.Now().Format("15")
	}
}

func GetLastHour() string {
	if DevMode {
		return config.Config.GetString("LastHour")
	} else {
		return time.Now().Add(time.Hour * -1).Format("15")
	}
}

func ExitHandler() (error, chan bool) {
	done := make(chan bool)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			fmt.Println("***Catched an ", sig, " signal. The app is about to exit! ***")

			// call cleanup functions before close the program
			for _, caller := range handler.callList {
				caller()
			}

			// Wait 3 seconds for cleanup complete
			time.Sleep(time.Second * 6)
			done <- true
			os.Exit(0)
			fmt.Println("done.")
		}
	}()

	return nil, done
}
