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

type Trait struct {
	TraitID    byte
	Name       string
	Note       string
	Job        byte
	Level      byte
	Rank       byte
	Modifier   int
	Value      int
	ContentTag *string
	MeritID    int16
	Mod        *Mods
}

type SkillRank struct {
	SkillId byte
	Name    string
	Value   byte
}

type ValByte struct {
	Value int16
}
type ValString struct {
	Value *string
}

type Merit struct {
	MeritID   int16
	Name      *string
	Upgrade   byte
	Value     int16
	Jobs      int
	UpgradeID byte
	CatID     byte
}

func GetTrait(w http.ResponseWriter, r *http.Request) {
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
	/*
		err = db.Ping()
		if err != nil {
			fmt.Println("error verifying connection with db.Ping")
			panic(err.Error())
		}
	*/
	rows, err := db.Query("SELECT * from traits where traitid =?", aID)
	if err != nil {
		fmt.Println("error selecting trait")
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var row Trait
		if err := rows.Scan(&row.TraitID, &row.Name, &row.Job, &row.Level, &row.Rank, &row.Modifier, &row.Value, &row.ContentTag, &row.MeritID); err != nil {
			fmt.Printf("traits error: %v", err)
		}
		w.Write([]byte(fmt.Sprintf(`{"traitID": %d, "name": "%s", "job": %d, "level": %d, "rank": %d, "modifier": %d, "value": %d, "meritId": %d }`,
			row.TraitID, row.Name, row.Job, row.Level, row.Rank, row.Modifier, row.Value, row.MeritID)))
	}
}

func GetTraitsByJob(w http.ResponseWriter, r *http.Request) {
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

	rows, err := db.Query("SELECT * from traits where job =? ORDER BY level", sID)
	if err != nil {
		fmt.Println("error selecting trait by job")
		panic(err.Error())
	}
	defer rows.Close()

	var traits []*Trait
	for rows.Next() {
		row := new(Trait)
		if err := rows.Scan(&row.TraitID, &row.Name, &row.Job, &row.Level, &row.Rank, &row.Modifier, &row.Value, &row.ContentTag, &row.MeritID); err != nil {
			fmt.Printf("traits by job error: %v", err)
		}
		row.Note = AbilityDes[strings.ReplaceAll(row.Name, " ", "_")].Value
		if row.Note != "" {
			row.Name = AbilityDes[strings.ReplaceAll(row.Name, " ", "_")].Name
		}
		row.Mod = new(Mods)
		row.Mod.Id = row.Modifier
		row.Mod.Val = row.Value
		row.Mod = GetMods(row.Mod)
		traits = append(traits, row)
	}

	jsonData, _ := json.Marshal(&traits)
	w.Write(jsonData)
}

func GetTraitsByLevel(w http.ResponseWriter, r *http.Request) {
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

	rows, err := db.Query("SELECT * from traits where level =? ORDER BY job", sID)
	if err != nil {
		fmt.Println("error selecting trait by level")
		panic(err.Error())
	}
	defer rows.Close()

	var traits []*Trait
	for rows.Next() {
		row := new(Trait)
		if err := rows.Scan(&row.TraitID, &row.Name, &row.Job, &row.Level, &row.Rank, &row.Modifier, &row.Value, &row.ContentTag, &row.MeritID); err != nil {
			fmt.Printf("traits by level error: %v", err)
		}
		row.Note = AbilityDes[strings.ReplaceAll(row.Name, " ", "_")].Value
		if row.Note != "" {
			row.Name = AbilityDes[strings.ReplaceAll(row.Name, " ", "_")].Name
		}
		traits = append(traits, row)
	}

	jsonData, _ := json.Marshal(&traits)
	w.Write(jsonData)
}

