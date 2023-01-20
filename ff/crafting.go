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

type Recipe struct {
	ID          int
	Desynth     byte
	KeyItem     string
	Wood        byte
	Smith       byte
	Gold        byte
	Cloth       byte
	Leather     byte
	Bone        byte
	Alchemy     byte
	Cook        byte
	Crystal     int16
	HQCrys      int16
	Ing1        int
	Ing2        int
	Ing3        int
	Ing4        int
	Ing5        int
	Ing6        int
	Ing7        int
	Ing8        int
	Result      int
	ResHQ1      int
	ResHQ2      int
	ResHQ3      int
	ResQty      byte
	ResH1Qty    byte
	ResH2Qty    byte
	ResH3Qty    byte
	CrystalName *string
	HQCrysName  *string
	Ing1Name    *string
	Ing2Name    *string
	Ing3Name    *string
	Ing4Name    *string
	Ing5Name    *string
	Ing6Name    *string
	Ing7Name    *string
	Ing8Name    *string
	ResultName  *string
	ResHQ1Name  *string
	ResHQ2Name  *string
	ResHQ3Name  *string
}

type RecipeShort struct {
	Desynth    byte
	KeyItem    string
	Skill      byte
	Crystal    int16
	ResultName *string
	Result     int
	ResQty     byte
}

type RecipeIngSearch struct {
	ResultName *string
	Result     int
}

type FishShort struct {
	FishId    int
	Name      *string
	MinSkill  byte
	MaxSkill  byte
	Size      byte
	Legendary byte
	Item      byte
}

type FishGroup struct {
	GroupId int
	FishId  int
	Catch   []*FishCatch
}

type FishCatch struct {
	ZoneId   int16
	AreaId   byte
	AreaName *string
	ZoneName *string
	ZoneDiff byte
}

type FishLure struct {
	LureId  int16   // Lure ID
	Name    *string // Lure Name
	Power   int16
	Type    byte  // Type of lure (stackable bait 0/lure 1/ special one time use 2)
	Maxhook byte  // Maximum number of fish lure can hook (sabiki rig can hook up to 3 of certain fish)
	Flags   int32 // Lure Flags (sinking, item bonus)
	Losable byte  // Can the lure be lost?
	IsMMM   byte  // Is Moblin Maze Monger bait? (probably not special, haven't tested)
}

type FishDets struct {
	Fish  Fish
	Lures []*FishLure
	Rods  []*Rod
}

type Rod struct {
	RodId        int32   // Rod ID
	Name         *string // Rod Name
	Material     byte    // Rod Material (wood 0/synthetic 1/legendary 2)
	SizeType     byte    // small/large
	Flags        int32   // Rod Flags (large bonus/small penalty)
	Durability   byte    // Rod Durability on scale of 10-100
	FishAttack   byte    // Fish Attack Multiplier
	LgdBonusAtk  byte    // Legendary Fish Bonus Attack (added to fishAttack on legendary fish)
	MissRegen    byte    // Miss Regen multiplier | formula:(floor((missRegen/20) * fishSize) * 10)
	LgdMissRegen byte    // Miss Regen against legendary fish
	FishTime     byte    // Rod base catch time limit
	LgdBonusTime byte    // Legendary fish bonus time.
	SmDelayBonus byte    // Small fish arrow delay bonus
	SmMoveBonus  byte    // Small fish movement bonus
	LgDelayBonus byte    // Large fish arrow delay bonus
	LgMoveBonus  byte    // Large fish movement bonus
	Multiplier   byte    // muliplier used in time formulas, possibly other things
	Breakable    byte    // Is the rod breakable?
	BrokenRodId  int32   // Replacement broken rod ID
	IsMMM        byte    // Is Moblin Maze Monger rod? (does crazy stat mods)
}

