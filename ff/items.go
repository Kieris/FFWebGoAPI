package ff

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Equipment struct {
	ItemID     int16
	Level      byte
	ILevel     byte
	Name       *string
	JobsInt    int
	Jobs       string
	Model      int16
	ShieldSize byte
	Script     int16
	Slot       int64
	RSlot      int16
	Race       int16
}

type Item struct {
	ItemID    int
	SubID     int16
	Name      string
	SortName  *string
	StackSize byte
	Flags     int64
	Ah        byte
	NoSale    byte
	BaseSell  int32
	Modifs    []*Mods
	Armor     *Equipment
	Weap      *Weapon
	Notes     string
}

type ItemDets struct {
	//ItemId in drops is used for dropId to keep same struct as other functions
	Drops   []*DropMob
	Recipes []*Recipe
}

type DropMob struct {
	Drop      *Drop
	Name      *string
	ZoneName  *string
	ZoneId    int
	PoolId    int
	GrpId     int
	SpawnType int16
}

func GetItemByName(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	InitHeader(w)
	sID := pathParams["sID"]

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		fmt.Println("error validating sql.Open arguments")
		panic(err.Error())
	}
	defer db.Close()
	sID = strings.ReplaceAll(sID, " ", "_")
	rows, err := db.Query("SELECT q.itemid, q.subid, q.name, q.sortname, q.stackSize, q.flags, q.aH, q.NoSale, q.BaseSell FROM item_basic as q WHERE q.name LIKE ? LIMIT 30", "%"+sID+"%")
	if err != nil {
		fmt.Println("error selecting item by name")
		panic(err.Error())
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		row := new(Item)
		if err := rows.Scan(&row.ItemID, &row.SubID, &row.Name, &row.SortName, &row.StackSize, &row.Flags, &row.Ah, &row.NoSale, &row.BaseSell); err != nil {
			fmt.Printf("item by name error: %v", err)
		}
		erows, err := db.Query("SELECT q.modId, q.value FROM item_mods as q WHERE q.itemId = ?", row.ItemID)
		if err != nil {
			fmt.Println("error selecting mods for item name")
			panic(err.Error())
		}
		defer erows.Close()

		row.SortName = nil // making null because it is not necessary after this point
		if row.ItemID >= 8449 && row.ItemID <= 8683 {
			row.Notes = "This item is only available in the Puppetmaster attachment menu"
		} else {
			row.Notes = ItemDes[strconv.Itoa(int(row.ItemID))]
		}
		var mds []*Mods
		for erows.Next() {
			erow := new(Mods)
			if err := erows.Scan(&erow.Id, &erow.Val); err != nil {
				fmt.Printf("item name mod error: %v", err)
			}
			erow = GetMods(erow)
			mds = append(mds, erow)
		}
		row.Modifs = mds

		arows, err := db.Query("SELECT q.itemId, q.name, q.level, q.ilevel, q.jobs, q.MId, q.shieldSize, q.scriptType, q.slot, q.rslot, q.race FROM item_equipment as q WHERE q.itemId = ?", row.ItemID)
		if err != nil {
			fmt.Println("error selecting equip for item")
			panic(err.Error())
		}
		defer arows.Close()

		var eq Equipment
		for arows.Next() {
			rowa := new(Equipment)
			if err := arows.Scan(&rowa.ItemID, &rowa.Name, &rowa.Level, &rowa.ILevel, &rowa.JobsInt, &rowa.Model, &rowa.ShieldSize, &rowa.Script, &rowa.Slot, &rowa.RSlot, &rowa.Race); err != nil {
				fmt.Printf("equipment job error: %v", err)
			}
			eq = *rowa
		}

		if eq.Level <= lvlcap { // filtering out anything over the level cap
			if eq.JobsInt > 0 { //leave armor null in struct unless it has real values
				row.Armor = &eq
				if row.ItemID >= 17860 && row.ItemID <= 17906 {
					//least bst item notes
				} else {
					row.Notes = ""
				}
				if row.Armor.JobsInt == 4194303 { // check for all jobs first
					row.Armor.Jobs = "All Jobs"
				} else {
					arr := GetPowersOfTwo(row.Armor.JobsInt)
					row.Armor.Jobs = GetJobsString(arr)
				}

				//get weapon data
				wrows, err := db.Query("SELECT * from item_weapon where itemId =?", row.ItemID)
				if err != nil {
					fmt.Println("error selecting weapon for item")
					panic(err.Error())
				}
				defer wrows.Close()

				for wrows.Next() {
					var wrow Weapon
					if err := wrows.Scan(&wrow.ItemID, &wrow.Name, &wrow.Skill, &wrow.SubSkill, &wrow.ILvlSkill, &wrow.ILvlParry, &wrow.ILvlMagic, &wrow.DmgType, &wrow.Hit, &wrow.Delay, &wrow.Dmg, &wrow.UnlockPoints, &wrow.Category); err != nil {
						fmt.Printf("weapons item error: %v", err)
					}
					row.Weap = &wrow
				}
			}
			items = append(items, row)
		}

	}
	jsonData, _ := json.Marshal(&items)
	w.Write(jsonData)
}

func GetItemByID(w http.ResponseWriter, r *http.Request) {
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

	var items = GetItems(sID)
	jsonData, _ := json.Marshal(&items)
	w.Write(jsonData)
}

func GetItems(sID int) []*Item {
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		fmt.Println("error validating sql.Open arguments")
		panic(err.Error())
	}
	defer db.Close()

	rows, err := db.Query("SELECT q.itemid, q.subid, q.name, q.sortname, q.stackSize, q.flags, q.aH, q.NoSale, q.BaseSell FROM item_basic as q WHERE q.itemid = ? LIMIT 1", sID)
	if err != nil {
		fmt.Println("error selecting item by name")
		panic(err.Error())
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		row := new(Item)
		if err := rows.Scan(&row.ItemID, &row.SubID, &row.Name, &row.SortName, &row.StackSize, &row.Flags, &row.Ah, &row.NoSale, &row.BaseSell); err != nil {
			fmt.Printf("item by name error: %v", err)
		}
		erows, err := db.Query("SELECT q.modId, q.value FROM item_mods as q WHERE q.itemId = ?", row.ItemID)
		if err != nil {
			fmt.Println("error selecting mods for item name")
			panic(err.Error())
		}
		defer erows.Close()

		if row.ItemID >= 8449 && row.ItemID <= 8683 {
			row.Notes = "This item is only available in the Puppetmaster attachment menu"
		} else {
			row.Notes = ItemDes[strconv.Itoa(sID)]
		}

		row.SortName = nil // making null because it is not necessary after this point

		var mds []*Mods
		for erows.Next() {
			erow := new(Mods)
			if err := erows.Scan(&erow.Id, &erow.Val); err != nil {
				fmt.Printf("item name mod error: %v", err)
			}
			erow = GetMods(erow)
			mds = append(mds, erow)
		}
		row.Modifs = mds

		arows, err := db.Query("SELECT q.itemId, q.name, q.level, q.ilevel, q.jobs, q.MId, q.shieldSize, q.scriptType, q.slot, q.rslot, q.race FROM item_equipment as q WHERE q.itemId = ?", row.ItemID)
		if err != nil {
			fmt.Println("error selecting equip for item")
			panic(err.Error())
		}
		defer arows.Close()

		var eq Equipment
		for arows.Next() {
			rowa := new(Equipment)
			if err := arows.Scan(&rowa.ItemID, &rowa.Name, &rowa.Level, &rowa.ILevel, &rowa.JobsInt, &rowa.Model, &rowa.ShieldSize, &rowa.Script, &rowa.Slot, &rowa.RSlot, &rowa.Race); err != nil {
				fmt.Printf("equipment job error: %v", err)
			}
			eq = *rowa
		}
		if eq.JobsInt > 0 { //leave armor null in struct unless it has real values
			row.Armor = &eq
			if row.ItemID >= 17860 && row.ItemID <= 17906 {
				//least bst item notes
			} else {
				row.Notes = ""
			}
			if row.Armor.JobsInt == 4194303 { // check for all jobs first
				row.Armor.Jobs = "All Jobs"
			} else {
				arr := GetPowersOfTwo(row.Armor.JobsInt)
				row.Armor.Jobs = GetJobsString(arr)
			}

			//get weapon data
			wrows, err := db.Query("SELECT * from item_weapon where itemId =?", row.ItemID)
			if err != nil {
				fmt.Println("error selecting weapon for item")
				panic(err.Error())
			}
			defer wrows.Close()

			for wrows.Next() {
				var wrow Weapon
				if err := wrows.Scan(&wrow.ItemID, &wrow.Name, &wrow.Skill, &wrow.SubSkill, &wrow.ILvlSkill, &wrow.ILvlParry, &wrow.ILvlMagic, &wrow.DmgType, &wrow.Hit, &wrow.Delay, &wrow.Dmg, &wrow.UnlockPoints, &wrow.Category); err != nil {
					fmt.Printf("weapons item error: %v", err)
				}
				row.Weap = &wrow
			}
		}
		items = append(items, row)

	}
	return items
}

func GetItemDetsByID(w http.ResponseWriter, r *http.Request) {
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

	rows, err := db.Query("SELECT q.dropType, q.groupId, q.groupRate, q.dropId, q.itemRate FROM mob_droplist as q WHERE q.itemId = ?", sID)
	if err != nil {
		fmt.Println("error selecting item drop dets")
		panic(err.Error())
	}
	defer rows.Close()

	var items ItemDets
	for rows.Next() {
		row := new(Drop)
		if err := rows.Scan(&row.DropType, &row.GroupId, &row.GroupRate, &row.ItemId, &row.ItemRate); err != nil {
			fmt.Printf("mob drop item error: %v", err)
		}
		//ItemId in drops is used for dropId to keep same struct as other functions
		erows, err := db.Query("SELECT q.name, q.zoneid, q.groupid, q.poolid, q.spawntype, a.name FROM mob_groups as q JOIN zone_settings as a ON q.zoneid = a.zoneid WHERE q.dropId = ?", row.ItemId)
		if err != nil {
			fmt.Println("error selecting mods for item name")
			panic(err.Error())
		}
		defer erows.Close()

		var mds *DropMob
		for erows.Next() {
			erow := new(DropMob)
			if err := erows.Scan(&erow.Name, &erow.ZoneId, &erow.GrpId, &erow.PoolId, &erow.SpawnType, &erow.ZoneName); err != nil {
				fmt.Printf("item id drop error: %v", err)
			}
			erow.Drop = row
			mds = erow
		}
		if mds != nil {
			items.Drops = append(items.Drops, mds)
		}
	}

	rowsr, err := db.Query("SELECT q.ID, q.Desynth, q.KeyItem, q.Wood, q.Smith, q.Gold, q.Cloth, q.Leather, q.Bone, q.Alchemy, q.Cook, q.Crystal, q.HQCrystal, q.Ingredient1, q.Ingredient2, q.Ingredient3, q.Ingredient4, q.Ingredient5, q.Ingredient6, q.Ingredient7, q.Ingredient8, q.Result, q.ResultHQ1, q.ResultHQ2, q.ResultHQ3, q.ResultQty, q.ResultHQ1Qty, q.ResultHQ2Qty, q.ResultHQ3Qty FROM synth_recipes as q WHERE q.Result = ? || q.ResultHQ1 = ? ||  q.ResultHQ2 = ? || q.ResultHQ3 = ? ORDER BY Desynth", sID, sID, sID, sID)
	if err != nil {
		fmt.Println("error selecting item recipe dets")
		panic(err.Error())
	}
	defer rowsr.Close()

	for rowsr.Next() {
		rrow := new(Recipe)
		var key int
		if err := rowsr.Scan(&rrow.ID, &rrow.Desynth, &key, &rrow.Wood, &rrow.Smith, &rrow.Gold, &rrow.Cloth, &rrow.Leather, &rrow.Bone, &rrow.Alchemy, &rrow.Cook, &rrow.Crystal, &rrow.HQCrys, &rrow.Ing1, &rrow.Ing2, &rrow.Ing3, &rrow.Ing4, &rrow.Ing5, &rrow.Ing6, &rrow.Ing7, &rrow.Ing8, &rrow.Result, &rrow.ResHQ1, &rrow.ResHQ2, &rrow.ResHQ3, &rrow.ResQty, &rrow.ResH1Qty, &rrow.ResH2Qty, &rrow.ResH3Qty); err != nil {
			fmt.Printf("item id recipe error: %v", err)
		}
		if key > 0 {
			rrow.KeyItem = KeyItems[strconv.Itoa(key)].Name
		}

		db.QueryRow("SELECT name FROM item_basic WHERE itemid = ?", rrow.Crystal).Scan(&rrow.CrystalName)
		if rrow.HQCrys != 0 {
			db.QueryRow("SELECT name FROM item_basic WHERE itemid = ?", rrow.HQCrys).Scan(&rrow.HQCrysName)
		}
		if rrow.Ing1 != 0 {
			db.QueryRow("SELECT name FROM item_basic WHERE itemid = ?", rrow.Ing1).Scan(&rrow.Ing1Name)
		}
		if rrow.Ing2 != 0 {
			db.QueryRow("SELECT name FROM item_basic WHERE itemid = ?", rrow.Ing2).Scan(&rrow.Ing2Name)
		}
		if rrow.Ing3 != 0 {
			db.QueryRow("SELECT name FROM item_basic WHERE itemid = ?", rrow.Ing3).Scan(&rrow.Ing3Name)
		}
		if rrow.Ing4 != 0 {
			db.QueryRow("SELECT name FROM item_basic WHERE itemid = ?", rrow.Ing4).Scan(&rrow.Ing4Name)
		}
		if rrow.Ing5 != 0 {
			db.QueryRow("SELECT name FROM item_basic WHERE itemid = ?", rrow.Ing5).Scan(&rrow.Ing5Name)
		}
		if rrow.Ing6 != 0 {
			db.QueryRow("SELECT name FROM item_basic WHERE itemid = ?", rrow.Ing6).Scan(&rrow.Ing6Name)
		}
		if rrow.Ing7 != 0 {
			db.QueryRow("SELECT name FROM item_basic WHERE itemid = ?", rrow.Ing7).Scan(&rrow.Ing7Name)
		}
		if rrow.Ing8 != 0 {
			db.QueryRow("SELECT name FROM item_basic WHERE itemid = ?", rrow.Ing8).Scan(&rrow.Ing8Name)
		}
		if rrow.ResHQ1 != 0 {
			db.QueryRow("SELECT name FROM item_basic WHERE itemid = ?", rrow.ResHQ1).Scan(&rrow.ResHQ1Name)
		}
		if rrow.ResHQ2 != 0 {
			db.QueryRow("SELECT name FROM item_basic WHERE itemid = ?", rrow.ResHQ2).Scan(&rrow.ResHQ1Name)
		}
		if rrow.ResHQ3 != 0 {
			db.QueryRow("SELECT name FROM item_basic WHERE itemid = ?", rrow.ResHQ3).Scan(&rrow.ResHQ1Name)
		}
		db.QueryRow("SELECT name FROM item_basic WHERE itemid = ?", rrow.Result).Scan(&rrow.ResultName)
		items.Recipes = append(items.Recipes, rrow)
	}

	jsonData, _ := json.Marshal(&items)
	w.Write(jsonData)
}

