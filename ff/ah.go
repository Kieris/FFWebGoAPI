package ff

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type AHItem struct {
	ItemID     int16
	Stack      byte
	SellerName *string
	Price      int32
	BuyerName  *string
	Sale       int64
	SellDate   int64
	ItemName   string
	Ah         byte
}

func GetAHByName(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", corStr)
	sID := pathParams["sID"]

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		fmt.Println("error validating sql.Open arguments")
		panic(err.Error())
	}
	defer db.Close()
	sID = strings.ReplaceAll(sID, " ", "_")
	rows, err := db.Query("SELECT q.itemid, q.stack, q.seller_name, q.price, q.buyer_name, q.sale, q.sell_date, a.name, a.aH FROM item_basic as a JOIN auction_house as q ON q.itemid = a.itemid WHERE a.name LIKE ? AND q.buyer_name IS NULL ORDER BY q.price LIMIT 30", "%"+sID+"%")
	if err != nil {
		fmt.Println("error selecting ah by buyer")
		panic(err.Error())
	}
	defer rows.Close()

	var items []*AHItem
	for rows.Next() {
		row := new(AHItem)
		if err := rows.Scan(&row.ItemID, &row.Stack, &row.SellerName, &row.Price, &row.BuyerName, &row.Sale, &row.SellDate, &row.ItemName, &row.Ah); err != nil {
			fmt.Printf("ah by buyer error: %v", err)
		}
		items = append(items, row)
	}

	jsonData, _ := json.Marshal(&items)
	w.Write(jsonData)
}

func GetAHByBuyer(w http.ResponseWriter, r *http.Request) {
	finishReq(w, r, "buyer_name")
}

func GetAHBySeller(w http.ResponseWriter, r *http.Request) {
	finishReq(w, r, "seller_name")
}

func finishReq(w http.ResponseWriter, r *http.Request, s string) {
	pathParams := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", corStr)
	sID := pathParams["sID"]

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		fmt.Println("error validating sql.Open arguments")
		panic(err.Error())
	}
	defer db.Close()

	rows, err := db.Query("SELECT q.itemid, q.stack, q.seller_name, q.price, q.buyer_name, q.sale, q.sell_date, a.name FROM auction_house as q JOIN item_basic AS a ON q.itemid = a.itemid WHERE q."+s+" =? ORDER BY date DESC LIMIT 30", sID)
	if err != nil {
		fmt.Println("error selecting ah by buyer")
		panic(err.Error())
	}
	defer rows.Close()

	var items []*AHItem
	for rows.Next() {
		row := new(AHItem)
		if err := rows.Scan(&row.ItemID, &row.Stack, &row.SellerName, &row.Price, &row.BuyerName, &row.Sale, &row.SellDate, &row.ItemName); err != nil {
			fmt.Printf("ah by buyer error: %v", err)
		}
		items = append(items, row)
	}

	jsonData, _ := json.Marshal(&items)
	w.Write(jsonData)
}
