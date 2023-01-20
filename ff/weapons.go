package ff

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Weapon struct {
	ItemID       int16
	Name         string
	Skill        byte
	SubSkill     byte
	ILvlSkill    int16
	ILvlParry    int16
	ILvlMagic    int16
	DmgType      int32
	Hit          byte
	Delay        int32
	Dmg          int32
	UnlockPoints int16
	Category     int32
}

type WS struct {
	WSId     int16
	Name     string
	Note     string
	Jobs     *[]byte
	JobStr   string
	Type     byte
	Level    int16
	Element  byte
	Range    byte
	AOE      byte
	PrimSC   byte
	SecSC    byte
	TertSC   byte
	MainOnly byte
	UnlockID byte
}

func GetWeapon(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", corStr)
	aID := -1
	var err error
	if val, ok := pathParams["aID"]; ok {
		aID, err = strconv.Atoi(val)
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
	rows, err := db.Query("SELECT * from item_weapon where itemId =?", aID)
	if err != nil {
		fmt.Println("error selecting")
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var row Weapon
		if err := rows.Scan(&row.ItemID, &row.Name, &row.Skill, &row.SubSkill, &row.ILvlSkill, &row.ILvlParry, &row.ILvlMagic, &row.DmgType, &row.Hit, &row.Delay, &row.Dmg, &row.UnlockPoints, &row.Category); err != nil {
			fmt.Printf("weapons error: %v", err)
		}
		w.Write([]byte(fmt.Sprintf(`{"itemID": %d, "name": "%s", "skill": %d, "subskill": %d, "ilvl_skill": %d, "ilvl_parry": %d, "ilvl_magic": %d, "dmgType": %d, "hit": %d, "delay": %d, "dmg": %d, "unlock_points": %d, "category": %d }`,
			row.ItemID, row.Name, row.Skill, row.SubSkill, row.ILvlSkill, row.ILvlParry, row.ILvlMagic, row.DmgType, row.Hit, row.Delay, row.Dmg, row.UnlockPoints, row.Category)))
	}
}

func GetWeaponsBySkill(w http.ResponseWriter, r *http.Request) {
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

	rows, err := db.Query("SELECT * from item_weapon where skill =? ORDER BY itemId", sID)
	if err != nil {
		fmt.Println("error selecting weapon by skill")
		panic(err.Error())
	}
	defer rows.Close()

	var weapons []*Weapon
	for rows.Next() {
		row := new(Weapon)
		if err := rows.Scan(&row.ItemID, &row.Name, &row.Skill, &row.SubSkill, &row.ILvlSkill, &row.ILvlParry, &row.ILvlMagic, &row.DmgType, &row.Hit, &row.Delay, &row.Dmg, &row.UnlockPoints, &row.Category); err != nil {
			fmt.Printf("weapons by skill error: %v", err)
		}
		weapons = append(weapons, row)
	}

	jsonData, _ := json.Marshal(&weapons)
	w.Write(jsonData)
}

func GetWeaponsByDmgType(w http.ResponseWriter, r *http.Request) {
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

	rows, err := db.Query("SELECT * from item_weapon where dmgType =? ORDER BY itemId", sID)
	if err != nil {
		fmt.Println("error selecting weapon by dmgType")
		panic(err.Error())
	}
	defer rows.Close()

	var weapons []*Weapon
	for rows.Next() {
		row := new(Weapon)
		if err := rows.Scan(&row.ItemID, &row.Name, &row.Skill, &row.SubSkill, &row.ILvlSkill, &row.ILvlParry, &row.ILvlMagic, &row.DmgType, &row.Hit, &row.Delay, &row.Dmg, &row.UnlockPoints, &row.Category); err != nil {
			fmt.Printf("weapons by dmgType error: %v", err)
		}
		weapons = append(weapons, row)
	}

	jsonData, _ := json.Marshal(&weapons)
	w.Write(jsonData)
}

func GetWSBySkillType(w http.ResponseWriter, r *http.Request) {
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

	rows, err := db.Query("SELECT q.weaponskillid, q.name, q.jobs, q.type, q.skilllevel, q.element, q.range, q.aoe, q.primary_sc, q.secondary_sc, q.tertiary_sc, q.main_only, q.unlock_id FROM weapon_skills as q WHERE q.type = ? AND q.skilllevel != 0 ORDER BY q.skilllevel", sID)
	if err != nil {
		fmt.Println("error selecting ws by type")
		panic(err.Error())
	}
	defer rows.Close()

	var skills []*WS
	for rows.Next() {
		row3 := new(WS)
		if err := rows.Scan(&row3.WSId, &row3.Name, &row3.Jobs, &row3.Type, &row3.Level, &row3.Element, &row3.Range, &row3.AOE, &row3.PrimSC, &row3.SecSC, &row3.TertSC, &row3.MainOnly, &row3.UnlockID); err != nil {
			fmt.Printf("job spell error: %v", err)
		}
		row3.Note = AbilityDes[row3.Name].Value
		if row3.Note != "" {
			row3.Name = AbilityDes[row3.Name].Name
		}
		skills = append(skills, row3)
	}
	// Puts REM weaponskills at end of list while keeping the list sorted by skill level
	rows, err = db.Query("SELECT q.weaponskillid, q.name, q.jobs, q.type, q.skilllevel, q.element, q.range, q.aoe, q.primary_sc, q.secondary_sc, q.tertiary_sc, q.main_only, q.unlock_id FROM weapon_skills as q WHERE q.type = ? AND q.skilllevel = 0", sID)
	if err != nil {
		fmt.Println("error selecting ws by type")
		panic(err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		row3 := new(WS)
		if err := rows.Scan(&row3.WSId, &row3.Name, &row3.Jobs, &row3.Type, &row3.Level, &row3.Element, &row3.Range, &row3.AOE, &row3.PrimSC, &row3.SecSC, &row3.TertSC, &row3.MainOnly, &row3.UnlockID); err != nil {
			fmt.Printf("job spell error: %v", err)
		}
		row3.Note = AbilityDes[row3.Name].Value
		if row3.Note != "" {
			row3.Name = AbilityDes[row3.Name].Name
		}
		skills = append(skills, row3)
	}

	jsonData, _ := json.Marshal(&skills)
	w.Write(jsonData)
}

func GetWSBySCType(w http.ResponseWriter, r *http.Request) {
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

	rows, err := db.Query("SELECT q.weaponskillid, q.name, q.jobs, q.type, q.skilllevel, q.element, q.range, q.aoe, q.primary_sc, q.secondary_sc, q.tertiary_sc, q.main_only, q.unlock_id FROM weapon_skills as q WHERE q.primary_sc = ? || q.secondary_sc = ? || q.tertiary_sc = ? ORDER BY q.type, q.skilllevel", sID, sID, sID)
	if err != nil {
		fmt.Println("error selecting ws by type")
		panic(err.Error())
	}
	defer rows.Close()

	var skills []*WS
	for rows.Next() {
		row3 := new(WS)
		if err := rows.Scan(&row3.WSId, &row3.Name, &row3.Jobs, &row3.Type, &row3.Level, &row3.Element, &row3.Range, &row3.AOE, &row3.PrimSC, &row3.SecSC, &row3.TertSC, &row3.MainOnly, &row3.UnlockID); err != nil {
			fmt.Printf("job spell error: %v", err)
		}
		row3.Note = AbilityDes[row3.Name].Value
		if row3.Note != "" {
			row3.Name = AbilityDes[row3.Name].Name
			row3.Note = "" // remove Note bc to lower memory usage bc not needed here, but wanted correctly formatted name
		}
		skills = append(skills, row3)
	}

	jsonData, _ := json.Marshal(&skills)
	w.Write(jsonData)
}

func GetMapForSC(w http.ResponseWriter, r *http.Request) {
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
	scl := GetMapSC(byte(sID))
	jsonData, _ := json.Marshal(scl)
	w.Write(jsonData)
}

/*
   {{SC_LIGHT, SC_LIGHT},SC_LIGHT_II},
   {{SC_DARKNESS, SC_DARKNESS}, SC_DARKNESS_II},

   // Level 2 Pairs
   {{SC_GRAVITATION, SC_DISTORTION}, SC_DARKNESS},
   {{SC_GRAVITATION, SC_FRAGMENTATION}, SC_FRAGMENTATION},

   {{SC_DISTORTION, SC_GRAVITATION}, SC_DARKNESS},
   {{SC_DISTORTION, SC_FUSION}, SC_FUSION},

   {{SC_FUSION, SC_GRAVITATION}, SC_GRAVITATION},
   {{SC_FUSION, SC_FRAGMENTATION}, SC_LIGHT},

   {{SC_FRAGMENTATION, SC_DISTORTION}, SC_DISTORTION},
   {{SC_FRAGMENTATION, SC_FUSION}, SC_LIGHT},

       // Level 1 Pairs
   {{SC_TRANSFIXION, SC_COMPRESSION}, SC_COMPRESSION},
   {{SC_TRANSFIXION, SC_SCISSION}, SC_DISTORTION},
   {{SC_TRANSFIXION, SC_REVERBERATION}, SC_REVERBERATION},

   {{SC_COMPRESSION, SC_TRANSFIXION}, SC_TRANSFIXION},
   {{SC_COMPRESSION, SC_DETONATION}, SC_DETONATION},

   {{SC_LIQUEFACTION, SC_SCISSION}, SC_SCISSION},
   {{SC_LIQUEFACTION, SC_IMPACTION}, SC_FUSION},

   {{SC_SCISSION, SC_LIQUEFACTION}, SC_LIQUEFACTION},
   {{SC_SCISSION, SC_REVERBERATION}, SC_REVERBERATION},
   {{SC_SCISSION, SC_DETONATION}, SC_DETONATION},

   {{SC_REVERBERATION, SC_INDURATION}, SC_INDURATION},
   {{SC_REVERBERATION, SC_IMPACTION}, SC_IMPACTION},

   {{SC_DETONATION, SC_COMPRESSION}, SC_GRAVITATION},
   {{SC_DETONATION, SC_SCISSION}, SC_SCISSION},

   {{SC_INDURATION, SC_COMPRESSION}, SC_COMPRESSION},
   {{SC_INDURATION, SC_REVERBERATION}, SC_FRAGMENTATION},
   {{SC_INDURATION, SC_IMPACTION}, SC_IMPACTION},

   {{SC_IMPACTION, SC_LIQUEFACTION}, SC_LIQUEFACTION},
   {{SC_IMPACTION, SC_DETONATION}, SC_DETONATION}
*/

type SCMap struct {
	Lvl1Open  []SCInner
	Lvl1Close []SCInner
	Lvl2Open  *SCInner
	Lvl2Close *SCInner
	Lvl3Open  *SCInner
	Lvl3Close *SCInner
}

type SCInner struct {
	WSC    byte
	Create byte
}

func GetMapSC(val byte) SCMap {
	var scl SCMap
	// Key is first SC
	openm := map[byte][]SCInner{
		1: {{2, 2}, {5, 5}},
		2: {{1, 1}, {6, 6}},
		3: {{4, 4}},
		4: {{3, 3}, {5, 5}},
		5: {{7, 7}, {8, 8}},
		6: {{4, 4}},
		7: {{2, 2}, {8, 8}},
		8: {{3, 3}, {6, 6}},
	}
	// Key is second SC
	closem := map[byte][]SCInner{
		1: {{2, 1}},
		2: {{7, 2}, {1, 2}},
		3: {{4, 3}, {8, 3}},
		4: {{3, 4}, {6, 4}},
		5: {{1, 5}, {4, 5}},
		6: {{8, 6}, {4, 6}, {2, 6}},
		7: {{5, 7}},
		8: {{7, 8}, {5, 8}},
	}

	// Key is first SC
	openm2 := map[byte]*SCInner{
		1:  {4, 10},
		3:  {8, 12},
		4:  {6, 6},
		6:  {2, 9},
		7:  {5, 12},
		9:  {12, 12},
		10: {11, 11},
		11: {9, 9},
		12: {10, 10},
	}
	// Key is second SC
	closem2 := map[byte]*SCInner{
		2:  {6, 9},
		4:  {1, 10},
		5:  {7, 12},
		8:  {3, 11},
		9:  {11, 9},
		10: {12, 10},
		11: {10, 11},
		12: {9, 12},
	}

	openm3 := map[byte]*SCInner{
		9:  {10, 14},
		10: {9, 14},
		11: {12, 13},
		12: {11, 13},
		13: {13, 15},
		14: {14, 16},
	}
	// Key is second SC
	closem3 := map[byte]*SCInner{
		9:  {10, 14},
		10: {9, 14},
		11: {12, 13},
		12: {11, 13},
		13: {13, 15},
		14: {14, 16},
	}

	scl.Lvl1Open = openm[val]
	scl.Lvl1Close = closem[val]
	scl.Lvl2Open = openm2[val]
	scl.Lvl2Close = closem2[val]
	scl.Lvl3Close = closem3[val]
	scl.Lvl3Open = openm3[val]

	return scl
}
