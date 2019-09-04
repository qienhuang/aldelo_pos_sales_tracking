package access

import (
	"database/sql"
	"fmt"

	"sales_monitor/config"
	"sales_monitor/helper"
)

type ItemSale struct {
	MenuItemID   sql.NullInt64 //{Value, Valid}
	MenuItemText sql.NullString
	Quantity     sql.NullInt64
}

type MenuItem struct {
	MenuItemID          sql.NullInt64
	MenuItemText        sql.NullString
	SecLangMenuItemText sql.NullString
}

var (
	OdbcDB                *sql.DB
	ACCESS_CONNECT_STRING string
)

func InitOdbc() {

	if helper.DevMode {
		ACCESS_CONNECT_STRING = config.Config.GetString("ACCESS_DB_CONNECTION_STRING_DEV")
	} else {
		ACCESS_CONNECT_STRING = config.Config.GetString("ACCESS_DB_CONNECTION_STRING_PRODUCTION")
	}

	var err error
	OdbcDB, err = sql.Open("odbc", ACCESS_CONNECT_STRING)
	if err != nil {
		fmt.Println("Connecting Error")
		return
	}

}

func CloseOdbc() {
	//cliean up
	if OdbcDB != nil { 
		OdbcDB.Close()
		fmt.Println("Access db is closed.")
	}
	return
}

//  Each item sales in quantity from morning until now
func GetItemSales() []ItemSale {

	itemSales := []ItemSale{}
	tRes := ItemSale{}

	sql := fmt.Sprintf(`
		SELECT OT.MenuItemID, MI.MenuItemText as Menu_Item,SUM(OT.Quantity)  as Total_Sale
		FROM (
		OrderTransactions  OT 
		INNER JOIN MenuItems MI 
		ON OT.MenuItemID = MI.MenuItemID
		)
		INNER JOIN OrderHeaders OH
		ON OT.OrderID = OH.OrderID
		where DateValue(OH.OrderDateTime) = #%v#
		Group by    MI.MenuItemText,   OT.MenuItemID
		`, helper.GetToday())

	rows, err := OdbcDB.Query(sql)

	if err != nil {
		fmt.Println(err)
		return itemSales
	}

	if rows != nil {
		defer rows.Close()
	}

	for rows.Next() {
		rows.Scan(
			&tRes.MenuItemID,
			&tRes.MenuItemText,
			&tRes.Quantity,
		)

		itemSales = append(itemSales, tRes)

	}
	if helper.DevMode {
		fmt.Println("Item sales for today: ")
		fmt.Println("*************************************")
		for _, row := range itemSales {
			fmt.Println("ID#", row.MenuItemID.Int64, "	", row.MenuItemText.String, "	=", row.Quantity.Int64)
		}
	}

	return itemSales

}

// Each item sales in quantity per hour
func GetItemSalesByHour(day string, hour string) []ItemSale {

	itemSales := []ItemSale{}
	tRes := ItemSale{}

	sql := fmt.Sprintf(`
		SELECT	OT.MenuItemID, MI.MenuItemText AS Menu_Item, SUM(OT.Quantity) AS Total_Sale
		FROM  ((OrderTransactions OT INNER JOIN
		MenuItems MI ON OT.MenuItemID = MI.MenuItemID) INNER JOIN
		OrderHeaders OH ON OT.OrderID = OH.OrderID)
		WHERE (OH.OrderDateTime >= #%v %v:00#) AND (OH.OrderDateTime <= #%v %v:59:59#)
		GROUP BY MI.MenuItemText, OT.MenuItemID
		`, day, hour, day, hour)

	rows, err := OdbcDB.Query(sql)

	if err != nil {
		fmt.Println(err)
		return itemSales
	}

	if rows != nil {
		defer rows.Close()
	}

	for rows.Next() {
		rows.Scan(
			&tRes.MenuItemID,
			&tRes.MenuItemText,
			&tRes.Quantity,
		)

		itemSales = append(itemSales, tRes)

	}
	if helper.DevMode {
		fmt.Println("Item sales in specified hour: ")
		fmt.Println("*************************************")
		for _, row := range itemSales {
			fmt.Println("ID#", row.MenuItemID.Int64, "	", row.MenuItemText.String, "	=", row.Quantity.Int64)
		}
	}
	return itemSales
}

// total order from morning util now
func GetOrderAmount(day string, hour string) sql.NullInt64 {

	var totalOrders sql.NullInt64

	sql := fmt.Sprintf(`
		SELECT        COUNT(OrderID) AS TotalOrders
		FROM            OrderHeaders
		WHERE (OrderDateTime >= #%v %v:00#) AND (OrderDateTime <= #%v %v:59:59#)
		`, day, hour, day, hour)

	rows, err := OdbcDB.Query(sql)

	if err != nil {
		fmt.Println(err)
		return totalOrders
	}

	if rows != nil {
		defer rows.Close()
	}

	rows.Next()
	rows.Scan(&totalOrders) 

	fmt.Println("Total orders at ", hour, " clock:", totalOrders.Int64)
	return totalOrders

}

// Total revenue from morning util now
func GetTotalRevenue() sql.NullFloat64 {

	var TotalRevenue sql.NullFloat64
	_sql := fmt.Sprintf(`
		SELECT	SUM(AmountPaid) AS Total_Revenue
		FROM	OrderPayments
		WHERE	(PaymentDateTime > #%v#) AND (PaymentDateTime < #%v 11:59:59 PM#)
		`, helper.GetToday(), helper.GetToday())

	row := OdbcDB.QueryRow(_sql)

	row.Scan(&TotalRevenue)

	fmt.Println("Total Revenue : ", TotalRevenue.Float64)
	return TotalRevenue
}

// Update Menu item names
func GetAllMenuItems() []MenuItem {
	menuItems := []MenuItem{}

	item := MenuItem{}

	sql := fmt.Sprintf(`SELECT MenuItemID,MenuItemText,SecLangMenuItemText FROM MenuItems;`)

	rows, err := OdbcDB.Query(sql)

	if err != nil {
		fmt.Println(err)
		return menuItems
	}

	if rows != nil {
		defer rows.Close()
	}

	for rows.Next() {
		rows.Scan(
			&item.MenuItemID,
			&item.MenuItemText,
			&item.SecLangMenuItemText,
		)
		menuItems = append(menuItems, item)

	}

	if helper.DevMode {
		fmt.Println("All menu items: ")
		fmt.Println("*************************************")

		for _, item := range menuItems {
			fmt.Println(item.MenuItemID.Int64, "	", item.MenuItemText.String, "	", item.SecLangMenuItemText.String)
		}
	}
	return menuItems
}
