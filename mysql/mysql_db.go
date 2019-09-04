package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	"sales_monitor/access"
	"sales_monitor/config"
	"sales_monitor/helper"
)

var (
	MysqlDB                    *sql.DB
	mysqlDbPingOk              = false
	MYSQL_DB_CONNECTION_STRING = ""
)

const ()

type MenuItem struct {
	MenuItemID   sql.NullInt64  `json:"MenuItemID"`
	MenuItemText sql.NullString `json:"MenuITemText`
	Quantity     sql.NullInt64  `json:"Quantity"`
}

func InitMySql() {
	if helper.DevMode {
		MYSQL_DB_CONNECTION_STRING = config.Config.GetString("MYSQL_DB_CONNECTION_STRING_DEV")
	} else {
		MYSQL_DB_CONNECTION_STRING = config.Config.GetString("MYSQL_DB_CONNECTION_STRING_PRODUCTION")
	}

	var err error
	MysqlDB, err = sql.Open("mysql", MYSQL_DB_CONNECTION_STRING)
	err = MysqlDB.Ping()
	if err == nil {
		mysqlDbPingOk = true
	} else {
		log.Println("Ping MySQL host error: ", err)
	}

}

func CloseMySql() {
	if MysqlDB != nil {
		MysqlDB.Close()
		fmt.Println("Myysql db is closed.")
	}
	return
}

// Insert data into mysql table MenuItem_Sale, record hourly
func InsertMenuItemSale(itmeSales []access.ItemSale, day string, hour string) {

	for _, item := range itmeSales {
		// check if the data existing
		_sql := fmt.Sprintf("select MenuItemID from MenuItem_Sale where MenuItemID=%d and OrderTime=STR_TO_DATE( '%v %v:00', '%%m/%%d/%%Y %%H:%%i') and RestaurantID=1",
			item.MenuItemID.Int64, day, hour)
		row := MysqlDB.QueryRow(_sql)
		var count sql.NullInt64
		row.Scan(&count)
		if count.Valid {
			// update row
			_sql = fmt.Sprintf("update  MenuItem_Sale set Quantity=%d  where MenuItemID = %d and OrderTime=STR_TO_DATE( '%v %v:00', '%%m/%%d/%%Y %%H:%%i') and RestaurantID=1",
				item.Quantity.Int64, item.MenuItemID.Int64, day, hour)
			_, err := MysqlDB.Exec(_sql)
			if err != nil {
				log.Println("Error while processing update data on Menuitem_Sale: ", err)
			}

		} else {
			// insert row
			_sql = fmt.Sprintf("insert into MenuItem_Sale(MenuItemID, MenuItemText, Quantity, OrderTime,RestaurantID)  values( %d, '%v', %d, STR_TO_DATE( '%v %v:00', '%%m/%%d/%%Y %%H:%%i'), 1)",
				item.MenuItemID.Int64, "", item.Quantity.Int64, day, hour)
			_, err := MysqlDB.Exec(_sql)
			if err != nil {
				log.Println("Error while processing insert data into Menuitem_Sale: ", err)
			}
		}
	}

}

// Get total menu item sales, by day
func GetMenuItemSalesByDay(date string, restaurantID int) []MenuItem {
	menuItems := []MenuItem{}
	var item MenuItem
	_sql := fmt.Sprintf(`select MenuItem_Sale.MenuItemID,MenuItems.MenuItemText, sum(Quantity) as q from MenuItem_Sale inner join MenuItems 
		where MenuItem_Sale.MenuItemID=MenuItems.MenuItemID and 
		OrderTime between STR_TO_DATE('%v 00:00', '%%m/%%d/%%Y %%H:%%i')
		and STR_TO_DATE('%v 23:59', '%%m/%%d/%%Y %%H:%%i')
		and MenuItem_Sale.RestaurantID=%d
        group by MenuItem_Sale.MenuItemID
        order by q desc`,
		date, date, restaurantID)
	rows, err := MysqlDB.Query(_sql)
	if err != nil {
		fmt.Println("Error while processing GetMenuItemSalesByDay: ", err)
		return menuItems
	}
	if rows != nil {
		defer rows.Close()
	}

	for rows.Next() {
		rows.Scan(
			&item.MenuItemID,
			&item.MenuItemText,
			&item.Quantity,
		)
		temp := item.MenuItemText.String
		if i := strings.IndexByte(temp, '.'); i >= 0 {
			item.MenuItemText.String = temp[i+1:]
		}

		menuItems = append(menuItems, item)
		//fmt.Println(item)

	}
	return menuItems

}

// Insert sales revenue to table Sales_Renvenue, record by day
func InsertSalesRevenue(SaleRevenue sql.NullFloat64) {
	// check if the data existing
	_sql := fmt.Sprintf("select ID from Sales_Revenue where OrderDate=STR_TO_DATE( '%v', '%%m/%%d/%%Y') and RestaurantID=1",
		helper.GetToday())
	row := MysqlDB.QueryRow(_sql)
	var count sql.NullFloat64

	row.Scan(&count)

	//fmt.Println("Record ID:", count.Float64)
	if count.Valid {
		//Update row
		_sql = fmt.Sprintf("update  Sales_Revenue set Revenue=%f  where OrderDate=STR_TO_DATE( '%v', '%%m/%%d/%%Y') and RestaurantID=1",
			SaleRevenue.Float64, helper.GetToday())
		_, err := MysqlDB.Exec(_sql)
		if err != nil {
			log.Println("Error while processing update data on Sales_Revenue: ", err)
		}
	} else {
		//Insert row
		_sql = fmt.Sprintf("insert into Sales_Revenue(Revenue, OrderDate, RestaurantID) values (%f, STR_TO_DATE( '%v', '%%m/%%d/%%Y'), 1)",
			SaleRevenue.Float64, helper.GetToday())
		_, err := MysqlDB.Exec(_sql)
		if err != nil {
			log.Println("Error while processing insert data into Sales_Revenue: ", err)
		}
	}
}

