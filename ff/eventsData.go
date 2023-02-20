package ff

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/gorilla/mux"
)

var EventData FFXIEvents

type Actors struct {
	Acts []ActName `json:"acts"`
}

type ActName struct {
	Name string `json:"name"`
	Id   int32  `json:"Id"`
}

// NAMES MUST MATCH NAME IN JSON FILE TO MARSHAL PROPERLY
type FFXIEvents struct {
	Actors []Actor `json:"Actors"`
	Id     int32   `json:"Id"`
	Name   string  `json:"name"`
}

type Actor struct {
	Events []SingleEvent `json:"Events"`
	Id     int32         `json:"Id"`
	Name   string        `json:"name"`
}

type SingleEvent struct {
	Dialogue []EDialog `json:"dialogue"`
	Id       int32     `json:"id"`
}

type EDialog struct {
	Id  int32  `json:"Id"`
	Str string `json:"string"`
}

// JSON from Scripts
type PupMod struct {
	Maneuver byte
	Frame    string
	Mod      string
	Val      string
}

type BCTGroup struct {
	Group      byte
	BCTreasure []*BCTreasure
}

type BCTreasure struct {
	ItemId int32
	Name   string
	Rate   int16
}

type NameVal struct {
	Name  string
	Value string
}

type NameNum struct {
	Name  string
	Value byte
}

type Assault struct {
	Index     byte
	Level     byte
	Rank      byte
	Orders    string
	Area      string
	Name      string
	Objective string
}

func Start() {
	iterate(jsonPath)
}

func iterate(path string) {
	EventData := make(map[string]FFXIEvents)
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf(err.Error())
		}

		// Open our jsonFile
		jsonFile, err := os.Open(jsonPath + info.Name())
		// if we os.Open returns an error then handle it
		if err != nil {
			fmt.Println(err)
		}
		// defer the closing of our jsonFile so that we can parse it later on
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)

		var events FFXIEvents
		json.Unmarshal(byteValue, &events)
		EventData[events.Name] = events
		return nil
	})
}

func GetEventsByZone(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	InitHeader(w)

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

	// Open our jsonFile
	jsonFile, err := os.Open(fmt.Sprintf("%s%d%s", jsonPath, aID, ".json"))
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var events FFXIEvents
	json.Unmarshal(byteValue, &events)

	var actNames Actors
	//	var actNames = make(map[int32]ActName)

	//SENDS JUST THE NAMES OF EVENT ACTORS
	for _, actor := range events.Actors {
		candidates := ActName{
			Name: actor.Name,
			Id:   actor.Id,
		}
		actNames.Acts = append(actNames.Acts, candidates)
	}
	//remove the : if sorting a slice
	sort.Slice(actNames.Acts[:], func(i, j int) bool {
		return actNames.Acts[i].Name < actNames.Acts[j].Name
	})
	lastByte, _ := json.Marshal(actNames)

	w.Write(lastByte)
}

func GetEventsActorData(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	InitHeader(w)

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

	jID := -1
	if val, ok := pathParams["jID"]; ok {
		jID, err = strconv.Atoi(val)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "need a number"}`))
			return
		}
	}

	// Open our jsonFile
	jsonFile, err := os.Open(fmt.Sprintf("%s%d%s", jsonPath, aID, ".json"))
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var events FFXIEvents
	json.Unmarshal(byteValue, &events)

	var act Actor
	//SENDS JUST THE NAMES OF EVENT ACTORS
	for _, actor := range events.Actors {
		if actor.Id == int32(jID) {
			act = actor
			break
		}
	}

	lastByte, _ := json.Marshal(act)
	w.Write(lastByte)
}

func GetAttachmentJson() {
	mods := make(map[string][]PupMod)
	// open file
	f, err := os.Open(jsonPath + "Attachments.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	byteValue, _ := ioutil.ReadAll(f)

	err = json.Unmarshal(byteValue, &mods)

	if err != nil {
		panic(err)
	}
	f.Close()
	AttachmentMods = mods
}

func GetBCNMJson() {
	treasure := make(map[string][]BCTGroup)
	// open file
	f, err := os.Open(jsonPath + "BCNMLoots.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	byteValue, _ := ioutil.ReadAll(f)

	err = json.Unmarshal(byteValue, &treasure)

	if err != nil {
		panic(err)
	}
	f.Close()
	BCNMTreasure = treasure
}

func GetDesJson() {
	keyitems := make(map[string]NameVal)
	f, err := os.Open(jsonPath + "Key_Items.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	byteValue, _ := ioutil.ReadAll(f)

	err = json.Unmarshal(byteValue, &keyitems)

	if err != nil {
		panic(err)
	}
	f.Close()
	KeyItems = keyitems

	titles := make(map[string]string)
	f, err = os.Open(jsonPath + "Titles.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	byteValue, _ = ioutil.ReadAll(f)

	err = json.Unmarshal(byteValue, &titles)

	if err != nil {
		panic(err)
	}
	f.Close()
	Titles = titles

	abils := make(map[string]NameVal)
	f, err = os.Open(jsonPath + "Ability_Des.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	byteValue, _ = ioutil.ReadAll(f)

	err = json.Unmarshal(byteValue, &abils)

	if err != nil {
		panic(err)
	}
	f.Close()
	AbilityDes = abils

	spells := make(map[string]string)
	f, err = os.Open(jsonPath + "Spell_Des.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	byteValue, _ = ioutil.ReadAll(f)

	err = json.Unmarshal(byteValue, &spells)

	if err != nil {
		panic(err)
	}
	f.Close()
	SpellDes = spells

	items := make(map[string]string)

	f, err = os.Open(jsonPath + "Item_Des.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	byteValue, _ = ioutil.ReadAll(f)

	err = json.Unmarshal(byteValue, &items)

	if err != nil {
		panic(err)
	}
	f.Close()
	ItemDes = items
}

func GetMissionsJson() {
	missions := make(map[string]map[string]NameVal)
	for k := range expansionsOn {
		f, err := os.Open(jsonPath +  expansionsOn[k] + "_missions.json")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		if expansionsOn[k] == "assault" {
			asltItems := make(map[string]Assault)

			byteValue, _ := ioutil.ReadAll(f)

			err = json.Unmarshal(byteValue, &asltItems)

			if err != nil {
				panic(err)
			}
			AssaultMissions = asltItems
		} else {
			mishItems := make(map[string]NameVal)

			byteValue, _ := ioutil.ReadAll(f)

			err = json.Unmarshal(byteValue, &mishItems)

			if err != nil {
				panic(err)
			}
			missions[expansionsOn[k]] = mishItems
		}
		f.Close()
	}
	Missions = missions
}

func GetZoneMapPaths() {
	items := make(map[string][]string)

	f, err := os.Open(jsonPath + "Map_Paths.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	byteValue, _ := ioutil.ReadAll(f)

	err = json.Unmarshal(byteValue, &items)

	if err != nil {
		panic(err)
	}
	f.Close()
	ZoneMapPaths = items
}
