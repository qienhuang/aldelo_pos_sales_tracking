package main

/*
Programming by Kevin Huang
08/2019
kevin11206@gmail.com
web-based Aldelo POS sales monitoring
Platform: Windows x86/x64
*/

import (
	"fmt"
	_ "log"
	_ "time"

	"sales_monitor/access"
	_ "sales_monitor/config"
	"sales_monitor/helper"
	"sales_monitor/mysql"
	"sales_monitor/scheduler"
	"sales_monitor/webserver"

	_ "github.com/alexbrainman/odbc"
	_ "github.com/go-sql-driver/mysql"
)

func init() {

}

func main() {
	fmt.Println("Sales tracker is starting ~~~~~~~~~~~~~~~~~~~~~~")

	// catchs ctl+c SIGINT/SIGTERM signals, cleanup connections before exit
	_, done := helper.ExitHandler()

	access.InitOdbc()
	mysql.InitMySql()

	scheduler.Start()
	webserver.Start()

	// append cleanup functions to ExitHandler
	helper.OnClosing(access.CloseOdbc)
	helper.OnClosing(mysql.CloseMySql)
	helper.OnClosing(webserver.Shutdown)

	// blocks and wait for the ExitHandler
	<-done

}