type Fish struct {
	FishId    int32   // Fish ID
	FishName  *string // Fish Name
	MinSkill  byte    // Minimum hook skill level
	MaxSkill  byte    // Maximum hook skill level
	Size      byte    // 'Size' of fish, used for most hook/rod calculations
	BaseDelay byte    // base hook arrow delay
	BaseMove  byte    // base hook movement
	MinLength int16   // minimum fish length (in lms)
	MaxLength int16   // maximum fish length (in lms)
	SizeType  byte    // small/large
	WaterType byte    // all 0, fresh 1 / salt 2
	Log       byte    // quest/mission log
	Quest     int16   // quest/mission id
	Flags     int32   // fish flags (half size, tropical, bottom dweller)
	Legendary byte    // is this a legendary fish? (affects certain rod calcs)
	LFlags    int32   // legendary flags (half fish time)
	Item      byte    // item/fish
	Maxhook   byte    // maximum that can be hooked (with sabiki rig)
	Rarity    int16   // [0-1000] : 0 = not rare, 1 = rarest, 1000 = most common
	KeyItem   string  // required key item
}

type GuildStoreItem struct {
	GuildId  int32
	MinPrice int
	MaxPrice int
	MaxQty   int16
	DailyInc int16
	InitQty  int16
	Npc      []NPCGuild
}

type NPCGuild struct {
	Area  string
	Name  string
	Guild string
	POS   string
}