func GetArmorByJob(w http.ResponseWriter, r *http.Request) {
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
	numPow := IntPow(2, sID-1) // searchs for armor with only this job on it and proper af models
	rows, err := db.Query(fmt.Sprintf("SELECT q.itemId, a.name, q.level, q.ilevel, q.jobs, q.MId, q.shieldSize, q.scriptType, q.slot, q.rslot, q.race, a.flags FROM item_equipment as q JOIN item_basic as a ON q.itemId = a.itemid WHERE q.jobs = %d AND q.level <= %d AND q.MId IN %s ORDER BY q.level", numPow, lvlcap, afMap[sID]))
	if err != nil {
		fmt.Println("error selecting equip by job")
		panic(err.Error())
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		item := new(Item)
		row := new(Equipment)
		if err := rows.Scan(&item.ItemID, &item.Name, &row.Level, &row.ILevel, &row.JobsInt, &row.Model, &row.ShieldSize, &row.Script, &row.Slot, &row.RSlot, &row.Race, &item.Flags); err != nil {
			fmt.Printf("equipment job error: %v", err)
		}
		row.Jobs = jobsCapped[sID]

		erows, err := db.Query("SELECT q.modId, q.value FROM item_mods as q WHERE q.itemId = ?", item.ItemID)
		if err != nil {
			fmt.Println("error selecting mods for item name")
			panic(err.Error())
		}
		defer erows.Close()
		var mds []*Mods
		for erows.Next() {
			erow := new(Mods)
			if err := erows.Scan(&erow.Id, &erow.Val); err != nil {
				fmt.Printf("item name mod error: %v", err)
			}
			erow = GetMods(erow)
			mds = append(mds, erow)
		}
		item.Armor = row
		item.Modifs = mds
		items = append(items, item)
	}

	jsonData, _ := json.Marshal(&items)
	w.Write(jsonData)
}

type Mods struct {
	Id          int
	Val         int
	Str         string
	Weaponskill string
	Gmitem      bool
	IsMobMod    byte
}