// Fetch sales revenue for today
func GetSalesRevenue(date string, restaurantID int) sql.NullFloat64 {
	_sql := fmt.Sprintf("select Revenue from Sales_Revenue where OrderDate=STR_TO_DATE('%v', '%%m/%%d/%%Y') and RestaurantID=%d ",
		date, restaurantID)
	row := MysqlDB.QueryRow(_sql)
	var revenue sql.NullFloat64
	row.Scan(&revenue)
	fmt.Println("Get today revenue: ", revenue.Float64)
	return revenue
}

// Insert Order Amount into mysql table MenuItem_Sale, record hourly
func InserOrderAmount(orderAmount sql.NullInt64, day string, hour string) {
	// check if the data existing
	_sql := fmt.Sprintf("select OrderAmount from Number_Of_Orders where OrderTime=STR_TO_DATE( '%v %v:00', '%%m/%%d/%%Y %%H:%%i') and RestaurantID=1",
		day, hour)
	row := MysqlDB.QueryRow(_sql)
	var count sql.NullInt64
	row.Scan(&count)
	if count.Valid {
		// update row
		_sql = fmt.Sprintf("update  Number_Of_Orders set OrderAmount=%d  where OrderTime=STR_TO_DATE( '%v %v:00', '%%m/%%d/%%Y %%H:%%i') and RestaurantID=1",
			orderAmount.Int64, day, hour)
		_, err := MysqlDB.Exec(_sql)
		if err != nil {
			log.Println("Error while processing update data on Number_Of_Orders: ", err)
		}

	} else {
		// insert row
		_sql = fmt.Sprintf("insert into Number_Of_Orders(OrderAmount, OrderTime, RestaurantID)  values( %d, STR_TO_DATE( '%v %v:00', '%%m/%%d/%%Y %%H:%%i'), 1)",
			orderAmount.Int64, day, hour)
		_, err := MysqlDB.Exec(_sql)
		if err != nil {
			log.Println("Error while processing insert data into Number_Of_Orders: ", err)
		}
	}
}

// Number of orders by day
func GetOrderAmountByDay(date string, restaurantID int) sql.NullInt64 {
	_sql := fmt.Sprintf(`select sum(OrderAmount) from Number_Of_Orders where 
		OrderTime between STR_TO_DATE('%v 00:00', '%%m/%%d/%%Y %%H:%%i') 
		and STR_TO_DATE('%v 23:59', '%%m/%%d/%%Y %%H:%%i')
		and RestaurantID=%d `,
		date, date, restaurantID)
	row := MysqlDB.QueryRow(_sql)
	var totalOrders sql.NullInt64
	row.Scan(&totalOrders)
	fmt.Println("Get today number of orders by day: ", totalOrders.Int64)
	return totalOrders
}

// Number of orders by hour
func GetOrderAmountByHour(date string, hour string, restaurantID int) sql.NullInt64 {
	_sql := fmt.Sprintf(`select sum(OrderAmount) from Number_Of_Orders where 
		OrderTime between STR_TO_DATE('%v %v:00', '%%m/%%d/%%Y %%H:%%i') 
		and STR_TO_DATE('%v %v:59', '%%m/%%d/%%Y %%H:%%i')
		and RestaurantID=%d `,
		date, hour, date, hour, restaurantID)
	//fmt.Println("_sql:", _sql)
	row := MysqlDB.QueryRow(_sql)
	var totalOrders sql.NullInt64
	row.Scan(&totalOrders)
	//fmt.Println("Get today number of orders by hour: ", totalOrders.Int64)
	return totalOrders
}

func GetOrderAmountForEveryHour(date string, restaurantID int) [24]int64 {
	var OrderInHours [24]int64
	for i := 0; i < 24; i++ {
		OrderInHours[i] = GetOrderAmountByHour(date, strconv.Itoa(i), restaurantID).Int64
	}
	//fmt.Println("OrderInHours", OrderInHours)
	return OrderInHours
}

// Insert/Update MenItem's ID/Text into mysql table MenuItems, update daily
func InserAllMenuItems(menuItems []access.MenuItem, restaurantID int) {
	// check if the data existing
	for _, item := range menuItems {
		_sql := fmt.Sprintf("select MenuItemID from MenuItems where MenuItemID=%d and RestaurantID=%d",
			item.MenuItemID.Int64, restaurantID)
		row := MysqlDB.QueryRow(_sql)
		var itemID sql.NullInt64
		row.Scan(&itemID)
		if itemID.Valid {
			// update row
			_sql = fmt.Sprintf("update  MenuItems set MenuItemText='%v',SecLangMenuItemText='%v'  where MenuItemID=%d and RestaurantID=%d",
				item.MenuItemText.String, item.SecLangMenuItemText.String, item.MenuItemID.Int64, restaurantID)
			_, err := MysqlDB.Exec(_sql)
			if err != nil {
				log.Println("Error while processing update data on MenuItems: ", err)
			}

		} else {
			// insert row
			_sql = fmt.Sprintf("insert into MenuItems(MenuItemID, MenuItemText, SecLangMenuItemText, RestaurantID)  values( %d, '%v','%v', %d)",
				item.MenuItemID.Int64, item.MenuItemText.String, item.SecLangMenuItemText.String, restaurantID)
			_, err := MysqlDB.Exec(_sql)
			if err != nil {
				log.Println("Error while processing insert data into MenuItems: ", err)
			}
		}
	}

}