func GetFishAreasByID(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", corStr)

	fID := -1
	var err error
	if val, ok := pathParams["fID"]; ok {
		fID, err = strconv.Atoi(val)
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

	rows, err := db.Query("SELECT fishid, groupid FROM fishing_group WHERE fishid = ?", fID)
	if err != nil {
		fmt.Println("error selecting fish by level")
		panic(err.Error())
	}
	defer rows.Close()

	var rps []*FishGroup
	for rows.Next() {
		row := new(FishGroup)
		if err := rows.Scan(&row.FishId, &row.GroupId); err != nil {
			fmt.Printf("fish by grp error: %v", err)
		}

		rows2, err := db.Query("SELECT q.zoneid, q.areaid, a.name, b.name, b.difficulty FROM fishing_catch as q JOIN fishing_area as a ON q.zoneid = a.zoneid && q.areaid = a.areaid JOIN fishing_zone as b ON q.zoneid = b.zoneid WHERE q.groupid = ?", row.GroupId)
		if err != nil {
			fmt.Println("error selecting fish by area")
			panic(err.Error())
		}
		defer rows2.Close()

		for rows2.Next() {
			row2 := new(FishCatch)
			if err := rows2.Scan(&row2.ZoneId, &row2.AreaId, &row2.AreaName, &row2.ZoneName, &row2.ZoneDiff); err != nil {
				fmt.Printf("fish area by grp error: %v", err)
			}
			row.Catch = append(row.Catch, row2)
		}
		rps = append(rps, row)
	}
	jsonData, _ := json.Marshal(&rps)
	w.Write(jsonData)
}

func GetFishDetsByID(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", corStr)

	fID := -1
	var err error
	if val, ok := pathParams["fID"]; ok {
		fID, err = strconv.Atoi(val)
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

	frows, err := db.Query("SELECT q.fishid, q.name, q.min_skill_level, q.skill_level, q.size, q.base_delay, q.base_move, q.min_length, q.max_length, q.size_type, q.water_type, q.log, q.quest, q.flags, q.legendary, q.legendary_flags, q.item, q.max_hook, q.rarity, q.required_keyitem FROM fishing_fish as q WHERE q.fishid = ?", fID)
	if err != nil {
		fmt.Println("error selecting fish by fishid")
		panic(err.Error())
	}
	defer frows.Close()

	var fish FishDets
	for frows.Next() {
		var key int
		row1 := new(Fish)
		if err := frows.Scan(&row1.FishId, &row1.FishName, &row1.MinSkill, &row1.MaxSkill, &row1.Size, &row1.BaseDelay, &row1.BaseMove, &row1.MinLength, &row1.MaxLength, &row1.SizeType, &row1.WaterType, &row1.Log, &row1.Quest, &row1.Flags, &row1.Legendary, &row1.LFlags, &row1.Item, &row1.Maxhook, &row1.Rarity, &key); err != nil {
			fmt.Printf("fish by id error: %v", err)
		}
		if key > 0 {
			row1.KeyItem = KeyItems[strconv.Itoa(key)].Name
		}
		fish.Fish = *row1
	}
	if fish.Fish.FishName == nil {
		w.Write(nil)
		return
	}

	rows, err := db.Query("SELECT q.lureid, q.power, a.name, a.luretype, a.maxhook, a.losable, a.flags, a.mmm FROM fishing_lure_affinity as q JOIN fishing_lure as a ON q.lureid = a.lureid WHERE q.fishid = ? ORDER BY q.power DESC LIMIT 3", fID)
	if err != nil {
		fmt.Println("error selecting fish lure by fishid")
		panic(err.Error())
	}
	defer rows.Close()

	var rps []*FishLure
	for rows.Next() {
		row := new(FishLure)
		if err := rows.Scan(&row.LureId, &row.Power, &row.Name, &row.Type, &row.Maxhook, &row.Losable, &row.Flags, &row.IsMMM); err != nil {
			fmt.Printf("fish lure by id error: %v", err)
		}
		rps = append(rps, row)
	}
	fish.Lures = rps

	rows3, err := db.Query("SELECT q.rodid, q.name, q.material, q.size_type, q.flags, q.durability, q.fish_attack, q.miss_regen, q.fish_time, q.sm_delay_bonus, q.sm_move_bonus, q.lg_delay_bonus, q.lg_move_bonus, q.multiplier, q.breakable FROM fishing_rod as q WHERE q.size_type = ? or q.rodid = ? or q.rodid = ? ORDER BY q.durability DESC LIMIT 5", fish.Fish.SizeType, 17011, 17386)
	if err != nil {
		fmt.Println("error selecting fish rod by fishid")
		panic(err.Error())
	}
	defer rows3.Close()

	var rods []*Rod
	for rows3.Next() {
		row3 := new(Rod)
		if err := rows3.Scan(&row3.RodId, &row3.Name, &row3.Material, &row3.SizeType, &row3.Flags, &row3.Durability, &row3.FishAttack, &row3.MissRegen, &row3.FishTime, &row3.SmDelayBonus, &row3.SmMoveBonus, &row3.LgDelayBonus, &row3.LgMoveBonus, &row3.Multiplier, &row3.Breakable); err != nil {
			fmt.Printf("fish rod by id error: %v", err)
		}
		rods = append(rods, row3)
	}
	fish.Rods = rods

	jsonData, _ := json.Marshal(fish)
	w.Write(jsonData)
}

// grab all the group ids (list) for a fish from fishing_group. Then grab zoneid and area id (list) from fishing catch.
//then grab all items that match zoneid and areaid from catch from fishing_area.

// for bait, grab the lureid and power > 0 from lure_affinity using fishid then grab lure info from fishing_lure

/*
enum LUREFLAG : uint32
{
    LUREFLAG_NORMAL = 0x00,
    LUREFLAG_SINKING = 0x01,
    LUREFLAG_ITEM_BONUS = 0x02,
    LUREFLAG_ITEM_MEGA_BONUS = 0x04
};

enum FISHINGLEGENDARY : uint32
{
    FISHINGLEGENDARY_NORMAL = 0x00,
    FISHINGLEGENDARY_HALFTIME = 0x01,  // cuts base rod fishing time in half
    FISHINGLEGENDARY_NORODTIMEBONUS = 0x02,  // do not add the normal legendary rod bonus
    FISHINGLEGENDARY_ADDTIMEBONUS = 0x04   // add bonus fishing time based on rod multiplier
};
*/

func GetFishByLvl(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", corStr)
	sID := pathParams["sID"]
	incID := pathParams["incID"]
	l1ID := -1
	var err error
	if val, ok := pathParams["l1ID"]; ok {
		l1ID, err = strconv.Atoi(val)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "need a number"}`))
			return
		}
	}

	l2ID := -1
	if val, ok := pathParams["l2ID"]; ok {
		l2ID, err = strconv.Atoi(val)
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

	var rows *sql.Rows
	if incID == "false" {
		rows, err = db.Query(fmt.Sprintf("SELECT fishid, name, min_skill_level, skill_level, size, legendary, item FROM fishing_fish WHERE %s >= %d && %s <= %d && item = 0 ORDER BY %s", sID, l1ID, sID, l2ID, sID))
		if err != nil {
			fmt.Println("error selecting fish by level")
			panic(err.Error())
		}
	} else {
		rows, err = db.Query(fmt.Sprintf("SELECT fishid, name, min_skill_level, skill_level, size, legendary, item FROM fishing_fish WHERE %s >= %d && %s <= %d ORDER BY %s", sID, l1ID, sID, l2ID, sID))
		if err != nil {
			fmt.Println("error selecting fish by level")
			panic(err.Error())
		}
	}
	defer rows.Close()

	var rps []*FishShort
	for rows.Next() {
		row := new(FishShort)
		if err := rows.Scan(&row.FishId, &row.Name, &row.MinSkill, &row.MaxSkill, &row.Size, &row.Legendary, &row.Item); err != nil {
			fmt.Printf("fish by level error: %v", err)
		}
		rps = append(rps, row)
	}
	jsonData, _ := json.Marshal(&rps)
	w.Write(jsonData)
}

func GetRecipesByCraft(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", corStr)
	sID := pathParams["sID"]
	l1ID := -1
	var err error
	if val, ok := pathParams["l1ID"]; ok {
		l1ID, err = strconv.Atoi(val)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "need a number"}`))
			return
		}
	}

	l2ID := -1
	if val, ok := pathParams["l2ID"]; ok {
		l2ID, err = strconv.Atoi(val)
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

	rows, err := db.Query(fmt.Sprintf("SELECT Desynth, KeyItem, %s, Crystal, Result, ResultName, ResultQty FROM synth_recipes WHERE %s >= %d && %s <= %d ORDER BY %s", sID, sID, l1ID, sID, l2ID, sID))
	if err != nil {
		fmt.Println("error selecting recipes by craft")
		panic(err.Error())
	}
	defer rows.Close()

	var rps []*RecipeShort
	for rows.Next() {
		var key int
		row := new(RecipeShort)
		if err := rows.Scan(&row.Desynth, &key, &row.Skill, &row.Crystal, &row.Result, &row.ResultName, &row.ResQty); err != nil {
			fmt.Printf("recipes by craft error: %v", err)
		}
		if key > 0 {
			row.KeyItem = KeyItems[strconv.Itoa(key)].Name
		}
		rps = append(rps, row)
	}

	jsonData, _ := json.Marshal(&rps)
	w.Write(jsonData)
}

