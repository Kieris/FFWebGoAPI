package ff

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type MapToIndexContainer struct {
	Type   string
	Values []*MapToInIndex
}

type MapToInIndex struct {
	Index int
	Name  string
}

func GetMissions(w http.ResponseWriter, r *http.Request) {
	InitHeader(w)
	mishList := [len(expansionsOn)]NameVal{}
	for k := range expansionsOn {
		mish := new(NameVal)
		mish.Name = missionMap[expansionsOn[k]]
		mish.Value = expansionsOn[k]
		mishList[k] = *mish
	}
	jsonData, _ := json.Marshal(&mishList)
	w.Write(jsonData)
}

func GetMissionList(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	InitHeader(w)
	sID := pathParams["sID"]

	var retList []*MapToInIndex

	// map keys needed to be strings but not need int
	for k, v := range Missions[sID] {
		num, err := strconv.Atoi(k)
		if err == nil {
			temp := new(MapToInIndex)
			temp.Index = num
			temp.Name = v.Name
			retList = append(retList, temp)
		}
	}

	// put the missions in the right order
	sort.Slice(retList[:], func(i, j int) bool {
		return retList[i].Index < retList[j].Index
	})

	retFinal := new(MapToIndexContainer)
	retFinal.Type = missionMap[sID]
	retFinal.Values = retList

	jsonData, _ := json.Marshal(&retFinal)
	w.Write(jsonData)
}

func GetAssaultList(w http.ResponseWriter, r *http.Request) {
	InitHeader(w)
	keys := []Assault{}
	for _, val := range AssaultMissions {
		if val.Level <= lvlcap {
			keys = append(keys, val)
		}
	}

	// Sort by level
	sort.Slice(keys[:], func(i, j int) bool {
		return keys[i].Level < keys[j].Level
	})
	jsonData, _ := json.Marshal(&keys)
	w.Write(jsonData)
}

func GetCatMission(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	InitHeader(w)
	sID := pathParams["sID"]
	itemID := pathParams["itemID"]

	retItem := Missions[sID][itemID]

	jsonData, _ := json.Marshal(&retItem)
	w.Write(jsonData)
}
