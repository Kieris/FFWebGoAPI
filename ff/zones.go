package ff

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	// "github.com/gorilla/mux"
)

type Zone struct {
	ZoneId   int16
	ZoneType int16
	Name     *string
}

type ZoneMob struct {
	GroupId   int32
	PoolId    int32
	Name      *string
	SpawnType int16
	Respawn   int32
	MaxLevel  byte
}

func GetZones(w http.ResponseWriter, r *http.Request) {
	InitHeader(w)

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		fmt.Println("error validating sql.Open arguments")
		panic(err.Error())
	}
	defer db.Close()

	rows, err := db.Query("SELECT zoneid, zonetype, name from zone_settings ORDER BY name")
	if err != nil {
		fmt.Println("error selecting")
		panic(err.Error())
	}
	defer rows.Close()

	var zones []*Zone
	for rows.Next() {
		row := new(Zone)
		if err := rows.Scan(&row.ZoneId, &row.ZoneType, &row.Name); err != nil {
			fmt.Printf("job abilities error: %v", err)
		}
		zones = append(zones, row)
	}

	jsonData, _ := json.Marshal(&zones)
	w.Write(jsonData)
}

func GetZoneMobs(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	InitHeader(w)
	sID := -1
	var err error
	if val, ok := pathParams["sID"]; ok {
		sID, err = strconv.Atoi(val)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "need a number"}`))
			return
		}
	}

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		fmt.Println("error validating sql.Open arguments")
		panic(err.Error())
	}
	defer db.Close()

	rows, err := db.Query("SELECT q.groupid, q.poolid, q.respawntime, q.spawntype, q.name, q.maxLevel FROM mob_groups as q WHERE q.zoneid = ? ORDER BY q.spawntype, q.name", sID)
	if err != nil {
		fmt.Println("error selecting mob zone")
		panic(err.Error())
	}
	defer rows.Close()

	var items []*ZoneMob
	for rows.Next() {
		row := new(ZoneMob)
		if err := rows.Scan(&row.GroupId, &row.PoolId, &row.Respawn, &row.SpawnType, &row.Name, &row.MaxLevel); err != nil {
			fmt.Printf("mob short zone error: %v", err)
		}
		items = append(items, row)
	}

	jsonData, _ := json.Marshal(&items)
	w.Write(jsonData)
}

func GetZoneMaps(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	InitHeader(w)
	sID := -1
	var err error
	if val, ok := pathParams["sID"]; ok {
		sID, err = strconv.Atoi(val)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "need a number"}`))
			return
		}
	}
	maps := ZoneMapPaths[strconv.Itoa(sID)]

	jsonData, _ := json.Marshal(&maps)
	w.Write(jsonData)
}

func GetTuts(w http.ResponseWriter, r *http.Request) {
	InitHeader(w)
	jsonData, _ := json.Marshal(&tuts)
	w.Write(jsonData)
}
