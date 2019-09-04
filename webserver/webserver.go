package webserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"sales_monitor/helper"
	"sales_monitor/mysql"

	"github.com/gin-gonic/gin"
)

/*

	InitServer( )
	   |
	   V
	InitRouter( )

*/
var (
	Router  *gin.Engine
	httpSvr *http.Server
)

func Start() {
	Router = gin.Default()
	initRouter(Router)

	httpSvr = &http.Server{
		Addr:           ":80",
		Handler:        Router,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// web server run in background
	go func() {
		httpSvr.ListenAndServe()
	}()

}

func initRouter(Router *gin.Engine) {
	// Load templates
	Router.LoadHTMLGlob("webserver/templates/*")
	// Specifies static foler
	Router.Static("static", "webserver/static")
	// favicon.ico
	Router.StaticFile("/favicon.ico", "webserver/static/favicon.ico")

	// Authentication
	/*
		authorized := Router.Group("/", gin.BasicAuth(gin.Accounts{
			"admin": "password",
		}))
	*/

	Router.GET("/", func(c *gin.Context) {

		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Sales tracking",
		})
	})

	// Return data to jquery
	Router.POST("/api/get_menuitem_sales", func(c *gin.Context) {
		revenue := mysql.GetSalesRevenue(helper.GetToday(), 1)
		odersInEachHour := mysql.GetOrderAmountForEveryHour(helper.GetToday(), 1)
		menuItems := mysql.GetMenuItemSalesByDay(helper.GetToday(), 1)

		c.JSON(200, gin.H{
			"revenue":        revenue.Float64,
			"orders_by_hour": odersInEachHour,
			"menuItemSales":  menuItems,
		})

	})

}

// Shutdown server before the app close, invoked by helper.OnClosing()
func Shutdown() {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if httpSvr != nil {
		httpSvr.Shutdown(ctx)
		fmt.Println("Http server is closed.")
	}
}
