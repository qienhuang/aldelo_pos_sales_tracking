package scheduler

import (
	"fmt"
	"log"

	"sales_monitor/access"
	"sales_monitor/helper"
	"sales_monitor/mysql"

	"gopkg.in/robfig/cron.v3"
)

var (
	C *cron.Cron
)

func Start() {
	C = cron.New()
	//Run every 5 minutes
	C.AddFunc("4/5 * * * *", func() {
		log.Println("Schdueled task is running:")

		thisHour := helper.GetThisHour()
		lastHour := helper.GetLastHour()
		today := helper.GetToday()
		yesterday := helper.GetYesterday()

		// 12AM
		if thisHour == "0" {
			// Save menu item sales for last hour and this hour
			mysql.InsertMenuItemSale(access.GetItemSalesByHour(yesterday, lastHour), yesterday, lastHour)
			mysql.InsertMenuItemSale(access.GetItemSalesByHour(today, thisHour), today, thisHour)

			// Save number of orders for last hour and this hour
			mysql.InserOrderAmount(access.GetOrderAmount(yesterday, lastHour), yesterday, lastHour)
			mysql.InserOrderAmount(access.GetOrderAmount(today, thisHour), today, thisHour)

		} else {
			// Save menu item sales for last hour and this hour
			mysql.InsertMenuItemSale(access.GetItemSalesByHour(today, lastHour), today, lastHour)
			mysql.InsertMenuItemSale(access.GetItemSalesByHour(today, thisHour), today, thisHour)

			// Save number of orders for last hour and this hour
			mysql.InserOrderAmount(access.GetOrderAmount(today, lastHour), today, lastHour)
			mysql.InserOrderAmount(access.GetOrderAmount(today, thisHour), today, thisHour)

		}

		mysql.InsertSalesRevenue(access.GetTotalRevenue())

	})

	C.AddFunc("* 12,18 * * *", func() {
		// Only update MenuItems every day at 12:00 and 18:00
		mysql.InserAllMenuItems(access.GetAllMenuItems(), 1)

	})
	C.Start()
	fmt.Println("Scheduler is started.")
}