func GetRecipesUsingItem(w http.ResponseWriter, r *http.Request) {
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

	rows, err := db.Query(fmt.Sprintf("SELECT Result, ResultName FROM synth_recipes WHERE Ingredient1 = %d || Ingredient2 = %d || Ingredient3 = %d || Ingredient4 = %d || Ingredient5 = %d || Ingredient6 = %d || Ingredient7 = %d || Ingredient8 = %d", sID, sID, sID, sID, sID, sID, sID, sID))
	if err != nil {
		fmt.Println("error selecting recipes by ing item")
		panic(err.Error())
	}
	defer rows.Close()

	var rps []*RecipeIngSearch
	for rows.Next() {
		row := new(RecipeIngSearch)
		if err := rows.Scan(&row.Result, &row.ResultName); err != nil {
			fmt.Printf("recipes by ing item error: %v", err)
		}
		rps = append(rps, row)
	}

	jsonData, _ := json.Marshal(&rps)
	w.Write(jsonData)
}

func GetStoreItems(w http.ResponseWriter, r *http.Request) {
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

	rows, err := db.Query("SELECT q.guildid, q.min_price, q.max_price, q.max_quantity, q.daily_increase, q.initial_quantity FROM guild_shops as q WHERE q.itemid = ?", sID)
	if err != nil {
		fmt.Println("error selecting guild shop item")
		panic(err.Error())
	}
	defer rows.Close()

	var rps []*GuildStoreItem
	for rows.Next() {
		row := new(GuildStoreItem)
		if err := rows.Scan(&row.GuildId, &row.MinPrice, &row.MaxPrice, &row.MaxQty, &row.DailyInc, &row.InitQty); err != nil {
			fmt.Printf("guild shop error: %v", err)
		}
		row.Npc = GuildMap[row.GuildId]
		if row.Npc != nil { // some npc scripts may not exist
			rps = append(rps, row)
		}
	}

	jsonData, _ := json.Marshal(&rps)
	w.Write(jsonData)
}

