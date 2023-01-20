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

type Ability struct {
	AbilityID     int32
	Name          string
	Note          string
	Job           byte
	Level         byte
	ValidTarget   int32
	RecastTime    int32
	RecastID      int32
	Message1      int32
	Message2      int32
	Animation     int32
	AnimationTime int32
	CastTime      int32
	ActionType    byte
	Range         float32
	IsAOE         byte
	CE            int32
	VE            int32
	MeritModID    int32
	AddType       int32
	ContentTag    *string
}

func GetAbility(w http.ResponseWriter, r *http.Request) {
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
	rows, err := db.Query("SELECT * from abilities where abilityId =?", aID)
	if err != nil {
		fmt.Println("error selecting")
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var row Ability
		if err := rows.Scan(&row.AbilityID, &row.Name, &row.Job, &row.Level, &row.ValidTarget, &row.RecastTime, &row.RecastID, &row.Message1, &row.Message2, &row.Animation, &row.AnimationTime, &row.CastTime,
			&row.ActionType, &row.Range, &row.IsAOE, &row.CE, &row.VE, &row.MeritModID, &row.AddType, &row.ContentTag); err != nil {
			fmt.Printf("abilities error: %v", err)
		}
		w.Write([]byte(fmt.Sprintf(`{"abilityID": %d, "name": "%s", "job": "%s", "level": %d, "validTarget": %d, "recastTime": %d, "recastId": %d, "message1": %d, "message2": %d, "animation": %d, "animationTime": %d, "castTime": %d, "actionType": %d, "range": %1f, "isAOE": %d, "CE": %d, "VE": %d, "meritModID": %d, "addType": %d }`,
			row.AbilityID, row.Name, GetJobString(row.Job), row.Level, row.ValidTarget, row.RecastTime,
			row.RecastID, row.Message1, row.Message2, row.Animation, row.AnimationTime, row.CastTime, row.ActionType, row.Range, row.IsAOE, row.CE, row.VE, row.MeritModID, row.AddType)))
	}
}

func GetAbilitiesByJob(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", corStr)
	jID := -1
	var err error
	if val, ok := pathParams["jID"]; ok {
		jID, err = strconv.Atoi(val)
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
	// 512 - 632, 656-662 is smn pets abs. 639-654 drg, 672 -778 bst
	rows, err := db.Query("SELECT * from abilities where job =? AND level <= ? AND abilityid < 512 ORDER BY level", jID, lvlcap)
	if err != nil {
		fmt.Println("error selecting")
		panic(err.Error())
	}
	defer rows.Close()

	var abilities []*Ability
	for rows.Next() {
		row := new(Ability)
		if err := rows.Scan(&row.AbilityID, &row.Name, &row.Job, &row.Level, &row.ValidTarget, &row.RecastTime, &row.RecastID, &row.Message1, &row.Message2, &row.Animation, &row.AnimationTime, &row.CastTime,
			&row.ActionType, &row.Range, &row.IsAOE, &row.CE, &row.VE, &row.MeritModID, &row.AddType, &row.ContentTag); err != nil {
			fmt.Printf("job abilities error: %v", err)
		}
		row.Note = AbilityDes[row.Name].Value
		if row.Note != "" {
			row.Name = AbilityDes[row.Name].Name
		}
		abilities = append(abilities, row)
	}

	jsonData, _ := json.Marshal(&abilities)
	w.Write(jsonData)
}

func GetMiscNotes(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", corStr)
	jID := pathParams["jID"]
	typ := pathParams["typ"]

	jID = strings.ToLower(strings.ReplaceAll(jID, "'", ""))
	jID = strings.ReplaceAll(jID, " ", "_")

	s := GetScriptDets(typ + "/" + jID)
	jsonData, _ := json.Marshal(&s)
	w.Write(jsonData)

}
