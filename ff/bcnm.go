package ff

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type BCNM struct {
	ID       int16
	ZoneId   int16
	Name     string
	ZName    string
	Limit    int16
	Cap      byte
	PtSize   byte
	LootID   int16
	IsMish   byte
	RuleStr  *string
	Mobs     []*BcnmMob
	Treasure []*BcnmNpc
}

type BcnmMob struct {
	ID         int
	Conditions byte
	Name       string
	Group      int32
}

type BcnmNpc struct {
	ID   int
	Loot []BCTGroup
}

func GetBCNMs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", corStr)
	bcnms := GetBCList()
	jsonData, _ := json.Marshal(&bcnms)
	w.Write(jsonData)
}

func GetBCList() []*BCNM {
	var err error
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		fmt.Println("error validating sql.Open arguments")
		panic(err.Error())
	}
	defer db.Close()

	rows, err := db.Query("SELECT q.bcnmId, q.zoneId, q.name, q.timeLimit, q.levelCap, q.partySize, q.lootDropId, q.rules, q.isMission, a.name FROM bcnm_info as q JOIN zone_settings as a ON q.zoneId = a.zoneid WHERE q.levelCap != 0 AND q.levelCap <= ? ORDER BY q.levelCap, q.isMission DESC", lvlcap)
	if err != nil {
		fmt.Println("error selecting bcnms")
		panic(err.Error())
	}
	defer rows.Close()
	var bcnms []*BCNM
	for rows.Next() {
		rules := 0
		row := new(BCNM)
		if err := rows.Scan(&row.ID, &row.ZoneId, &row.Name, &row.Limit, &row.Cap, &row.PtSize, &row.LootID, &rules, &row.IsMish, &row.ZName); err != nil {
			fmt.Printf("bcnms error: %v", err)
		}
		row.RuleStr = GetBcnmRules(rules)
		bcnms = append(bcnms, row)
	}
	rows.Close()
	db.Close()
	return bcnms
}

func GetBCNMDets(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", corStr)
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

	rules := 0
	row := new(BCNM)
	err = db.QueryRow("SELECT q.bcnmId, q.zoneId, q.name, q.timeLimit, q.levelCap, q.partySize, q.lootDropId, q.rules, q.isMission, a.name FROM bcnm_info as q JOIN zone_settings as a ON q.zoneId = a.zoneid WHERE q.bcnmId = ?", sID).Scan(&row.ID, &row.ZoneId, &row.Name, &row.Limit, &row.Cap, &row.PtSize, &row.LootID, &rules, &row.IsMish, &row.ZName)
	if err != nil {
		fmt.Println("error selecting bcnm")
		panic(err.Error())
	}
	row.RuleStr = GetBcnmRules(rules)

	rows, err := db.Query("SELECT q.monsterId, q.conditions, a.mobname, a.groupid FROM bcnm_battlefield as q JOIN mob_spawn_points as a ON q.monsterId = a.mobid WHERE q.bcnmId = ? AND q.battlefieldNumber = 1", sID)
	if err != nil {
		fmt.Println("error selecting bcnm mobs")
		panic(err.Error())
	}
	defer rows.Close()
	var mobs []*BcnmMob
	for rows.Next() {
		rowm := new(BcnmMob)
		if err := rows.Scan(&rowm.ID, &rowm.Conditions, &rowm.Name, &rowm.Group); err != nil {
			fmt.Printf("bcnm mobs error: %v", err)
		}
		mobs = append(mobs, rowm)
	}

	rows2, err := db.Query("SELECT q.npcId FROM bcnm_treasure_chests as q WHERE q.bcnmId = ? AND q.battlefieldNumber = 1", sID)
	if err != nil {
		fmt.Println("error selecting bcnm mobs")
		panic(err.Error())
	}
	defer rows2.Close()
	var npcs []*BcnmNpc
	for rows2.Next() {
		rowp := new(BcnmNpc)
		if err := rows2.Scan(&rowp.ID); err != nil {
			fmt.Printf("bcnm mobs error: %v", err)
		}
		if row.LootID > 0 { //try to populate loot list
			tempStr := strconv.Itoa(int(row.LootID))
			rowp.Loot = BCNMTreasure[tempStr]
		}
		npcs = append(npcs, rowp)
	}
	row.Mobs = mobs
	row.Treasure = npcs

	jsonData, _ := json.Marshal(&row)
	w.Write(jsonData)
	rows2.Close()
	rows.Close()
	db.Close()
}

func GetBcnmRules(rules int) *string {
	str := ""
	if (rules & 0x0001) > 0 {
		str += "Allows Subjobs, "
	}
	if (rules & 0x0002) > 0 {
		str += "Can lose EXP, "
	}
	if (rules & 0x0004) > 0 {
		str += "Removes at 3 min, "
	}
	if (rules & 0x0008) > 0 {
		str += "Treasure on Win, "
	}
	if (rules & 0x0010) > 0 {
		str += "Maat Fight, "
	}
	str = strings.TrimSuffix(str, ", ")
	return &str
}