var GuildMap = map[int32][]NPCGuild{
	514: {{
		Area:  "Windurst Woods",
		Name:  "Shih Tayuun",
		Guild: "Bonecrafting Guild Merchant",
		POS:   "-3.064 -6.25 -131.374 241",
	}},
	516: {{
		Area:  "Selbina",
		Name:  "Gibol",
		Guild: "Clothcrafting Guild Merchant",
		POS:   "13.591 -7.287 8.569 248",
	}, {
		Area:  "Selbina",
		Name:  "Tilala",
		Guild: "Clothcrafting Guild Merchant",
		POS:   "14.344 -7.912 10.276 248",
	}},
	517: {{
		Area:  "Port Windurst",
		Name:  "Babubu",
		Guild: "Fishing Guild Merchant",
		POS:   "-175.185 -3.324 70.445 240",
	}},
	519: {{
		Area:  "Bibiki Bay",
		Name:  "Mep Nhapopoluko",
		Guild: "Fishing Guild Merchant",
		POS:   "464.350 -6 752.731 4",
	}},
	520: {{
		Area:  "Ship bound for Selbina or Pirates",
		Name:  "Rajmonda",
		Guild: "Fishing Guild Merchant",
		POS:   "1.841 -2.101 -9.000 220",
	}},
	521: {{
		Area:  "Ship bound for Mhaura or Pirates",
		Name:  "Lokhong",
		Guild: "Fishing Guild Merchant",
		POS:   "1.841 -2.101 -9.000 221",
	}},
	522: {{
		Area:  "Open sea route to Al Zahbi",
		Name:  "Cehn Teyohngo",
		Guild: "Fishing Guild Merchant",
		POS:   "4.986 -2.101 -12.026 46",
	}},
	523: {{
		Area:  "Open sea route to Mhaura",
		Name:  "Pashi Maccaleh",
		Guild: "Fishing Guild Merchant",
		POS:   "4.986 -2.101 -12.026 47",
	}},
	524: {{
		Area:  "Silver Sea route to Nashmau",
		Name:  "Jidwahn",
		Guild: "Fishing Guild Merchant",
		POS:   "4.986 -2.101 -12.026 58",
	}},
	525: {{
		Area:  "Silver Sea route to Al Zahbi",
		Name:  "Yahliq",
		Guild: "Fishing Guild Merchant",
		POS:   "4.986 -2.101 -12.026 59",
	}},
	528: {{
		Area:  "Mhaura",
		Name:  "Celestina",
		Guild: "Goldsmith Guild Merchant",
		POS:   "-37.624 -16.050 75.681 249",
	}, {
		Area:  "Mhaura",
		Name:  "Yabby Tanmikey",
		Guild: "Goldsmith Guild Merchant",
		POS:   "-36.459 -16.000 76.840 249",
	}},
	529: {{
		Area:  "Southern San d'Oria",
		Name:  "Cletae",
		Guild: "Leathercraft Guild Merchant",
		POS:   "-189.142 -8.800 14.449 230",
	}, {
		Area:  "Southern San d'Oria",
		Name:  "Kueh Igunahmori",
		Guild: "Leathercraft Guild Merchant",
		POS:   "-194.791 -8.800 13.130 230",
	}},
	530: {{
		Area:  "Windurst Waters",
		Name:  "Kopopo",
		Guild: "Cooking Guild Merchant",
		POS:   "-103.935 -2.875 74.304 238",
	}, {
		Area:  "Windurst Waters",
		Name:  "Chomo Jinjahl",
		Guild: "Cooking Guild Merchant",
		POS:   "-105.094 -2.222 73.791 238",
	}},
	531: {{
		Area:  "Northern San d'Oria",
		Name:  "Doggomehr",
		Guild: "Blacksmith Guild Merchant",
		POS:   "-193.920 3.999 162.027 231",
	}, {
		Area:  "Northern San d'Oria",
		Name:  "Lucretia",
		Guild: "Blacksmith Guild Merchant",
		POS:   "-193.729 3.999 159.412 231",
	}},
	532: {{
		Area:  "Mhaura",
		Name:  "Kamilah",
		Guild: "Blacksmith Guild Merchant",
		POS:   "-64.302 -16.000 35.261 249",
	}, {
		Area:  "Mhaura",
		Name:  "Mololo",
		Guild: "Blacksmith Guild Merchant",
		POS:   "-64.278 -16.624 34.120 249",
	}},
	534: {{
		Area:  "Carpenters' Landing",
		Name:  "Beugungel",
		Guild: "Woodworking Guild Merchant",
		POS:   "-333.729, -5.512, 475.647 2",
	}},
	5132: {{
		Area:  "Northern San d'Oria",
		Name:  "Chaupire",
		Guild: "Woodworking Guild Merchant",
		POS:   "-174.476 3.999 281.854 231",
	}, {
		Area:  "Northern San d'Oria",
		Name:  "Cauzeriste",
		Guild: "Woodworking Guild Merchant",
		POS:   "-175.946 3.999 280.301 231",
	}},
	5152: {{
		Area:  "Windurst Woods",
		Name:  "Kuzah Hpirohpon",
		Guild: "Clothcrafting Guild Merchant",
		POS:   "-80.068 -3.25 -127.686 241",
	}, {
		Area:  "Windurst Woods",
		Name:  "Meriri",
		Guild: "Clothcrafting Guild Merchant",
		POS:   "-76.471 -3.55 -128.341 241",
	}},
	5182: {{
		Area:  "Selbina",
		Name:  "Graegham",
		Guild: "Fishing Guild Merchant",
		POS:   "-12.423 -7.287 8.665 248",
	}, {
		Area:  "Selbina",
		Name:  "Mendoline",
		Guild: "Fishing Guild Merchant",
		POS:   "-13.603 -7.287 10.916 248",
	}},
	5262: {{
		Area:  "Bastok Mines",
		Name:  "Maymunah",
		Guild: "Alchemy Guild Merchant",
		POS:   "108.738 5.017 -3.129 234",
	}, {
		Area:  "Bastok Mines",
		Name:  "Odoba",
		Guild: "Alchemy Guild Merchant",
		POS:   "108.473 5.017 1.089 234",
	}},
	5272: {{
		Area:  "Bastok Markets",
		Name:  "Teerth",
		Guild: "Goldsmithing Guild Merchant",
		POS:   "-205.190 -7.814 -56.507 235",
	}, {
		Area:  "Bastok Markets",
		Name:  "Visala",
		Guild: "Goldsmithing Guild Merchant",
		POS:   "-202.000 -7.814 -56.823 235",
	}},
	5332: {{
		Area:  "Metalworks",
		Name:  "Amulya",
		Guild: "Blacksmithing Guild Merchant",
		POS:   "-106.093 0.999 -24.564 237",
	}, {
		Area:  "Metalworks",
		Name:  "Vicious Eye",
		Guild: "Blacksmithing Guild Merchant",
		POS:   "-106.132 0.999 -28.757 237",
	}},
	60427: {{
		Area:  "Al Zahbi",
		Name:  "Ndego",
		Guild: "Smithing Guild Merchant",
		POS:   "-37.192 0.000 -33.949 48",
	}},
	60428: {{
		Area:  "Al Zahbi",
		Name:  "Dehbi Moshal",
		Guild: "Woodworking Guild Merchant",
		POS:   "-71.563 -5.999 -57.544 48",
	}},
	60429: {{
		Area:  "Al Zahbi",
		Name:  "Bornahn",
		Guild: "Goldsmithing Guild Merchant",
		POS:   "46.011 0.000 -42.713 48",
	}},
	60430: {{
		Area:  "Al Zahbi",
		Name:  "Taten-Bilten",
		Guild: "Clothcraft Guild Merchant",
		POS:   "71.598 -6.000 -56.930 48",
	}},
	60417: {{
		Area:  "Lower Jeuno",
		Name:  "Akamafula",
		Guild: "Tenshodo Merchant",
		POS:   "28.465 2.899 -46.699 245",
	}},
	60418: {{
		Area:  "Windurst Walls",
		Name:  "Scavnix",
		Guild: "Standard Merchant",
		POS:   "17.731 0.106 239.626 239",
	}, {
		Area:  "Port Bastok",
		Name:  "Blabbivix",
		Guild: "Standard Merchant",
		POS:   "-110.209 4.898 22.957 236",
	}, {
		Area:  "Northern San d'Oria",
		Name:  "Gaudylox",
		Guild: "Standard Merchant",
		POS:   "-147.593 11.999 222.550 231",
	}},
	60419: {{
		Area:  "Port Bastok",
		Name:  "Jabbar",
		Guild: "Tenshodo Merchant",
		POS:   "-99.718 -2.299 26.027 236",
	}},
	60420: {{
		Area:  "Port Bastok",
		Name:  "Silver Owl",
		Guild: "Tenshodo Merchant",
		POS:   "-99.155 4.649 23.292 236",
	}},
	60421: {{
		Area:  "Norg",
		Name:  "Achika",
		Guild: "Tenshodo Merchant",
		POS:   "1.300 0.000 19.259 252",
	}},
	60422: {{
		Area:  "Norg",
		Name:  "Chiyo",
		Guild: "Tenshodo Merchant",
		POS:   "5.801 0.020 -18.739 252",
	}},
	60423: {{
		Area:  "Norg",
		Name:  "Jirokichi",
		Guild: "Tenshodo Merchant",
		POS:   "-1.463 0.000 18.846 252",
	}},
	60424: {{
		Area:  "Norg",
		Name:  "Vuliaie",
		Guild: "Tenshodo Merchant",
		POS:   "-24.259 0.891 -19.556 252",
	}},
	60425: {{
		Area:  "Aht Urhgan Whitegate",
		Name:  "Gathweeda",
		Guild: "Alchemist Guild Merchant",
		POS:   "-81.322 -6.000 140.273 50",
	}, {
		Area:  "Aht Urhgan Whitegate",
		Name:  "Wahraga",
		Guild: "Alchemist Guild Merchant",
		POS:   "-76.836 -6.000 140.331 50",
	}},
	60426: {{
		Area:  "Aht Urhgan Whitegate",
		Name:  "Wahnid",
		Guild: "Fishing Guild Merchant",
		POS:   "-31.720 -6.000 -94.919 50",
	}},
	60431: {{
		Area:  "Nashmau",
		Name:  "Tsutsuroon",
		Guild: "Tenshodo Merchant",
		POS:   "-15.193 0.000 31.356 53",
	}},
}