func GetMods(mods *Mods) *Mods {
	mods.Str = ""
	temp := ""
	if mods.Val > 0 {
		temp = " +"
	}
	itemdmg := 5
	mods.Gmitem = false
	switch mods.Id {
	case 1:
		mods.Str = "DEF: " + strconv.Itoa(mods.Val)
	case 2:
		mods.Str = "HP" + temp + strconv.Itoa(mods.Val)
	case 3:
		mods.Str = "HP" + temp + strconv.Itoa(mods.Val) + "%"
	case 4:
		mods.Str = "Converts " + strconv.Itoa(mods.Val) + " MP to HP"
	case 5:
		mods.Str = "MP" + temp + strconv.Itoa(mods.Val)
	case 6:
		mods.Str = "MP" + temp + strconv.Itoa(mods.Val) + "%"
	case 7:
		mods.Str = "Converts " + strconv.Itoa(mods.Val) + " HP to MP"
	case 8:
		mods.Str = "STR" + temp + strconv.Itoa(mods.Val)
	case 9:
		mods.Str = "DEX" + temp + strconv.Itoa(mods.Val)
	case 10:
		mods.Str = "VIT" + temp + strconv.Itoa(mods.Val)
	case 11:
		mods.Str = "AGI" + temp + strconv.Itoa(mods.Val)
	case 12:
		mods.Str = "INT" + temp + strconv.Itoa(mods.Val)
	case 13:
		mods.Str = "MND" + temp + strconv.Itoa(mods.Val)
	case 14:
		mods.Str = "CHR" + temp + strconv.Itoa(mods.Val)
	case 15:
		mods.Str = "[Element: Fire]" + temp + strconv.Itoa(mods.Val)
	case 16:
		mods.Str = "[Element: Ice]" + temp + strconv.Itoa(mods.Val)
	case 17:
		mods.Str = "[Element: Wind]" + temp + strconv.Itoa(mods.Val)
	case 18:
		mods.Str = "[Element: Earth]" + temp + strconv.Itoa(mods.Val)
	case 19:
		mods.Str = "[Element: Lightning]" + temp + strconv.Itoa(mods.Val)
	case 20:
		mods.Str = "[Element: Water]" + temp + strconv.Itoa(mods.Val)
	case 21:
		mods.Str = "[Element: Light]" + temp + strconv.Itoa(mods.Val)
	case 22:
		mods.Str = "[Element: Dark]" + temp + strconv.Itoa(mods.Val)
	case 23:
		mods.Str = "Attack" + temp + strconv.Itoa(mods.Val)
	case 24:
		mods.Str = "Ranged Attack" + temp + strconv.Itoa(mods.Val)
	case 25:
		mods.Str = "Accuracy" + temp + strconv.Itoa(mods.Val)
	case 26:
		mods.Str = "Ranged Accuracy" + temp + strconv.Itoa(mods.Val)
	case 27:
		mods.Str = "Enmity" + temp + strconv.Itoa(mods.Val)
	case 28:
		mods.Str = "\"Magic Atk. Bonus\"" + temp + strconv.Itoa(mods.Val)
	case 29:
		mods.Str = "\"Magic Def. Bonus\"" + temp + strconv.Itoa(mods.Val)
	case 30:
		mods.Str = "Magic Accuracy" + temp + strconv.Itoa(mods.Val)
	case 31:
		mods.Str = "Magic Evasion" + temp + strconv.Itoa(mods.Val)
	case 32:
		mods.Str = "Fire elemental \"Magic Atk. Bonus\"" + temp + strconv.Itoa(mods.Val)
	case 33:
		mods.Str = "Ice elemental \"Magic Atk. Bonus\"" + temp + strconv.Itoa(mods.Val)
	case 34:
		mods.Str = "Wind elemental \"Magic Atk. Bonus\"" + temp + strconv.Itoa(mods.Val)
	case 35:
		mods.Str = "Earth elemental \"Magic Atk. Bonus\"" + temp + strconv.Itoa(mods.Val)
	case 36:
		mods.Str = "Thunder elemental \"Magic Atk. Bonus\"" + temp + strconv.Itoa(mods.Val)
	case 37:
		mods.Str = "Water elemental \"Magic Atk. Bonus\"" + temp + strconv.Itoa(mods.Val)
	case 38:
		mods.Str = "Light elemental \"Magic Atk. Bonus\"" + temp + strconv.Itoa(mods.Val)
	case 39:
		mods.Str = "Dark elemental \"Magic Atk. Bonus\"" + temp + strconv.Itoa(mods.Val)
	case 40:
		mods.Str = "Fire elemental Magic Accuracy" + temp + strconv.Itoa(mods.Val)
	case 41:
		mods.Str = "Ice elemental Magic Accuracy" + temp + strconv.Itoa(mods.Val)
	case 42:
		mods.Str = "Wind elemental Magic Accuracy" + temp + strconv.Itoa(mods.Val)
	case 43:
		mods.Str = "Earth elemental Magic Accuracy" + temp + strconv.Itoa(mods.Val)
	case 44:
		mods.Str = "Thunder elemental Magic Accuracy" + temp + strconv.Itoa(mods.Val)
	case 45:
		mods.Str = "Water elemental Magic Accuracy" + temp + strconv.Itoa(mods.Val)
	case 46:
		mods.Str = "Light elemental Magic Accuracy" + temp + strconv.Itoa(mods.Val)
	case 47:
		mods.Str = "Dark elemental Magic Accuracy" + temp + strconv.Itoa(mods.Val)
	case 48:
		mods.Str = "Weapon Skill Accuracy" + temp + strconv.Itoa(mods.Val)
	case 49:
		mods.Str = "Slashing Resistance" + temp + strconv.Itoa(mods.Val) + "%"
	case 50:
		mods.Str = "Piercing Resistance" + temp + strconv.Itoa(mods.Val) + "%"
	case 51:
		mods.Str = "Impact Resistance" + temp + strconv.Itoa(mods.Val) + "%"
	case 52:
		mods.Str = "Hand-to-hand Resistance" + temp + strconv.Itoa(mods.Val) + "%"
	case 54:
		mods.Str = "<img src=\"../../../assets/imgs/fire.png\" />" + temp + strconv.Itoa(mods.Val)
	case 55:
		mods.Str = "<img src=\"../../../assets/imgs/ice.png\" />" + temp + strconv.Itoa(mods.Val)
	case 56:
		mods.Str = "<img src=\"../../../assets/imgs/wind.png\" />" + temp + strconv.Itoa(mods.Val)
	case 57:
		mods.Str = "<img src=\"../../../assets/imgs/earth.png\" />" + temp + strconv.Itoa(mods.Val)
	case 58:
		mods.Str = "<img src=\"../../../assets/imgs/lightning.png\" /> " + temp + strconv.Itoa(mods.Val)
	case 59:
		mods.Str = "<img src=\"../../../assets/imgs/water.png\" /> " + temp + strconv.Itoa(mods.Val)
	case 60:
		mods.Str = "<img src=\"../../../assets/imgs/light.png\" />" + temp + strconv.Itoa(mods.Val)
	case 61:
		mods.Str = "<img src=\"../../../assets/imgs/dark.png\" />" + temp + strconv.Itoa(mods.Val)
	case 1054:
		mods.Str = "Specific Damage Taken (Fire)" + temp + strconv.Itoa(mods.Val) + "%"
	case 1055:
		mods.Str = "Specific Damage Taken (Ice)" + temp + strconv.Itoa(mods.Val) + "%"
	case 1056:
		mods.Str = "Specific Damage Taken (Wind)" + temp + strconv.Itoa(mods.Val) + "%"
	case 1057:
		mods.Str = "Specific Damage Taken (Earth)" + temp + strconv.Itoa(mods.Val) + "%"
	case 1058:
		mods.Str = "Specific Damage Taken (Thunder)" + temp + strconv.Itoa(mods.Val) + "%"
	case 1059:
		mods.Str = "Specific Damage Taken (Water)" + temp + strconv.Itoa(mods.Val) + "%"
	case 1060:
		mods.Str = "Specific Damage Taken (Light)" + temp + strconv.Itoa(mods.Val) + "%"
	case 1061:
		mods.Str = "Specific Damage Taken (Dark)" + temp + strconv.Itoa(mods.Val) + "%"
	case 62:
		mods.Str = "Attack" + temp + strconv.Itoa(mods.Val) + "%"
	case 63:
		mods.Str = "Defense" + temp + strconv.Itoa(mods.Val) + "%"
	case 64:
		mods.Str = "Combat skillup rate" + temp + strconv.Itoa(mods.Val) + "%"
	case 65:
		mods.Str = "Magic skillup rate" + temp + strconv.Itoa(mods.Val) + "%"
	case 66:
		mods.Str = "Ranged Attack" + temp + strconv.Itoa(mods.Val) + "%"
	case 68:
		mods.Str = "Evasion" + temp + strconv.Itoa(mods.Val)
	case 69:
		mods.Str = "Ranged Defense" + temp + strconv.Itoa(mods.Val)
	case 70:
		mods.Str = "Ranged Evasion" + temp + strconv.Itoa(mods.Val)
	case 71:
		mods.Str = "MP recovered while healing " + temp + strconv.Itoa(mods.Val)
	case 72:
		mods.Str = "HP recovered while healing " + temp + strconv.Itoa(mods.Val)
	case 73:
		mods.Str = "\"Store TP\"" + temp + strconv.Itoa(mods.Val)
	case 486:
		mods.Str = "\"Tactical Parry\" TP Bonus" + temp + strconv.Itoa(mods.Val)
	case 487:
		mods.Str = "Magic Burst Bonus Modifier" + temp + strconv.Itoa(mods.Val) + "%"
	case 488:
		mods.Str = "Inhibits TP Gain" + temp + strconv.Itoa(mods.Val) + "%"
	case 80:
		mods.Str = "Hand-to-Hand skill" + temp + strconv.Itoa(mods.Val)
	case 81:
		mods.Str = "Dagger skill" + temp + strconv.Itoa(mods.Val)
	case 82:
		mods.Str = "Sword skill" + temp + strconv.Itoa(mods.Val)
	case 83:
		mods.Str = "Great Sword skill" + temp + strconv.Itoa(mods.Val)
	case 84:
		mods.Str = "Axe skill" + temp + strconv.Itoa(mods.Val)
	case 85:
		mods.Str = "Great Axe skill" + temp + strconv.Itoa(mods.Val)
	case 86:
		mods.Str = "Scythe skill" + temp + strconv.Itoa(mods.Val)
	case 87:
		mods.Str = "Polearm skill" + temp + strconv.Itoa(mods.Val)
	case 88:
		mods.Str = "Katana skill" + temp + strconv.Itoa(mods.Val)
	case 89:
		mods.Str = "Great Katana skill" + temp + strconv.Itoa(mods.Val)
	case 90:
		mods.Str = "Club skill" + temp + strconv.Itoa(mods.Val)
	case 91:
		mods.Str = "Staff skill" + temp + strconv.Itoa(mods.Val)
	case 101:
		mods.Str = "Automation melee skill" + temp + strconv.Itoa(mods.Val)
	case 102:
		mods.Str = "Automation range skill" + temp + strconv.Itoa(mods.Val)
	case 103:
		mods.Str = "Automation magic skill" + temp + strconv.Itoa(mods.Val)
	case 104:
		mods.Str = "Archery skill" + temp + strconv.Itoa(mods.Val)
	case 105:
		mods.Str = "Marksmanship skill" + temp + strconv.Itoa(mods.Val)
	case 106:
		mods.Str = "Throwing skill" + temp + strconv.Itoa(mods.Val)
	case 107:
		mods.Str = "Guard skill" + temp + strconv.Itoa(mods.Val)
	case 108:
		mods.Str = "Evasion skill" + temp + strconv.Itoa(mods.Val)
	case 109:
		mods.Str = "Shield skill" + temp + strconv.Itoa(mods.Val)
	case 110:
		mods.Str = "Parrying skill" + temp + strconv.Itoa(mods.Val)
	case 111:
		mods.Str = "Divine magic skill" + temp + strconv.Itoa(mods.Val)
	case 112:
		mods.Str = "Healing magic skill" + temp + strconv.Itoa(mods.Val)
	case 113:
		mods.Str = "Enhancing magic skill" + temp + strconv.Itoa(mods.Val)
	case 114:
		mods.Str = "Enfeebling magic skill" + temp + strconv.Itoa(mods.Val)
	case 115:
		mods.Str = "Elemental magic skill" + temp + strconv.Itoa(mods.Val)
	case 116:
		mods.Str = "Dark magic skill" + temp + strconv.Itoa(mods.Val)
	case 117:
		mods.Str = "Summoning magic skill" + temp + strconv.Itoa(mods.Val)
	case 118:
		mods.Str = "Ninjutsu skill" + temp + strconv.Itoa(mods.Val)
	case 119:
		mods.Str = "Singing skill" + temp + strconv.Itoa(mods.Val)
	case 120:
		mods.Str = "String skill" + temp + strconv.Itoa(mods.Val)
	case 121:
		mods.Str = "Wind skill" + temp + strconv.Itoa(mods.Val)
	case 122:
		mods.Str = "Blue magic skill" + temp + strconv.Itoa(mods.Val)
	case 125:
		mods.Str = "Suppresses \"Overload\""
	case 127:
		mods.Str = "Fishing skill" + temp + strconv.Itoa(mods.Val)
	case 128:
		mods.Str = "Woodworking skill" + temp + strconv.Itoa(mods.Val)
	case 129:
		mods.Str = "Smithing skill" + temp + strconv.Itoa(mods.Val)
	case 130:
		mods.Str = "Goldsmithing skill" + temp + strconv.Itoa(mods.Val)
	case 131:
		mods.Str = "Clothcraft skill" + temp + strconv.Itoa(mods.Val)
	case 132:
		mods.Str = "Leathercraft skill" + temp + strconv.Itoa(mods.Val)
	case 133:
		mods.Str = "Bonecraft skill" + temp + strconv.Itoa(mods.Val)
	case 134:
		mods.Str = "Alchemy skill" + temp + strconv.Itoa(mods.Val)
	case 135:
		mods.Str = "Cooking skill" + temp + strconv.Itoa(mods.Val)
	case 136:
		mods.Str = "Synergy skill" + temp + strconv.Itoa(mods.Val)
	case 137:
		mods.Str = "Riding skill" + temp + strconv.Itoa(mods.Val)
	case 144:
		mods.Str = "Woodworking Success Rate" + temp + strconv.Itoa(mods.Val) + "%"
	case 145:
		mods.Str = "Smithing Success Rate" + temp + strconv.Itoa(mods.Val) + "%"
	case 146:
		mods.Str = "Goldsmithing Success Rate" + temp + strconv.Itoa(mods.Val) + "%"
	case 147:
		mods.Str = "Clothcraft Success Rate" + temp + strconv.Itoa(mods.Val) + "%"
	case 148:
		mods.Str = "Leathercraft Success Rate" + temp + strconv.Itoa(mods.Val) + "%"
	case 149:
		mods.Str = "Bonecraft Success Rate" + temp + strconv.Itoa(mods.Val) + "%"
	case 150:
		mods.Str = "Alchemy Success Rate" + temp + strconv.Itoa(mods.Val) + "%"
	case 151:
		mods.Str = "Cooking Success Rate" + temp + strconv.Itoa(mods.Val) + "%"
	case 160:
		mods.Str = "Damage taken " + temp + strconv.Itoa(mods.Val) + "%"
	case 161:
		mods.Str = "Physical damage taken " + temp + strconv.Itoa(mods.Val) + "%"
	case 190:
		mods.Str = "Physical damage taken II " + temp + strconv.Itoa(mods.Val) + "%"
	case 162:
		mods.Str = "Breath damage taken " + temp + strconv.Itoa(mods.Val) + "%"
	case 163:
		mods.Str = "Magic damage taken " + temp + strconv.Itoa(mods.Val) + "%"
	case 831:
		mods.Str = "Magic damage taken II " + temp + strconv.Itoa(mods.Val) + "%"
	case 164:
		mods.Str = "Ranged damage taken " + temp + strconv.Itoa(mods.Val) + "%"
	case 387:
		mods.Gmitem = true
		mods.Str = "Uncapped physical damage multipliers"
	case 388:
		mods.Gmitem = true
		mods.Str = "Uncapped breath damage multipliers"
	case 389:
		mods.Gmitem = true
		mods.Str = "Uncapped magic damage multipliers"
	case 390:
		mods.Gmitem = true
		mods.Str = "Uncapped ranged damage multipliers"
	case 165:
		mods.Str = "Critical hit rate " + temp + strconv.Itoa(mods.Val) + "%"
	case 421:
		mods.Str = "Critical hit damage" + temp + strconv.Itoa(mods.Val) + "%"
	case 964:
		mods.Str = "Ranged critical damage" + temp + strconv.Itoa(mods.Val) + "%"
	case 166:
		mods.Str = "Enemy critical hit rate " + temp + strconv.Itoa(mods.Val) + "%"
	case 908:
		mods.Str = "Critical defense bonus" + temp + strconv.Itoa(mods.Val) + "%"
	case 562:
		mods.Str = "Magic critical hit rate " + temp + strconv.Itoa(mods.Val) + "%"
	case 563:
		mods.Str = "Increases magic critical hit damage" + temp + strconv.Itoa(mods.Val) + "%"
	case 898:
		mods.Str = "Enhances \"Smite\" effect" + temp + strconv.Itoa(mods.Val)
	case 899:
		mods.Str = "Enhances \"Tactical Guard\" effect" + temp + strconv.Itoa(mods.Val)
	case 901:
		mods.Str = "Quickens Elemental Magic Casting" + temp + strconv.Itoa(mods.Val) + "%"
	case 903:
		mods.Str = "\"Fencer\" TP Bonus" + temp + strconv.Itoa(mods.Val)
	case 904:
		mods.Str = "\"Fencer\" Critical Rate" + temp + strconv.Itoa(mods.Val)
	case 976:
		mods.Str = "\"Guard\"" + temp + strconv.Itoa(mods.Val) + "%"
	case 167:
		mods.Str = fmt.Sprintf("Magic Haste %.1f", math.Round(float64(mods.Val/100))) + "%"
	case 383:
		mods.Str = fmt.Sprintf("Ability Haste %.1f", math.Round(float64(mods.Val/100))) + "%"
	case 384:
		mods.Str = fmt.Sprintf("Gear Haste %.1f", math.Round(float64(mods.Val/100))) + "%"
	case 168:
		mods.Str = "Spell interruption rate down " + temp + strconv.Itoa(mods.Val) + "%"
	case 169:
		mods.Str = "Movement speed " + temp + strconv.Itoa(mods.Val) + "%"
	case 170:
		mods.Str = "Enhances \"Fast Cast\" effect"
	case 407:
		mods.Str = "\"Fast Cast\"" + temp + strconv.Itoa(mods.Val) + "%"
	case 394:
		mods.Str = "Elem. magic casting time -" + strconv.Itoa(Abs(mods.Val)) + "%"
	case 519:
		mods.Str = "\"Cure\" spellcasting time -" + strconv.Itoa(Abs(mods.Val)) + "%"
	case 171:
		mods.Str = "Two-handed weapon delay " + temp + strconv.Itoa(mods.Val) + "%"
	case 172:
		mods.Str = "Ranged weapon delay " + temp + strconv.Itoa(mods.Val) + "%"
	case 173:
		mods.Str = "Enhances \"Martial Arts\" effect " + temp + strconv.Itoa(mods.Val)
	case 174:
		mods.Str = "\"Skillchain Bonus\"" + temp + strconv.Itoa(mods.Val)
	case 175:
		mods.Str = "\"Skillchain Damage\"" + temp + strconv.Itoa(mods.Val)
	case 978:
		mods.Str = "Max occ. attacks twice swings = " + strconv.Itoa(mods.Val)
	case 979:
		mods.Str = "Chance for additional swing " + strconv.Itoa(mods.Val) + "%"
	case 311:
		mods.Str = "Magic Damage" + temp + strconv.Itoa(mods.Val)
	case 176:
		mods.Str = "Food HP" + temp + strconv.Itoa(mods.Val)
	case 177:
		mods.Str = "Food HP Cap = " + strconv.Itoa(mods.Val)
	case 178:
		mods.Str = "Food MP" + temp + strconv.Itoa(mods.Val)
	case 179:
		mods.Str = "Food MP Cap = " + strconv.Itoa(mods.Val)
	case 180:
		mods.Str = "Food ATT" + temp + strconv.Itoa(mods.Val)
	case 181:
		mods.Str = "Food ATT Cap = " + strconv.Itoa(mods.Val)
	case 182:
		mods.Str = "Food DEF" + temp + strconv.Itoa(mods.Val)
	case 183:
		mods.Str = "Food DEF Cap = " + strconv.Itoa(mods.Val)
	case 184:
		mods.Str = "Food ACC" + temp + strconv.Itoa(mods.Val)
	case 185:
		mods.Str = "Food ACC Cap = " + strconv.Itoa(mods.Val)
	case 186:
		mods.Str = "Food RATT" + temp + strconv.Itoa(mods.Val)
	case 187:
		mods.Str = "Food RATT Cap = " + strconv.Itoa(mods.Val)
	case 188:
		mods.Str = "Food RACC" + temp + strconv.Itoa(mods.Val)
	case 189:
		mods.Str = "Food RACC Cap = " + strconv.Itoa(mods.Val)
	case 99:
		mods.Str = "Food MACC" + temp + strconv.Itoa(mods.Val)
	case 100:
		mods.Str = "Food MACC Cap = " + strconv.Itoa(mods.Val)
	case 937:
		mods.Str = "Food Duration" + temp + strconv.Itoa(mods.Val) + "%"
	case 224:
		mods.Str = "Enhances \"Vermin Killer\" effect"
	case 225:
		mods.Str = "Enhances \"Bird Killer\" effect"
	case 226:
		mods.Str = "Enhances \"Amorph Killer\" effect"
	case 227:
		mods.Str = "Enhances \"Lizard Killer\" effect"
	case 228:
		mods.Str = "Enhances \"Aquan Killer\" effect"
	case 229:
		mods.Str = "Enhances \"Plantoid Killer\" effect"
	case 230:
		mods.Str = "Enhances \"Beast Killer\" effect"
	case 231:
		mods.Str = "Enhances \"Undead Killer\" effect"
	case 232:
		mods.Str = "Enhances \"Arcana Killer\" effect"
	case 233:
		mods.Str = "Enhances \"Dragon Killer\" effect"
	case 234:
		mods.Str = "Enhances \"Demon Killer\" effect"
	case 235:
		mods.Str = "Enhances \"Empty Killer\" effect"
	case 236:
		mods.Str = "Enhances \"Humanoid Killer\" effect"
	case 237:
		mods.Str = "Enhances \"Lumorian Killer\" effect"
	case 238:
		mods.Str = "Enhances \"Luminion Killer\" effect"
	case 239:
		mods.Str = "Resistance to All Status Ailments"
	case 240:
		mods.Str = "Enhances \"Resist Sleep\" effect"
	case 241:
		mods.Str = "Enhances \"Resist Poison\" effect"
	case 242:
		mods.Str = "Enhances \"Resist Paralyze\" effect"
	case 243:
		mods.Str = "Enhances \"Resist Blind\" effect"
	case 244:
		mods.Str = "Enhances \"Resist Silence\" effect"
	case 245:
		mods.Str = "Enhances \"Resist Virus\" effect"
	case 246:
		mods.Str = "Enhances \"Resist Petrify\" effect"
	case 247:
		mods.Str = "Enhances \"Resist Bind\" effect"
	case 248:
		mods.Str = "Enhances \"Resist Curse\" effect"
	case 249:
		mods.Str = "Enhances \"Resist Gravity\" effect"
	case 250:
		mods.Str = "Enhances \"Resist Slow\" effect"
	case 251:
		mods.Str = "Enhances \"Resist Stun\" effect"
	case 252:
		mods.Str = "Enhances \"Resist Charm\" effect"
	case 253:
		mods.Str = "Enhances \"Resist Amnesia\" effect"
	case 254:
		mods.Str = "Enhances \"Resist Lullaby\" effect"
	case 255:
		mods.Str = "Enhances resistance against \"Death\""
	case 1240:
		mods.Str = "\"Resist Sleep\"" + temp + strconv.Itoa(mods.Val)
	case 1241:
		mods.Str = "\"Resist Poison\"" + temp + strconv.Itoa(mods.Val)
	case 1242:
		mods.Str = "\"Resist Paralyze\"" + temp + strconv.Itoa(mods.Val)
	case 1243:
		mods.Str = "\"Resist Blind\"" + temp + strconv.Itoa(mods.Val)
	case 1244:
		mods.Str = "\"Resist Silence\"" + temp + strconv.Itoa(mods.Val)
	case 1245:
		mods.Str = "\"Resist Virus\"" + temp + strconv.Itoa(mods.Val)
	case 1246:
		mods.Str = "\"Resist Petrify\"" + temp + strconv.Itoa(mods.Val)
	case 1247:
		mods.Str = "\"Resist Bind\"" + temp + strconv.Itoa(mods.Val)
	case 1248:
		mods.Str = "\"Resist Curse\"" + temp + strconv.Itoa(mods.Val)
	case 1249:
		mods.Str = "\"Resist Gravity\"" + temp + strconv.Itoa(mods.Val)
	case 1250:
		mods.Str = "\"Resist Slow\"" + temp + strconv.Itoa(mods.Val)
	case 1251:
		mods.Str = "\"Resist Stun\"" + temp + strconv.Itoa(mods.Val)
	case 1252:
		mods.Str = "\"Resist Charm\"" + temp + strconv.Itoa(mods.Val)
	case 1253:
		mods.Str = "\"Resist Amnesia\"" + temp + strconv.Itoa(mods.Val)
	case 1254:
		mods.Str = "\"Resist Lullaby\"" + temp + strconv.Itoa(mods.Val)
	case 1255:
		mods.Str = "\"Resist Death\"" + temp + strconv.Itoa(mods.Val)
	case 2002:
		mods.Str = "\"Build Sleep Resistance\"" + temp + strconv.Itoa(mods.Val)
	case 2003:
		mods.Str = "\"Build Poison Resistance\"" + temp + strconv.Itoa(mods.Val)
	case 2004:
		mods.Str = "\"Build Paralyze Resistance\"" + temp + strconv.Itoa(mods.Val)
	case 2005:
		mods.Str = "\"Build Blind Resistance\"" + temp + strconv.Itoa(mods.Val)
	case 2006:
		mods.Str = "\"Build Silence Resistance\"" + temp + strconv.Itoa(mods.Val)
	case 2010:
		mods.Str = "\"Build Stun Resistance\"" + temp + strconv.Itoa(mods.Val)
	case 2011:
		mods.Str = "\"Build Bind Resistance\"" + temp + strconv.Itoa(mods.Val)
	case 2012:
		mods.Str = "\"Build Gravity Resistance\"" + temp + strconv.Itoa(mods.Val)
	case 2013:
		mods.Str = "\"Build Slow Resistance\"" + temp + strconv.Itoa(mods.Val)
	case 2193:
		mods.Str = "\"Build Lullaby Resistance\"" + temp + strconv.Itoa(mods.Val)
	case 256: //Relics and Mythics
		/*		switch mods.Val {
				case 2:
					mods.Str = "Additional Effect: Poison"
				case 3:
					mods.Str = "Additional Effect: Damage proportionate to current HP"
				case 5:
					mods.Str = "Additional Effect: \"Choke\""
				case 6:
					mods.Str = "Additional Effect: Impairs evasion"
				case 7:
					mods.Str = "Additional Effect: Blindness"
				case 8:
					mods.Str = "Additional Effect: Weakens defense"
				case 9:
					mods.Str = "Additional Effect: Paralysis"
				case 10:
					mods.Str = "Additional Effect: Weakens attacks"
				case 11:
					mods.Str = "Additional Effect: Recover MP"
				case 12:
					mods.Str = "Additional Effect: Dispel"
				case 29:
					mods.Str = "Aftermath: Increases Acc./Att. Occasionally Attacks Twice"
				case 30:
					mods.Str = "Aftermath: Increases Magic Acc./Acc. Occasionally Attacks Twice"
				case 31:
					mods.Str = "Aftermath: Increases Magic Acc./Att. Occasionally Attacks Twice"
				case 32:
					mods.Str = "Aftermath: Increases Acc./Magic Acc. Occasionally Attacks Twice"
				case 33:
					mods.Str = "Aftermath: Inc. Rng. Acc./Rng. Atk. Occasionally deals double damage"
				} */
	case 257:
		mods.Str = "Paralyze Proc Chance" + temp + strconv.Itoa(mods.Val) + "%"
	case 258:
		mods.Str = "Augments \"Mijin Gakure\" Effect"
	case 259:
		mods.Str = "Enhances \"Dual Wield\" effect " + temp + strconv.Itoa(mods.Val) + "%"
	case 191: //191 and 834 may need to be swapped
		mods.Str = "Enhances \"Quick Draw\" Effect damage" + temp + strconv.Itoa(mods.Val) + "%"
	case 834:
		mods.Str = "Enhances \"Quick Draw\" Effect acc." + temp + strconv.Itoa(mods.Val) + "%"
	case 288:
		mods.Str = "\"Double Attack\"" + temp + strconv.Itoa(mods.Val) + "%"
	case 483:
		mods.Str = "Enhances \"Warcry\" Duration" + temp + strconv.Itoa(mods.Val)
	case 948:
		mods.Str = "Enhances \"Berserk\" Effect" + temp + strconv.Itoa(mods.Val)
	case 954:
		mods.Str = "Enhances \"Berserk\" Duration" + temp + strconv.Itoa(mods.Val)
	case 955:
		mods.Str = "Enhances \"Aggressor\" Duration" + temp + strconv.Itoa(mods.Val)
	case 956:
		mods.Str = "Enhances \"Defender\" Duration" + temp + strconv.Itoa(mods.Val)
	case 97:
		mods.Str = fmt.Sprintf("Enhances \"Boost\" effect %.1f ", math.Round(float64(mods.Val/10))) + "%"
	case 123:
		mods.Str = "\"Chakra\" effect multiplier" + temp + strconv.Itoa(mods.Val)
	case 124:
		mods.Str = "Extra statuses removed by \"Chakra\""
	case 289:
		mods.Str = "\"Subtle Blow\"" + temp + strconv.Itoa(mods.Val)
	case 291:
		mods.Str = "\"Counter\"" + temp + strconv.Itoa(mods.Val)
	case 292:
		mods.Str = "\"Kick Attacks\"" + temp + strconv.Itoa(mods.Val)
	case 428:
		mods.Str = "Increases \"Perfect Counter\" attack " + temp + strconv.Itoa(mods.Val)
	case 429:
		mods.Str = "Enhances \"Footwork\" effect " + temp + strconv.Itoa(mods.Val)
	case 543:
		mods.Str = "Enhances \"Counterstance\" effect " + temp + strconv.Itoa(mods.Val) + "%"
	case 552:
		mods.Str = "Enhances \"Dodge\" effect " + temp + strconv.Itoa(mods.Val) + "%"
	case 561:
		mods.Str = "Enhances \"Focus\" effect " + temp + strconv.Itoa(mods.Val) + "%"
	case 293:
		mods.Str = "\"Afflatus Solace\"" + temp + strconv.Itoa(mods.Val)
	case 294:
		mods.Str = "\"Afflatus Misery\"" + temp + strconv.Itoa(mods.Val)
	case 484:
		mods.Str = "Enhances \"Auspice\" effect " + temp + strconv.Itoa(mods.Val)
	case 524:
		mods.Str = "Enhances \"Divine Veil\" effect"
	case 838:
		mods.Str = "Enhances \"Regen\" potency " + temp + strconv.Itoa(mods.Val) + "%"
	case 860:
		mods.Str = "Converts \"Cure\" Amount to MP" + temp + strconv.Itoa(mods.Val) + "%"
	case 910:
		mods.Str = "-Na spell \"Fast Cast\" and enmity reduction"
	case 295:
		mods.Str = "Increases MP Healing" + temp + strconv.Itoa(mods.Val)
	case 296:
		mods.Str = "\"Conserve MP\"" + temp + strconv.Itoa(mods.Val) + "%"
	case 297:
		mods.Str = "Enhances \"Saboteur\" Potency" + temp + strconv.Itoa(mods.Val) + "%"
	case 299:
		mods.Str = "Tracks \"Blink\" Shadows"
	case 300:
		if mods.Val < 0 {
			mods.Str = "\"Stoneskin\" casting time " + temp + strconv.Itoa(mods.Val) + "%"
		} else {
			mods.Str = "Enhances \"Stoneskin\" effect " + temp + strconv.Itoa(mods.Val) + "%"
		}
	case 301:
		mods.Str = "Tracks \"Phalanx\" damage reduction"
	case 290:
		mods.Str = "Enhances \"Enfeebling Magic\" Potency" + temp + strconv.Itoa(mods.Val) + "%"
	case 93:
		mods.Str = "Increases \"Flee\" duration " + temp + strconv.Itoa(mods.Val)
	case 298:
		mods.Str = "\"Steal\"" + temp + strconv.Itoa(mods.Val)
	case 896:
		mods.Str = "Increases \"Despoil\" Chance" + temp + strconv.Itoa(mods.Val) + "%"
	case 883:
		mods.Str = "Increases \"Perfect Dodge\" duration" + temp + strconv.Itoa(mods.Val)
	case 302:
		mods.Str = "\"Triple Attack\"" + temp + strconv.Itoa(mods.Val) + "%"
	case 303:
		mods.Str = "\"Treasure Hunter\"" + temp + strconv.Itoa(mods.Val)
	case 959:
		mods.Str = "Increases \"Sneak Attack\" damage" + temp + strconv.Itoa(mods.Val) + "%"
	case 520:
		mods.Str = "Increases \"Trick Attack\" damage " + temp + strconv.Itoa(mods.Val) + "%"
	case 835:
		mods.Str = "Enhances \"Accomplice/Collaborator\" effect " + temp + strconv.Itoa(mods.Val)
	case 884:
		mods.Str = "Enhances \"Saboteur\" Potency" + temp + strconv.Itoa(mods.Val) + "%"
	case 885:
		mods.Str = "Increases \"Hide\" duration" + temp + strconv.Itoa(mods.Val)
	case 897:
		mods.Str = "\"Gilfinder\""
	case 857:
		mods.Str = "Increases \"Holy Circle\" duration" + temp + strconv.Itoa(mods.Val)
	case 92:
		mods.Str = "Increases \"Rampart\" duration" + temp + strconv.Itoa(mods.Val)
	case 426:
		mods.Str = "Physical dmg taken to MP" + temp + strconv.Itoa(mods.Val)
	case 485:
		mods.Str = "\"Shield Mastery\" TP Bonus" + temp + strconv.Itoa(mods.Val)
	case 905:
		mods.Str = "\"Shield Defense Bonus\"" + temp + strconv.Itoa(mods.Val)
	case 965:
		mods.Str = "Convert \"Cover\" dmg taken to MP" + temp + strconv.Itoa(mods.Val)
	case 966:
		mods.Str = "Absorb ranged/magic using \"Cover\"" + temp + strconv.Itoa(mods.Val)
	case 967:
		mods.Str = "Increases \"Cover\" duration" + temp + strconv.Itoa(mods.Val)
	case 516:
		mods.Str = "Converts " + strconv.Itoa(mods.Val) + "% of physical damage taken to MP"
	case 427:
		mods.Str = "Reduces Enmity decrease when taking damage -" + strconv.Itoa(Abs(mods.Val))
	case 837:
		mods.Str = "Enhances \"Sentinel\" effect " + temp + strconv.Itoa(mods.Val) + "%"
	case 858:
		mods.Str = "\"Arcane Circle\" duration" + temp + strconv.Itoa(mods.Val)
	case 96:
		mods.Str = "Enhances \"Souleater\" effect " + temp + strconv.Itoa(mods.Val) + "%"
	case 906:
		mods.Str = "Ability Haste to \"Last Resort\"" + temp + strconv.Itoa(mods.Val/100)
	case 907:
		mods.Str = "Damage taken from \"Souleater\"" + temp + strconv.Itoa(mods.Val)
	case 304:
		mods.Str = "\"Tame\"" + temp + strconv.Itoa(mods.Val)
	case 360:
		mods.Str = "Increases \"Charm\" duration " + temp + strconv.Itoa(mods.Val)
	case 364:
		mods.Str = "Enhances \"Reward\" effect " + temp + strconv.Itoa(mods.Val) + "%"
	case 391:
		mods.Str = "\"Charm\" Chance" + temp + strconv.Itoa(mods.Val) + "%"
	case 503:
		mods.Str = "\"Feral Howl\" duration" + temp + strconv.Itoa(mods.Val)
	case 564:
		mods.Str = "Augments \"Call Beast\" Level" + temp + strconv.Itoa(mods.Val)
	case 433:
		mods.Str = "\"Minne\"" + temp + strconv.Itoa(mods.Val)
	case 434:
		mods.Str = "\"Minuet\"" + temp + strconv.Itoa(mods.Val)
	case 435:
		mods.Str = "\"Paeon\"" + temp + strconv.Itoa(mods.Val)
	case 436:
		mods.Str = "\"Requiem\"" + temp + strconv.Itoa(mods.Val)
	case 437:
		mods.Str = "\"Threnody\"" + temp + strconv.Itoa(mods.Val)
	case 438:
		mods.Str = "\"Madrigal\"" + temp + strconv.Itoa(mods.Val)
	case 439:
		mods.Str = "\"Mambo\"" + temp + strconv.Itoa(mods.Val)
	case 440:
		mods.Str = "\"Lullaby\"" + temp + strconv.Itoa(mods.Val)
	case 441:
		mods.Str = "\"Etude\"" + temp + strconv.Itoa(mods.Val)
	case 442:
		mods.Str = "\"Ballad\"" + temp + strconv.Itoa(mods.Val)
	case 443:
		mods.Str = "\"March\"" + temp + strconv.Itoa(mods.Val)
	case 444:
		mods.Str = "\"Finale\"" + temp + strconv.Itoa(mods.Val)
	case 445:
		mods.Str = "\"Carol\"" + temp + strconv.Itoa(mods.Val)
	case 446:
		mods.Str = "\"Mazurka\"" + temp + strconv.Itoa(mods.Val)
	case 447:
		mods.Str = "\"Elegy\"" + temp + strconv.Itoa(mods.Val)
	case 448:
		mods.Str = "\"Prelude\"" + temp + strconv.Itoa(mods.Val)
	case 449:
		mods.Str = "\"Hymnus\"" + temp + strconv.Itoa(mods.Val)
	case 450:
		mods.Str = "\"Virelai\"" + temp + strconv.Itoa(mods.Val)
	case 451:
		mods.Str = "\"Scherzo\"" + temp + strconv.Itoa(mods.Val)
	case 452:
		mods.Str = "All songs " + temp + strconv.Itoa(mods.Val)
	case 453:
		mods.Str = "Grants two additional song effects"
	case 454:
		mods.Str = "Increases song effect duration " + temp + strconv.Itoa(mods.Val)
	case 455:
		mods.Str = "Song spellcasting time -" + strconv.Itoa(Abs(mods.Val)) + "%"
	case 833:
		mods.Str = "Song recast delay -" + strconv.Itoa(Abs(mods.Val/1000))
	case 98:
		mods.Str = "Enhances \"Camouflage\" effect " + temp + strconv.Itoa(mods.Val) + "%"
	case 305:
		mods.Str = "\"Recycle\"" + temp + strconv.Itoa(mods.Val) + "%"
	case 365:
		mods.Str = "Enhances \"Snapshot\" Effect " + strconv.Itoa(mods.Val) + "%"
	case 359:
		mods.Str = "Increases \"Rapid Shot\" activation rate " + temp + strconv.Itoa(mods.Val) + "%"
	case 340:
		mods.Str = "\"Widescan\""
	case 420:
		mods.Str = "Increases \"Barrage\" accuracy " + temp + strconv.Itoa(mods.Val)
	case 422:
		mods.Str = "Enhances \"Double Shot\" effect " + temp + strconv.Itoa(mods.Val) + "%"
	case 423:
		mods.Str = "Enhances \"Snapshot\" effect of \"Velocity Shot\"" + temp + strconv.Itoa(mods.Val) + "%"
	case 424:
		mods.Str = "Enhances \"Velocity Shot\" ranged attack" + temp + strconv.Itoa(mods.Val) + "%"
	case 425:
		mods.Str = "Increases \"Shadowbind\" duration " + temp + strconv.Itoa(mods.Val)
	case 312:
		mods.Str = "Enhances \"Scavange\" effect " + temp + strconv.Itoa(mods.Val) + "%"
	case 314:
		mods.Str = "Enhances \"Snapshot\" effect " + temp + strconv.Itoa(mods.Val) + "%"

	case 94:
		mods.Str = "Increases \"Meditate\" duration " + temp + strconv.Itoa(mods.Val)
	case 95:
		mods.Str = "Increases \"Warding Circle\" duration " + temp + strconv.Itoa(mods.Val)
	case 306:
		mods.Str = "Enhances \"Zanshin\" effect " + temp + strconv.Itoa(mods.Val) + "%"
	case 508:
		mods.Str = "\"Third Eye\": \"Counter\" rate " + temp + strconv.Itoa(mods.Val)
	case 839:
		mods.Str = "Enhances \"Third Eye\" Effect " + strconv.Itoa(mods.Val) + "%"
	case 307:
		mods.Str = "Tracks \"Utsusemi\" Shadows"
	case 900:
		mods.Str = "\"Utsusemi\" Bonus " + temp + strconv.Itoa(mods.Val)
	case 308:
		mods.Str = "\"Ninja Tool Expertise\" "
	case 522:
		mods.Str = "Enhances ninjutsu damage " + temp + strconv.Itoa(mods.Val) + "%"
	case 911:
		mods.Str = "Increases \"Daken\" Chance" + temp + strconv.Itoa(mods.Val)
	case 859:
		mods.Str = "\"Ancient Circle\" duration" + temp + strconv.Itoa(mods.Val)
	case 361:
		mods.Str = "\"Jump\" TP Bonus " + temp + strconv.Itoa(mods.Val)
	case 362:
		mods.Str = "\"Jump\" Attack Bonus " + temp + strconv.Itoa(mods.Val) + "%"
	case 363:
		mods.Str = "Enhances \"High Jump\" effect " + temp + strconv.Itoa(mods.Val) + "%"
	case 828:
		mods.Str = "Augments jump attacks"
	case 829:
		mods.Str = "Wyvern uses breaths more effectively"
	case 974:
		mods.Str = "Adds subjob traits to wyvern"
	case 371:
		mods.Str = "Perpetuation cost -" + strconv.Itoa(Abs(mods.Val))
	case 372:
		mods.Str = "Weather: Avatar perpetuation cost -" + strconv.Itoa(Abs(mods.Val))
	case 373:
		mods.Str = "Depending on day or weather: Halves avatar perpetuation cost"
	case 346:
		mods.Str = "Avatar perpetuation cost -" + strconv.Itoa(Abs(mods.Val))
	case 357:
		mods.Str = "\"Blood Pact\" ability delay -" + strconv.Itoa(Abs(mods.Val))
	case 540:
		mods.Str = "Enhances \"Elemental Siphon\" effect " + temp + strconv.Itoa(mods.Val)
	case 541:
		mods.Str = "\"Blood Pact\" recast time II -" + strconv.Itoa(Abs(mods.Val))
	case 126:
		mods.Str = "\"Blood Pact\" Rage damage " + temp + strconv.Itoa(mods.Val) + "%"
	case 913:
		mods.Str = "Occasionally reduces \"Blood Pact\" MP Cost"
	case 960:
		mods.Str = "Reduces \"spirits\" cooldown time"

		//STOPPED
	case 528:
		mods.Str = "Increases \"Phantom Roll\" area of effect " + temp + strconv.Itoa(mods.Val)
	case 411:
		mods.Str = "Enhances \"Quick Draw\" effect " + temp + strconv.Itoa(mods.Val)
	case 504:
		mods.Str = "Enhances \"Maneuver\" effects " + temp + strconv.Itoa(mods.Val)
	case 505:
		mods.Str = "Reduces \"Overload\" rate " + temp + strconv.Itoa(mods.Val) + "%"
	case 490:
		mods.Str = "Increases \"Samba\" duration " + temp + strconv.Itoa(mods.Val)
	case 491:
		mods.Str = "\"Waltz\" potency " + temp + strconv.Itoa(mods.Val) + "%"
	case 492:
		mods.Str = "Increases \"Jig\" duration " + temp + strconv.Itoa(mods.Val)
	case 493:
		mods.Str = "\"Violent Flourish\" accuracy " + temp + strconv.Itoa(mods.Val)
	case 494:
		if mods.Val == 20 {
			mods.Str = "Enhances \"Violent Flourish\" effects " + temp + strconv.Itoa(mods.Val)
		} else {
			mods.Str = "Increases \"Steps\" accuracy " + temp + strconv.Itoa(mods.Val)
			mods.Str = "Augments \"Steps\""
		}
	case 403:
		mods.Str = "Increases \"Steps\" accuracy " + temp + strconv.Itoa(mods.Val)
	case 498:
		mods.Str = "Increases \"Samba\" duration " + temp + strconv.Itoa(mods.Val) + "%"
	case 836:
		mods.Str = "\"Reverse Flourish\"" + temp + strconv.Itoa(mods.Val)
	case 395:
		mods.Str = "Black magic casting time -" + strconv.Itoa(Abs(mods.Val)) + "%"
	case 399:
		mods.Str = "Enhances \"Celerity\" and \"Alacrity\" effect " + temp + strconv.Itoa(mods.Val)
	case 334:
		mods.Str = "Enhances \"Light Arts\" effect " + temp + strconv.Itoa(mods.Val)
	case 335:
		mods.Str = "Enhances \"Dark Arts\" effect " + temp + strconv.Itoa(mods.Val)
	case 336:
		mods.Str = "Enhances \"Addendum: White\" effect " + temp + strconv.Itoa(mods.Val)
	case 337:
		mods.Str = "Enhances \"Addendum: Black\" effect " + temp + strconv.Itoa(mods.Val)
	case 339:
		mods.Str = "Increases \"Regen\" duration " + temp + strconv.Itoa(mods.Val)
	case 401:
		mods.Str = "Enhances \"Sublimation\" effect " + temp + strconv.Itoa(mods.Val)
	case 489:
		mods.Str = "Grimoire: Reduces spellcasting time -" + strconv.Itoa(Abs(mods.Val)) + "%"
	case 432:
		mods.Str = "Sword enhancement spell damage " + temp + strconv.Itoa(mods.Val)
	case 344:
		mods.Str = "\"Spikes\" spell damage " + temp + strconv.Itoa(mods.Val)
	case 345:
		mods.Str = "\"Conserve TP\"" + temp + strconv.Itoa(mods.Val)
	case 355:
		mods.Weaponskill = GetWSName(mods.Val)
	case 356:
		mods.Str = "In Dynamis: \"" + GetWSName(mods.Val) + "\""
	case 366:
		mods.Str = "Main hand: DMG:" + strconv.Itoa(itemdmg+mods.Val) + ""
	case 368:
		mods.Str = "\"Regain\"" + temp + strconv.Itoa(mods.Val)
	case 369:
		mods.Str = "Adds \"Refresh\" effect " + temp + strconv.Itoa(mods.Val)
	case 370:
		mods.Str = "Adds \"Regen\" effect " + temp + strconv.Itoa(mods.Val)
	case 374:
		mods.Str = "\"Cure\" potency " + temp + strconv.Itoa(mods.Val) + "%"
	case 375:
		mods.Str = "Potency of \"Cure\" effect received " + temp + strconv.Itoa(mods.Val) + "%"
	case 377:
		mods.Str = "Adds weapon rank to main weapon"
	case 378:
		mods.Str = "Adds weapon rank to sub weapon"
	case 379:
		mods.Str = "Adds weapon rank to ranged weapon"
	case 385:
		if mods.Val > 200 {
			mods.Str = "\"Shield Bash\"" + romanic(mods.Val)
		} else {
			mods.Str = "\"Shield Bash\"" + temp + strconv.Itoa(mods.Val)
		}
	case 386:
		mods.Str = "\"Kick Attacks\" damage" + temp + strconv.Itoa(mods.Val)
	case 482:
		mods.Str = "\"Kick Attacks\"" + temp + strconv.Itoa(mods.Val)
	case 392:
		mods.Str = "\"Weapon Bash\"" + temp + strconv.Itoa(mods.Val)
	case 402:
		mods.Str = "Enhances effect of wyvern's breath " + temp + strconv.Itoa(mods.Val)
	case 408:
		mods.Str = "Increases \"Double Attack\" damage " + temp + strconv.Itoa(mods.Val)
	case 409:
		mods.Str = "\"Triple Attack\" damage " + temp + strconv.Itoa(mods.Val)
	case 480:
		mods.Str = "Absorbs magic damage taken " + temp + strconv.Itoa(mods.Val) + "%"
	case 416:
		mods.Str = "Occasionally annuls damage from physical attacks " + temp + strconv.Itoa(mods.Val) + "%"
	case 430:
		mods.Str = "\"Quadruple Attack\"" + temp + strconv.Itoa(mods.Val) + "%"
	case 459:
		mods.Str = "Occasionally absorbs fire damage " + temp + strconv.Itoa(mods.Val) + "%"
	case 460:
		mods.Str = "Occasionally absorbs earth damage " + temp + strconv.Itoa(mods.Val) + "%"
	case 461:
		mods.Str = "Occasionally absorbs water damage " + temp + strconv.Itoa(mods.Val) + "%"
	case 462:
		mods.Str = "Occasionally absorbs wind damage " + temp + strconv.Itoa(mods.Val) + "%"
	case 463:
		mods.Str = "Occasionally absorbs ice damage " + temp + strconv.Itoa(mods.Val) + "%"
	case 464:
		mods.Str = "Occasionally absorbs lightning damage " + temp + strconv.Itoa(mods.Val) + "%"
	case 465:
		mods.Str = "Occasionally absorbs light damage " + temp + strconv.Itoa(mods.Val) + "%"
	case 466:
		mods.Str = "Occasionally absorbs dark damage " + temp + strconv.Itoa(mods.Val) + "%"
	case 475:
		mods.Str = "Occasionally absorbs magic damage taken " + temp + strconv.Itoa(mods.Val) + "%"
	case 476:
		mods.Str = "Occasionally annuls magic damage taken " + temp + strconv.Itoa(mods.Val) + "%"
	case 512:
		mods.Str = "Occasionally absorbs physical damage taken " + temp + strconv.Itoa(mods.Val)
	case 431:
		// mods.Str = "Additional Effect" + temp + strconv.Itoa(mods.Val)
	case 499:
		mods.Str = "Physical damage: \"" + GetSpikes(mods.Val) + " Spikes\" effect"
	case 310:
		mods.Str = "Enhances \"Cursna\" effect " + temp + strconv.Itoa(mods.Val) + "%"
	case 414:
		mods.Str = "\"Retaliation\"" + temp + strconv.Itoa(mods.Val)
	case 507:
		mods.Str = fmt.Sprintf("Occasionally does %.1fx damage", float64(mods.Val)/100)
	case 510:
		mods.Str = "Reduces clamming \"incidents\""
	case 511:
		mods.Str = "Extends chocobo riding time " + temp + strconv.Itoa(mods.Val)
	case 513:
		mods.Str = "Improves mining and harvesting results " + temp + strconv.Itoa(mods.Val)
		if mods.Val == 1 {
			mods.Str += "%"
		}
	case 514:
		mods.Str = "Improves logging and harvesting results " + temp + strconv.Itoa(mods.Val)
		if mods.Val == 1 {
			mods.Str += "%"
		}
	case 515:
		mods.Str = "Improves mining, logging and harvesting results " + temp + strconv.Itoa(mods.Val)
		if mods.Val == 1 {
			mods.Str += "%"
		}
	case 517:
		mods.Str = "Chocobo Digging Bonus"
	case 313:
		mods.Str = "Enhances \"Dia\" effect " + temp + strconv.Itoa(mods.Val)
	case 315:
		mods.Str = "Enhances the effect of \"Dia\" and \"Aspir\" " + temp + strconv.Itoa(mods.Val)
	case 521:
		mods.Str = "Augments \"Absorb\" spells " + romanic(mods.Val)
	case 826:
		mods.Str = "Virtue stone equipped: Occasionally attacks twice"
	case 525:
		mods.Str = "Augments \"Convert\" x" + strconv.Itoa(mods.Val)
	case 526:
		mods.Str = "Enhances \"Sneak Attack\" effect " + romanic(mods.Val)
	case 527:
		mods.Str = "Enhances \"Trick Attack\" effect " + romanic(mods.Val)
	case 529:
		mods.Str = "Enhances \"Refresh\" potency " + romanic(mods.Val)
	case 530:
		mods.Str = "MP not depleted when magic used  " + temp + strconv.Itoa(mods.Val)
	case 531:
		mods.Str = "Gain full benefit of Firesday/fire weather bonuses"
	case 532:
		mods.Str = "Gain full benefit of Earthsday/earth weather bonuses"
	case 533:
		mods.Str = "Gain full benefit of Watersday/water weather bonuses"
	case 534:
		mods.Str = "Gain full benefit of Windsday/wind weather bonuses"
	case 535:
		mods.Str = "Gain full benefit of Iceday/ice weather bonuses"
	case 536:
		mods.Str = "Gain full benefit of Lightningday/lightning weather bonuses"
	case 537:
		mods.Str = "Gain full benefit of Lightsday/light weather bonuses"
	case 538:
		mods.Str = "Gain full benefit of Darksday/dark weather bonuses"
	case 539:
		mods.Str = "Enhances \"Stoneskin\" effect " + temp + strconv.Itoa(mods.Val)
	case 565:
		mods.Str = "Elemental magic affected by day " + temp + strconv.Itoa(mods.Val)
	case 566:
		mods.Str = "\"Iridescence\""
	case 567:
		mods.Str = "Elemental resistance spells " + temp + strconv.Itoa(mods.Val)
	case 827:
		mods.Str = "Enhances elemental resistance spells " + temp + strconv.Itoa(mods.Val)
	case 568:
		mods.Str = "\"Rapture\"" + temp + strconv.Itoa(mods.Val)
	case 569:
		mods.Str = "\"Ebullience\"" + temp + strconv.Itoa(mods.Val)
	case 972:
		mods.Str = "Mount Movement Speed" + temp + strconv.Itoa(mods.Val) + "%"
	//generated added
	case 969:
		mods.Str = "Max Tp" + temp + strconv.Itoa(mods.Val) // Modifies a battle entity's maximum tp
	// err: strconv.Atoi: parsing "128/256": invalid syntaxerr: strconv.Atoi: parsing "100%": invalid syntaxerr: strconv.Atoi: parsing "100%": invalid syntax
	case 309:
		mods.Str = "Blue Points" + temp + strconv.Itoa(mods.Val) // Tracks extra blue points
	case 945:
		mods.Str = "Blue Learn Chance" + temp + strconv.Itoa(mods.Val) // Additional chance to learn blue magic
	case 936:
		mods.Str = "Monster Correlation Bonus" + temp + strconv.Itoa(mods.Val) // AF head
	case 382:
		mods.Str = "Exp Bonus" + temp + strconv.Itoa(mods.Val) //
	case 542:
		mods.Str = "Job Bonus Chance" + temp + strconv.Itoa(mods.Val) // Chance to apply job bonus to COR roll without having the job in the party.
	case 316:
		mods.Str = "Dmg Reflect" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 317:
		mods.Str = "Roll Rogues" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 318:
		mods.Str = "Roll Gallants" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 319:
		mods.Str = "Roll Chaos" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 320:
		mods.Str = "Roll Beast" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 321:
		mods.Str = "Roll Choral" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 322:
		mods.Str = "Roll Hunters" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 323:
		mods.Str = "Roll Samurai" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 324:
		mods.Str = "Roll Ninja" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 325:
		mods.Str = "Roll Drachen" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 326:
		mods.Str = "Roll Evokers" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 327:
		mods.Str = "Roll Magus" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 328:
		mods.Str = "Roll Corsairs" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 329:
		mods.Str = "Roll Puppet" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 330:
		mods.Str = "Roll Dancers" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 331:
		mods.Str = "Roll Scholars" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 869:
		mods.Str = "Roll Bolters" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 870:
		mods.Str = "Roll Casters" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 871:
		mods.Str = "Roll Coursers" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 872:
		mods.Str = "Roll Blitzers" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 873:
		mods.Str = "Roll Tacticians" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 874:
		mods.Str = "Roll Allies" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 875:
		mods.Str = "Roll Misers" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 876:
		mods.Str = "Roll Companions" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 877:
		mods.Str = "Roll Avengers" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 878:
		mods.Str = "Roll Naturalists" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 879:
		mods.Str = "Roll Runeists" + temp + strconv.Itoa(mods.Val) // Tracks totals
	case 332:
		mods.Str = "Bust" + temp + strconv.Itoa(mods.Val) // # of busts
	case 881:
		mods.Str = "Phantom Roll" + temp + strconv.Itoa(mods.Val) // Phantom Roll+ Effect from SOA Rings.
	case 882:
		mods.Str = "Phantom Duration" + temp + strconv.Itoa(mods.Val) // Phantom Roll Duration +.
	case 842:
		mods.Str = "Auto Decision Delay" + temp + strconv.Itoa(mods.Val) // Reduces the Automaton's global decision delay
	case 843:
		mods.Str = "Auto Shield Bash Delay" + temp + strconv.Itoa(mods.Val) // Reduces the Automaton's global shield bash delay
	case 844:
		mods.Str = "Auto Magic Delay" + temp + strconv.Itoa(mods.Val) // Reduces the Automaton's global magic delay
	case 845:
		mods.Str = "Auto Healing Delay" + temp + strconv.Itoa(mods.Val) // Reduces the Automaton's global healing delay
	case 846:
		mods.Str = "Auto Healing Threshold" + temp + strconv.Itoa(mods.Val) // Increases the healing trigger threshold
	case 847:
		mods.Str = "Burden Decay" + temp + strconv.Itoa(mods.Val) // Increases amount of burden removed per tick
	case 848:
		mods.Str = "Auto Shield Bash Slow" + temp + strconv.Itoa(mods.Val) // Adds a slow effect to Shield Bash
	case 849:
		mods.Str = "Auto Tp Efficiency" + temp + strconv.Itoa(mods.Val) + "%" // Causes the Automaton to wait to form a skillchain when its master is > 90% TP
	case 850:
		mods.Str = "Auto Scan Resists" + temp + strconv.Itoa(mods.Val) // Causes the Automaton to scan a target's resistances
	case 853:
		mods.Str = "Repair Effect" + temp + strconv.Itoa(mods.Val) // Removes # of status effects from the Automaton
	case 854:
		mods.Str = "Repair Potency" + temp + strconv.Itoa(mods.Val) + "%" // Note: Only affects amount regenerated by a %
	case 855:
		mods.Str = "Prevent Overload" + temp + strconv.Itoa(mods.Val) // Overloading erases a water maneuver (except on water overloads) instead
	case 938:
		mods.Str = "Auto Steam Jacket" + temp + strconv.Itoa(mods.Val) // Causes the Automaton to mitigate damage from successive attacks of the same type
	case 939:
		mods.Str = "Auto Steam Jacked Reduction" + temp + strconv.Itoa(mods.Val) // Amount of damage reduced with Steam Jacket
	case 940:
		mods.Str = "Auto Schurzen" + temp + strconv.Itoa(mods.Val) // Prevents fatal damage leaving the automaton at 1HP and consumes an Earth manuever
	case 941:
		mods.Str = "Auto Equalizer" + temp + strconv.Itoa(mods.Val) // Reduces damage received according to damage taken
	case 942:
		mods.Str = "Auto Performance Boost" + temp + strconv.Itoa(mods.Val) + "%" // Increases the performance of other attachments by a percentage
	case 943:
		mods.Str = "Auto Analyzer" + temp + strconv.Itoa(mods.Val) // Causes the Automaton to mitigate damage from a special attack a number of times
	case 333:
		mods.Str = "Finishing Moves" + temp + strconv.Itoa(mods.Val) // Tracks # of finishing moves
	case 497:
		mods.Str = "Waltz Delay" + temp + strconv.Itoa(mods.Val) // Waltz Ability Delay modifier (-1 mod is -1 second)
	case 502:
		mods.Str = "Spectral Jig Duration" + temp + strconv.Itoa(mods.Val) + "%" // Spectral Jig duration bonus in percents
	case 393:
		mods.Str = "Black Magic Cost" + temp + strconv.Itoa(mods.Val) // MP cost for black magic (light/dark arts)
	case 396:
		mods.Str = "White Magic Cast" + temp + strconv.Itoa(mods.Val) // Cast time for black magic (light/dark arts)
	case 397:
		mods.Str = "Black Magic Recast" + temp + strconv.Itoa(mods.Val) // Recast time for black magic (light/dark arts)
	case 398:
		mods.Str = "White Magic Recast" + temp + strconv.Itoa(mods.Val) // Recast time for white magic (light/dark arts)
	case 338:
		mods.Str = "Light Arts Regen" + temp + strconv.Itoa(mods.Val) // Regen bonus flat HP amount from Light Arts and Tabula Rasa
	case 478:
		mods.Str = "Helix Effect" + temp + strconv.Itoa(mods.Val) //
	case 477:
		mods.Str = "Helix Duration" + temp + strconv.Itoa(mods.Val) //
	case 400:
		mods.Str = "Stormsurge Effect" + temp + strconv.Itoa(mods.Val) //
	case 12100:
		mods.Str = "Cardinal Chant" + temp + strconv.Itoa(mods.Val)
	case 12101:
		mods.Str = "Indi Duration" + temp + strconv.Itoa(mods.Val)
	case 12102:
		mods.Str = "Geomancy" + temp + strconv.Itoa(mods.Val)
	case 12103:
		mods.Str = "Widened Compass" + temp + strconv.Itoa(mods.Val)
	case 12104:
		mods.Str = "Mending Halation" + temp + strconv.Itoa(mods.Val)
	case 12105:
		mods.Str = "Radial Arcana" + temp + strconv.Itoa(mods.Val)
	case 12106:
		mods.Str = "Curative Recantation" + temp + strconv.Itoa(mods.Val)
	case 12107:
		mods.Str = "Primeval Zeal" + temp + strconv.Itoa(mods.Val)
	case 341:
		mods.Str = "Enspell" + temp + strconv.Itoa(mods.Val) // stores the type of enspell active (0 if nothing)
	case 343:
		mods.Str = "Enspell Dmg" + temp + strconv.Itoa(mods.Val) // stores the base damage of the enspell before reductions
	case 856:
		mods.Str = "Enspell Chance" + temp + strconv.Itoa(mods.Val) // Chance of enspell activating (0
	case 342:
		mods.Str = "Spikes" + temp + strconv.Itoa(mods.Val) // store the type of spike spell active (0 if nothing)
	case 880:
		mods.Str = "Savetp" + temp + strconv.Itoa(mods.Val) // SAVETP Effect for Miser's Roll / ATMA / Hagakure.
	case 944:
		mods.Str = "Conserve Tp" + temp + strconv.Itoa(mods.Val) // Conserve TP trait
	case 963:
		mods.Str = "Inquartata" + temp + strconv.Itoa(mods.Val) + "%" // increases parry rate by a flat %.
	case 347:
		mods.Str = "Fire Affinity Dmg" + temp + strconv.Itoa(mods.Val) // They're stored separately due to Magian stuff - they can grant different levels of
	case 348:
		mods.Str = "Ice Affinity Dmg" + temp + strconv.Itoa(mods.Val) // the damage/acc/perp affinity on the same weapon
	case 349:
		mods.Str = "Wind Affinity Dmg" + temp + strconv.Itoa(mods.Val) + "%" // Each level of damage affinity is +/-5% damage
	case 350:
		mods.Str = "Earth Affinity Dmg" + temp + strconv.Itoa(mods.Val) // +/-1 mp/tic. This means that anyone adding these modifiers will have to add
	case 351:
		mods.Str = "Thunder Affinity Dmg" + temp + strconv.Itoa(mods.Val) // 1 to the wiki amount. For example
	case 352:
		mods.Str = "Water Affinity Dmg" + temp + strconv.Itoa(mods.Val) // DMG
	case 353:
		mods.Str = "Light Affinity Dmg" + temp + strconv.Itoa(mods.Val)
	case 354:
		mods.Str = "Dark Affinity Dmg" + temp + strconv.Itoa(mods.Val)
	case 544:
		mods.Str = "Fire Affinity Acc" + temp + strconv.Itoa(mods.Val)
	case 545:
		mods.Str = "Ice Affinity Acc" + temp + strconv.Itoa(mods.Val)
	case 546:
		mods.Str = "Wind Affinity Acc" + temp + strconv.Itoa(mods.Val)
	case 547:
		mods.Str = "Earth Affinity Acc" + temp + strconv.Itoa(mods.Val)
	case 548:
		mods.Str = "Thunder Affinity Acc" + temp + strconv.Itoa(mods.Val)
	case 549:
		mods.Str = "Water Affinity Acc" + temp + strconv.Itoa(mods.Val)
	case 550:
		mods.Str = "Light Affinity Acc" + temp + strconv.Itoa(mods.Val)
	case 551:
		mods.Str = "Dark Affinity Acc" + temp + strconv.Itoa(mods.Val)
	case 553:
		mods.Str = "Fire Affinity Perp" + temp + strconv.Itoa(mods.Val)
	case 554:
		mods.Str = "Ice Affinity Perp" + temp + strconv.Itoa(mods.Val)
	case 555:
		mods.Str = "Wind Affinity Perp" + temp + strconv.Itoa(mods.Val)
	case 556:
		mods.Str = "Earth Affinity Perp" + temp + strconv.Itoa(mods.Val)
	case 557:
		mods.Str = "Thunder Affinity Perp" + temp + strconv.Itoa(mods.Val)
	case 558:
		mods.Str = "Water Affinity Perp" + temp + strconv.Itoa(mods.Val)
	case 559:
		mods.Str = "Light Affinity Perp" + temp + strconv.Itoa(mods.Val)
	case 560:
		mods.Str = "Dark Affinity Perp" + temp + strconv.Itoa(mods.Val)
	case 358:
		mods.Str = "Stealth" + temp + strconv.Itoa(mods.Val) //
	case 946:
		mods.Str = "Sneak Duration" + temp + strconv.Itoa(mods.Val) // Additional duration in seconds
	case 947:
		mods.Str = "Invisible Duration" + temp + strconv.Itoa(mods.Val) // Additional duration in seconds
	case 367:
		mods.Str = "Sub Dmg Rating" + temp + strconv.Itoa(mods.Val) // adds damage rating to off hand weapon
	case 406:
		mods.Str = "Regain Down" + temp + strconv.Itoa(mods.Val) // plague
	case 405:
		mods.Str = "Refresh Down" + temp + strconv.Itoa(mods.Val) // plague
	case 404:
		mods.Str = "Regen Down" + temp + strconv.Itoa(mods.Val) // poison
	case 260:
		mods.Str = "Cure Potency Ii" + temp + strconv.Itoa(mods.Val) + "%" // % cure potency II | bonus from gear is capped at 30
	case 376:
		mods.Str = "Ranged Dmg Rating" + temp + strconv.Itoa(mods.Val) // adds damage rating to ranged weapon
	case 380:
		mods.Str = "Delayp" + temp + strconv.Itoa(mods.Val) + "%" // delay addition percent (does not affect tp gain)
	case 381:
		mods.Str = "Ranged Delayp" + temp + strconv.Itoa(mods.Val) + "%" // ranged delay addition percent (does not affect tp gain)
	case 410:
		mods.Str = "Zanshin Double Damage" + temp + strconv.Itoa(mods.Val) + "%" // Zanshin's double damage chance %.
	case 479:
		mods.Str = "Rapid Shot Double Damage" + temp + strconv.Itoa(mods.Val) + "%" // Rapid shot's double damage chance %.
	case 481:
		mods.Str = "Extra Dual Wield Attack" + temp + strconv.Itoa(mods.Val) // Chance to land an extra attack when dual wielding
	case 415:
		mods.Str = "Samba Double Damage" + temp + strconv.Itoa(mods.Val) // Double damage chance when samba is up.
	case 417:
		mods.Str = "Quick Draw Triple Damage" + temp + strconv.Itoa(mods.Val) // Chance to do triple damage with quick draw.
	case 418:
		mods.Str = "Bar Element Null Chance" + temp + strconv.Itoa(mods.Val) // Bar Elemental spells will occasionally NULLify damage of the same element.
	case 419:
		mods.Str = "Grimoire Instant Cast" + temp + strconv.Itoa(mods.Val) // Spells that match your current Arts will occasionally cast instantly
	case 456:
		mods.Str = "Reraise I" + temp + strconv.Itoa(mods.Val) // Reraise.
	case 457:
		mods.Str = "Reraise II" + temp + strconv.Itoa(mods.Val) // Reraise II.
	case 458:
		mods.Str = "Reraise III" + temp + strconv.Itoa(mods.Val) // Reraise III.
	case 467:
		mods.Str = "Fire Null" + temp + strconv.Itoa(mods.Val) //
	case 468:
		mods.Str = "Ice Null" + temp + strconv.Itoa(mods.Val) //
	case 469:
		mods.Str = "Wind Null" + temp + strconv.Itoa(mods.Val) //
	case 470:
		mods.Str = "Earth Null" + temp + strconv.Itoa(mods.Val) //
	case 471:
		mods.Str = "Ltng Null" + temp + strconv.Itoa(mods.Val) //
	case 472:
		mods.Str = "Water Null" + temp + strconv.Itoa(mods.Val) //
	case 473:
		mods.Str = "Light Null" + temp + strconv.Itoa(mods.Val) //
	case 474:
		mods.Str = "Dark Null" + temp + strconv.Itoa(mods.Val) //
	case 500:
		mods.Str = "Item Spikes Dmg" + temp + strconv.Itoa(mods.Val) // Damage of an items spikes
	case 501:
		mods.Str = "Item Spikes Chance" + temp + strconv.Itoa(mods.Val) // Chance of an items spike proc
	case 950:
		mods.Str = "//Item Addeffect Element" + temp + strconv.Itoa(mods.Val) // Element of the Additional Effect or Spikes
	case 951:
		mods.Str = "//Item Addeffect Status" + temp + strconv.Itoa(mods.Val) // Status Effect ID to try to apply via Additional Effect or Spikes
	case 952:
		mods.Str = "//Item Addeffect Power" + temp + strconv.Itoa(mods.Val) // Base Power for effect in MOD_ITEM_ADDEFFECT_STATUS
	case 953:
		mods.Str = "//Item Addeffect Duration" + temp + strconv.Itoa(mods.Val) // Base Duration for effect in MOD_ITEM_ADDEFFECT_STATUS
	case 496:
		mods.Str = "Gov Clears" + temp + strconv.Itoa(mods.Val) + "%" // 4% bonus per Grounds of Valor Page clear
	case 506:
		mods.Str = "Extra Dmg Chance" + temp + strconv.Itoa(mods.Val/10) + "%" // Proc rate of OCC_DO_EXTRA_DMG. 111 would be 11.1%
	case 863:
		mods.Str = "Rem Occ Do Double Dmg" + temp + strconv.Itoa(mods.Val) // Proc rate for REM Aftermaths that apply "Occasionally do double damage"
	case 864:
		mods.Str = "Rem Occ Do Triple Dmg" + temp + strconv.Itoa(mods.Val) // Proc rate for REM Aftermaths that apply "Occasionally do triple damage"
	case 867:
		mods.Str = "Rem Occ Do Double Dmg Ranged" + temp + strconv.Itoa(mods.Val) // Ranged attack specific
	case 868:
		mods.Str = "Rem Occ Do Triple Dmg Ranged" + temp + strconv.Itoa(mods.Val) // Ranged attack specific
	case 865:
		mods.Str = "Mythic Occ Att Twice" + temp + strconv.Itoa(mods.Val) // Proc rate for "Occasionally attacks twice"
	case 866:
		mods.Str = "Mythic Occ Att Thrice" + temp + strconv.Itoa(mods.Val) // Proc rate for "Occasionally attacks thrice"
	case 412:
		mods.Str = "Eat Raw Fish" + temp + strconv.Itoa(mods.Val) //
	case 413:
		mods.Str = "Eat Raw Meat" + temp + strconv.Itoa(mods.Val) //
	case 67:
		mods.Str = "Enhances Cursna Rcvd" + temp + strconv.Itoa(mods.Val) // Potency of "Cursna" effects received
	case 495:
		mods.Str = "Enhances Holywater" + temp + strconv.Itoa(mods.Val) // Used by gear with the "Enhances Holy Water" or "Holy Water+" attribute
	case 509:
		mods.Str = "Clamming Improved Results" + temp + strconv.Itoa(mods.Val) //
	case 518:
		mods.Str = "Shieldblockrate" + temp + strconv.Itoa(mods.Val) // Affects shield block rate
	case 523:
		mods.Str = "Ammo Swing" + temp + strconv.Itoa(mods.Val) // Extra swing rate w/ ammo (ie. Jailer weapons). Use gearsets
	case 886:
		mods.Str = "Augments Assassins Charge" + temp + strconv.Itoa(mods.Val) + "%" // Gives Assassin's Charge +1% Critical Hit Rate per merit level
	case 887:
		mods.Str = "Augments Ambush" + temp + strconv.Itoa(mods.Val) + "%" // Gives +1% Triple Attack per merit level when Ambush conditions are met
	case 888:
		mods.Str = "Augments Feint" + temp + strconv.Itoa(mods.Val) // Feint will give another -10 Evasion per merit level
	case 889:
		mods.Str = "Augments Aura Steal" + temp + strconv.Itoa(mods.Val) + "%" // 20% chance of 2 effects to be dispelled or stolen per merit level
	case 912:
		mods.Str = "Augments Conspirator" + temp + strconv.Itoa(mods.Val) // Applies Conspirator benefits to player at the top of the hate list
	case 832:
		mods.Str = "Aquaveil Count" + temp + strconv.Itoa(mods.Val) // Modifies the amount of hits that Aquaveil absorbs before being removed
	case 890:
		mods.Str = "Enh Magic Duration" + temp + strconv.Itoa(mods.Val) + "%" // Enhancing Magic Duration increase %
	case 891:
		mods.Str = "Enhances Coursers Roll" + temp + strconv.Itoa(mods.Val) + "%" // Courser's Roll Bonus % chance
	case 892:
		mods.Str = "Enhances Casters Roll" + temp + strconv.Itoa(mods.Val) + "%" // Caster's Roll Bonus % chance
	case 893:
		mods.Str = "Enhances Blitzers Roll" + temp + strconv.Itoa(mods.Val) + "%" // Blitzer's Roll Bonus % chance
	case 894:
		mods.Str = "Enhances Allies Roll" + temp + strconv.Itoa(mods.Val) + "%" // Allies' Roll Bonus % chance
	case 895:
		mods.Str = "Enhances Tacticians Roll" + temp + strconv.Itoa(mods.Val) + "%" // Tactician's Roll Bonus % chance
	case 902:
		mods.Str = "Occult Acumen" + temp + strconv.Itoa(mods.Val) // Grants bonus TP when dealing damage with elemental or dark magic
	case 909:
		mods.Str = "Quick Magic" + temp + strconv.Itoa(mods.Val) // Percent chance spells cast instantly (also reduces recast to 0
	case 851:
		mods.Str = "Synth Success" + temp + strconv.Itoa(mods.Val) // Rate of synthesis success
	case 852:
		mods.Str = "Synth Skill Gain" + temp + strconv.Itoa(mods.Val) // Synthesis skill gain rate
	case 861:
		mods.Str = "Synth Fail Rate" + temp + strconv.Itoa(mods.Val) + "%" // Synthesis failure rate (percent)
	case 862:
		mods.Str = "Synth Hq Rate" + temp + strconv.Itoa(mods.Val) + "%" // High-quality success rate (not a percent)
	case 916:
		mods.Str = "Desynth Success" + temp + strconv.Itoa(mods.Val) // Rate of desynthesis success
	case 917:
		mods.Str = "Synth Fail Rate Fire" + temp + strconv.Itoa(mods.Val) // Amount synthesis failure rate is reduced when using a fire crystal
	case 918:
		mods.Str = "Synth Fail Rate Ice" + temp + strconv.Itoa(mods.Val) // Amount synthesis failure rate is reduced when using a ice crystal
	case 919:
		mods.Str = "Synth Fail Rate Wind" + temp + strconv.Itoa(mods.Val) // Amount synthesis failure rate is reduced when using a wind crystal
	case 920:
		mods.Str = "Synth Fail Rate Earth" + temp + strconv.Itoa(mods.Val) // Amount synthesis failure rate is reduced when using a earth crystal
	case 921:
		mods.Str = "Synth Fail Rate Lightning" + temp + strconv.Itoa(mods.Val) // Amount synthesis failure rate is reduced when using a lightning crystal
	case 922:
		mods.Str = "Synth Fail Rate Water" + temp + strconv.Itoa(mods.Val) // Amount synthesis failure rate is reduced when using a water crystal
	case 923:
		mods.Str = "Synth Fail Rate Light" + temp + strconv.Itoa(mods.Val) // Amount synthesis failure rate is reduced when using a light crystal
	case 924:
		mods.Str = "Synth Fail Rate Dark" + temp + strconv.Itoa(mods.Val) // Amount synthesis failure rate is reduced when using a dark crystal
	case 925:
		mods.Str = "Synth Fail Rate Wood" + temp + strconv.Itoa(mods.Val) // Amount synthesis failure rate is reduced when doing woodworking
	case 926:
		mods.Str = "Synth Fail Rate Smith" + temp + strconv.Itoa(mods.Val) // Amount synthesis failure rate is reduced when doing smithing
	case 927:
		mods.Str = "Synth Fail Rate Goldsmith" + temp + strconv.Itoa(mods.Val) // Amount synthesis failure rate is reduced when doing goldsmithing
	case 928:
		mods.Str = "Synth Fail Rate Cloth" + temp + strconv.Itoa(mods.Val) // Amount synthesis failure rate is reduced when doing clothcraft
	case 929:
		mods.Str = "Synth Fail Rate Leather" + temp + strconv.Itoa(mods.Val) // Amount synthesis failure rate is reduced when doing leathercraft
	case 930:
		mods.Str = "Synth Fail Rate Bone" + temp + strconv.Itoa(mods.Val) // Amount synthesis failure rate is reduced when doing bonecraft
	case 931:
		mods.Str = "Synth Fail Rate Alchemy" + temp + strconv.Itoa(mods.Val) // Amount synthesis failure rate is reduced when doing alchemy
	case 932:
		mods.Str = "Synth Fail Rate Cook" + temp + strconv.Itoa(mods.Val) // Amount synthesis failure rate is reduced when doing cooking
	case 570:
		mods.Str = "Weaponskill Damage Base" + temp + strconv.Itoa(mods.Val)
	case 840:
		mods.Str = "All Wsdmg All Hits" + temp + strconv.Itoa(mods.Val) // Generic (all Weaponskills) damage
	case 841:
		mods.Str = "All Wsdmg First Hit" + temp + strconv.Itoa(mods.Val) // Generic (all Weaponskills) damage
	case 949:
		mods.Str = "Ws No Deplete" + temp + strconv.Itoa(mods.Val) + "%" // % chance a Weaponskill depletes no TP.
	case 980:
		mods.Str = "Ws Str Bonus" + temp + strconv.Itoa(mods.Val) + "%" // % bonus to str_wsc.
	case 957:
		mods.Str = "Ws Dex Bonus" + temp + strconv.Itoa(mods.Val) + "%" // % bonus to dex_wsc.
	case 981:
		mods.Str = "Ws Vit Bonus" + temp + strconv.Itoa(mods.Val) + "%" // % bonus to vit_wsc.
	case 982:
		mods.Str = "Ws Agi Bonus" + temp + strconv.Itoa(mods.Val) + "%" // % bonus to agi_wsc.
	case 983:
		mods.Str = "Ws Int Bonus" + temp + strconv.Itoa(mods.Val) + "%" // % bonus to int_wsc.
	case 984:
		mods.Str = "Ws Mnd Bonus" + temp + strconv.Itoa(mods.Val) + "%" // % bonus to mnd_wsc.
	case 985:
		mods.Str = "Ws Chr Bonus" + temp + strconv.Itoa(mods.Val) + "%" // % bonus to chr_wsc.
	case 914:
		mods.Str = "Experience Retained" + temp + strconv.Itoa(mods.Val) + "%" // Experience points retained upon death (this is a percentage)
	case 915:
		mods.Str = "Capacity Bonus" + temp + strconv.Itoa(mods.Val) // Capacity point bonus granted
	case 933:
		mods.Str = "Conquest Bonus" + temp + strconv.Itoa(mods.Val) + "%" // Conquest points bonus granted (percentage)
	case 934:
		mods.Str = "Conquest Region Bonus" + temp + strconv.Itoa(mods.Val) // Increases the influence points awarded to the player's nation when receiving conquest points
	case 935:
		mods.Str = "Campaign Bonus" + temp + strconv.Itoa(mods.Val) + "%" // Increases the evaluation for allied forces by percentage
	case 993:
		mods.Str = "Subtle Blow Ii" + temp + strconv.Itoa(mods.Val) + "%" // Subtle Blow II Effect (Cap 50%) Total Effect (SB + SB_II cap 75%)
	case 995:
		mods.Str = "Gardening Wilt Bonus" + temp + strconv.Itoa(mods.Val) // Increases the number of Vanadays a plant can survive before it wilts
	case 988:
		mods.Str = "Super Jump" + temp + strconv.Itoa(mods.Val)
	case 958:
		mods.Str = "Spdef Down" + temp + strconv.Itoa(mods.Val)
	case 1176:
		mods.Str = "Susc To Ws Stun" + temp + strconv.Itoa(mods.Val)
	case 1178:
		mods.Str = "Enhances Cover" + temp + strconv.Itoa(mods.Val)
	case 1179:
		mods.Str = "Augments Cover" + temp + strconv.Itoa(mods.Val)
	case 1180:
		mods.Str = "Covered Mp Flag" + temp + strconv.Itoa(mods.Val)
	case 1181:
		mods.Str = "Rampart Stoneskin" + temp + strconv.Itoa(mods.Val)
	case 1182:
		mods.Str = "Tame Success Rate" + temp + strconv.Itoa(mods.Val)
	case 1183:
		mods.Str = "Magic Stacking Mdt" + temp + strconv.Itoa(mods.Val)
	case 1184:
		mods.Str = "Fire Burden Decay" + temp + strconv.Itoa(mods.Val)
	case 1185:
		mods.Str = "Burden Decay Ignore Chance" + temp + strconv.Itoa(mods.Val)
	case 1186:
		mods.Str = "Fire Burden Perc Extra" + temp + strconv.Itoa(mods.Val) + "%"
	case 1187:
		mods.Str = "Super Intimidation" + temp + strconv.Itoa(mods.Val)
	case 1000:
		mods.Str = "Penguin Ring Effect" + temp + strconv.Itoa(mods.Val) // +2 on fishing arrow delay / fish movement for mini - game
	case 1001:
		mods.Str = "Albatross Ring Effect" + temp + strconv.Itoa(mods.Val) // adds 30 seconds to mini - game time
	case 1002:
		mods.Str = "Pelican Ring Effect" + temp + strconv.Itoa(mods.Val) // adds extra skillup roll for fishing
	case 1224:
		mods.Str = "Vermin Circle" + temp + strconv.Itoa(mods.Val)
	case 1225:
		mods.Str = "Bird Circle" + temp + strconv.Itoa(mods.Val)
	case 1226:
		mods.Str = "Amorph Circle" + temp + strconv.Itoa(mods.Val)
	case 1227:
		mods.Str = "Lizard Circle" + temp + strconv.Itoa(mods.Val)
	case 1228:
		mods.Str = "Aquan Circle" + temp + strconv.Itoa(mods.Val)
	case 1229:
		mods.Str = "Plantoid Circle" + temp + strconv.Itoa(mods.Val)
	case 1230:
		mods.Str = "Beast Circle" + temp + strconv.Itoa(mods.Val)
	case 1231:
		mods.Str = "Undead Circle" + temp + strconv.Itoa(mods.Val)
	case 1232:
		mods.Str = "Arcana Circle" + temp + strconv.Itoa(mods.Val)
	case 1233:
		mods.Str = "Dragon Circle" + temp + strconv.Itoa(mods.Val)
	case 1234:
		mods.Str = "Demon Circle" + temp + strconv.Itoa(mods.Val)
	case 1235:
		mods.Str = "Empty Circle" + temp + strconv.Itoa(mods.Val)
	case 1236:
		mods.Str = "Humanoid Circle" + temp + strconv.Itoa(mods.Val)
	case 1237:
		mods.Str = "Lumorian Circle" + temp + strconv.Itoa(mods.Val)
	case 1238:
		mods.Str = "Luminion Circle" + temp + strconv.Itoa(mods.Val)
	case 970:
		mods.Str = "Pet Att Latent" + temp + strconv.Itoa(mods.Val) // Pet Attack bonus used for latents
	case 971:
		mods.Str = "Pet Acc Latent" + temp + strconv.Itoa(mods.Val) // Pet Accuracy bonus used for latents
	default:
	}
	return mods
}

func GetSpikes(value int) string {
	switch value {
	case 1:
		return "Blaze"
	case 2:
		return "Ice"
	case 4:
		return "Curse"
	case 5:
		return "Shock"
	case 10:
		return "Death"
	default:
		return ""
	}
}