func GetSkillRanksByJob(w http.ResponseWriter, r *http.Request) {
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

	rows, err := db.Query(fmt.Sprintf("%s%s%s", "SELECT skillid, name, ", getShortJobString(sID), " from skill_ranks"))
	if err != nil {
		fmt.Println("error selecting skill by job")
		panic(err.Error())
	}
	defer rows.Close()

	var skillRanks []*SkillRank
	for rows.Next() {
		row := new(SkillRank)
		if err := rows.Scan(&row.SkillId, &row.Name, &row.Value); err != nil {
			fmt.Printf("skill ranks by job error: %v", err)
		}
		skillRanks = append(skillRanks, row)
	}

	jsonData, _ := json.Marshal(&skillRanks)
	w.Write(jsonData)
}

func GetSkillRanksByLevel(w http.ResponseWriter, r *http.Request) {
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

	rID := -1
	if val, ok := pathParams["rID"]; ok {
		rID, err = strconv.Atoi(val)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "need a number"}`))
			return
		}
	}

	skillRanks := GetSkillRanks(sID, rID)
	jsonData, _ := json.Marshal(&skillRanks)
	w.Write(jsonData)
}

func GetSkillRanks(sID int, rID int) []*ValByte {

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		fmt.Println("error validating sql.Open arguments")
		panic(err.Error())
	}
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("%s%s%s%d", "SELECT ", getRankCol(rID), " FROM skill_caps WHERE level =", sID))
	if err != nil {
		fmt.Println("error selecting skill by level")
		panic(err.Error())
	}
	defer rows.Close()

	var skillRanks []*ValByte
	for rows.Next() {
		row := new(ValByte)
		if err := rows.Scan(&row.Value); err != nil {
			fmt.Printf("skill ranks by level error: %v", err)
		}
		skillRanks = append(skillRanks, row)
	}
	return skillRanks
}

func GetMeritsByJob(w http.ResponseWriter, r *http.Request) {
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
	numPow := IntPow(2, sID-1)
	rows, err := db.Query("SELECT meritid, name, upgrade, value, jobs, upgradeid, catagoryid FROM merits WHERE jobs = ?", numPow)
	if err != nil {
		fmt.Println("error selecting merits by job")
		panic(err.Error())
	}
	defer rows.Close()

	var merits []*Merit
	for rows.Next() {
		row := new(Merit)
		if err := rows.Scan(&row.MeritID, &row.Name, &row.Upgrade, &row.Value, &row.Jobs, &row.UpgradeID, &row.CatID); err != nil {
			fmt.Printf("merits by job error: %v", err)
		}
		merits = append(merits, row)
	}

	jsonData, _ := json.Marshal(&merits)
	w.Write(jsonData)
}

func getShortJobString(num int) string {
	switch num {
	case 1:
		return "war"
	case 2:
		return "mnk"
	case 3:
		return "whm"
	case 4:
		return "blm"
	case 5:
		return "rdm"
	case 6:
		return "thf"
	case 7:
		return "pld"
	case 8:
		return "drk"
	case 9:
		return "bst"
	case 10:
		return "brd"
	case 11:
		return "rng"
	case 12:
		return "sam"
	case 13:
		return "nin"
	case 14:
		return "drg"
	case 15:
		return "smn"
	case 16:
		return "blu"
	case 17:
		return "cor"
	case 18:
		return "pup"
	case 19:
		return "dnc"
	case 20:
		return "sch"
	case 21:
		return "geo"
	case 22:
		return "run"
	default:
		return ""
	}
}

func getRankCol(num int) string {
	switch num {
	case 0:
		return "r0"
	case 1:
		return "r1"
	case 2:
		return "r2"
	case 3:
		return "r3"
	case 4:
		return "r4"
	case 5:
		return "r5"
	case 6:
		return "r6"
	case 7:
		return "r7"
	case 8:
		return "r8"
	case 9:
		return "r9"
	case 10:
		return "r10"
	case 11:
		return "r11"
	case 12:
		return "r12"
	case 13:
		return "r13"
	default:
		return ""
	}
}