/*
guild pattern update 310
[GUILD]pattern   7
Guild Test NPCs
        ["Abd-al-Raziq"] =      function (x)    TItemList = TI_Alchemy end,
        ["Peshi_Yohnts"] =      function (x)    TItemList = TI_Bonecraft end,
        ["Ponono"] =            function (x)    TItemList = TI_Clothcraft end,
        ["Piketo-Puketo"] =     function (x)    TItemList = TI_Cooking end,
        ["Thubu_Parohren"] =    function (x)    TItemList = TI_Fishing end,
        ["Reinberta"] =         function (x)    TItemList = TI_Goldsmithing end,
        ["Faulpie"] =           function (x)    TItemList = TI_Leathercraft end,
        ["Mevreauche"] =        function (x)    TItemList = TI_Smithing end,
        ["Ghemp"] =             function (x)    TItemList = TI_Smithing end,
        ["Cheupirudaux"] =      function (x)    TItemList = TI_Woodworking end,

pettype
pup 3  4 7   6
drg 2
smn  1


bst jugs between petid 20 and 68   in pet_list
enum MOBTYPE
{
    MOBTYPE_NORMAL      = 0x00,
    MOBTYPE_0X01        = 0x01, // available for use
    MOBTYPE_NOTORIOUS   = 0x02,
    MOBTYPE_FISHED      = 0x04,
    MOBTYPE_CALLED      = 0x08,
    MOBTYPE_BATTLEFIELD = 0x10,
    MOBTYPE_EVENT       = 0x20
};
enum DETECT : uint16
{
    DETECT_NONE        = 0x00,
    DETECT_SIGHT       = 0x01,
    DETECT_HEARING     = 0x02,
    DETECT_LOWHP       = 0x04,
    DETECT_NONE1       = 0x08,
    DETECT_NONE2       = 0x10,
    DETECT_MAGIC       = 0x20,
    DETECT_WEAPONSKILL = 0x40,
    DETECT_JOBABILITY  = 0x80,
    DETECT_SCENT       = 0x100
};

enum SPAWNTYPE
{
    SPAWNTYPE_NORMAL    = 0x00, // 00:00-24:00
    SPAWNTYPE_ATNIGHT   = 0x01, // 20:00-04:00
    SPAWNTYPE_ATEVENING = 0x02, // 18:00-06:00
    SPAWNTYPE_WEATHER   = 0x04,
    SPAWNTYPE_FOG       = 0x08, // 02:00-07:00
    SPAWNTYPE_MOONPHASE = 0x10,
    SPAWNTYPE_LOTTERY   = 0x20,
    SPAWNTYPE_WINDOWED  = 0x40,
    SPAWNTYPE_SCRIPTED  = 0x80, // scripted spawn
    SPAWNTYPE_PIXIE     = 0x100, // according to server amity
};

enum SPECIALFLAG
{
    SPECIALFLAG_NONE   = 0x0,
    SPECIALFLAG_HIDDEN = 0x1  // only use special when hidden
};

enum ROAMFLAG : uint16
{
    ROAMFLAG_NONE    = 0x00,
    ROAMFLAG_NONE0   = 0x01,  //
    ROAMFLAG_NONE1   = 0x02,  //
    ROAMFLAG_NONE2   = 0x04,  //
    ROAMFLAG_NONE3   = 0x08,  //
    ROAMFLAG_NONE4   = 0x10,  //
    ROAMFLAG_NONE5   = 0x20,  //
    ROAMFLAG_WORM    = 0x40,  // pop up and down when moving
    ROAMFLAG_AMBUSH  = 0x80,  // stays hidden until someone comes close (antlion)
    ROAMFLAG_EVENT   = 0x100, // calls lua method for roaming logic
    ROAMFLAG_IGNORE  = 0x200, // ignore all hate, except linking hate
    ROAMFLAG_STEALTH = 0x400  // stays name hidden and untargetable until someone comes close (chigoe)
};

enum BEHAVIOUR : uint16
{
    BEHAVIOUR_NONE         = 0x000,
    BEHAVIOUR_NO_DESPAWN   = 0x001, // mob does not despawn on death
    BEHAVIOUR_STANDBACK    = 0x002, // mob will standback forever
    BEHAVIOUR_RAISABLE     = 0x004, // mob can be raised via Raise spells
    BEHAVIOUR_NOHELP       = 0x008, // mob can not be targeted by helpful magic from players (cure, protect, etc)
    BEHAVIOUR_AGGRO_AMBUSH = 0x200, // mob aggroes by ambush
    BEHAVIOUR_NO_TURN      = 0x400  // mob does not turn to face target
};
tpz.itemType =
{
    BASIC       = 0x00,
    GENERAL     = 0x01,
    USABLE      = 0x02,
    PUPPET      = 0x04,
    ARMOR       = 0x08,
    WEAPON      = 0x10,
    CURRENCY    = 0x20,
    FURNISHING  = 0x40,
    LINKSHELL   = 0x80,
}

tpz.frames =
{
    HARLEQUIN  = 0x20,
    VALOREDGE  = 0x21,
    SHARPSHOT  = 0x22,
    STORMWAKER = 0x23,
}

enum ITEM_SUBTYPE
{
    ITEM_NORMAL             = 0x00,
    ITEM_LOCKED             = 0x01,
    ITEM_CHARGED            = 0x02,
    ITEM_AUGMENTED          = 0x04,
    ITEM_UNLOCKED           = 0xFE
};

enum ITEM_PUPPET_EQUIPSLOT
{
    ITEM_PUPPET_HEAD = 1,
    ITEM_PUPPET_FRAME = 2,
    ITEM_PUPPET_ATTACHMENT = 3
};

tpz.effectFlag =
{
    NONE            = 0x0000,
    DISPELABLE      = 0x0001,
    ERASABLE        = 0x0002,
    ATTACK          = 0x0004,
    EMPATHY         = 0x0008,
    DAMAGE          = 0x0010,
    DEATH           = 0x0020,
    MAGIC_BEGIN     = 0x0040,
    MAGIC_END       = 0x0080,
    ON_ZONE         = 0x0100,
    NO_LOSS_MESSAGE = 0x0200,
    INVISIBLE       = 0x0400,
    DETECTABLE      = 0x0800,
    NO_REST         = 0x1000,
    PREVENT_ACTION  = 0x2000,
    WALTZABLE       = 0x4000,
    FOOD            = 0x8000,
    SONG            = 0x10000,
    ROLL            = 0x20000,
    SYNTH_SUPPORT   = 0x40000,
    CONFRONTATION   = 0x80000,
    LOGOUT          = 0x100000,
    BLOODPACT       = 0x200000,
    ON_JOBCHANGE    = 0x400000,
    NO_CANCEL       = 0x800000,
    INFLUENCE       = 0x1000000,
    OFFLINE_TICK    = 0x2000000,
    AURA            = 0x4000000,
    ON_SYNC         = 0x8000000,
}
*/
