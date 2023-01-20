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

type MobGroup struct {
	GroupId     int32
	PoolId      int32
	ZoneId      int16
	MobId       int // This mob ID will be the ID for last spawn of this "mob" in this group
	Name        *string
	Respawn     int32
	SpawnType   byte
	DropId      int32
	HP          int32
	MP          int32
	MinLevel    byte
	MaxLevel    byte
	ZoneName    *string
	ZoneType    int16
	FamilyId    int16
	MJob        byte
	SJob        byte
	CmbSkill    byte
	CmbDelay    int16
	CmbDmgMult  int16
	Aggro       byte
	TrueSight   byte
	Links       byte
	MobType     int16
	Immunity    int32
	NamePrefix  byte
	Flag        int32
	EntityFlags int32
	SpellList   int16
	RoamFlag    int16
	SkillList   int16
	Family      *string
	Genre       *string
	SystemId    int16
	MobSize     byte
	Speed       byte
	FHP         int16
	FMP         int16
	STR         int16
	DEX         int16
	VIT         int16
	AGI         int16
	INT         int16
	MND         int16
	CHR         int16
	ATT         int16
	DEF         int16
	ACC         int16
	EVA         int16
	MEVA        int16
	Slash       float32
	Pierce      float32
	H2H         float32
	Impact      float32
	Fire        float32
	Ice         float32
	Wind        float32
	Earth       float32
	Lightning   float32
	Water       float32
	Light       float32
	Dark        float32
	Element     float32
	Detects     int32
	Charmable   byte
	Spells      []*Spell
	Skills      []*AttSkill
	Drops       []*Drop
	Modifs      []*Mods
}

type Drop struct {
	DropType  byte
	GroupId   byte
	GroupRate int16
	ItemId    int16
	ItemRate  int16
	ItemName  *string
}

type AttSkill struct {
	SkillId      int16
	Name         *string
	AOE          byte
	Distance     float32
	PrepTime     int16
	ValidTargets int16
	Flag         byte
	Param        int16
	Knockback    byte
	PrimSC       byte
	SecSC        byte
	TertSC       byte
}

type ListIDs struct {
	SkillId int32
}

type SpellListIDs struct {
	SpellId  int32
	MinLevel int16
	MaxLevel int16
}

type MobShort struct {
	GroupId  int32
	PoolId   int32
	Name     *string
	ZoneName *string
	ZoneId   int16
	MaxLevel byte
	Fish     byte
}

type FishMob struct {
	Log        byte    // Log ID
	Quest      byte    // Quest ID
	NM         byte    // Notorious Monster, no need to set for quest monsters
	NmFlags    int32   // Notorious Monster flags
	AreaId     byte    // Can this mob only be fished up from a certain area? i.e. PLD NM
	AreaName   *string // only populated if areaid is nonzero
	Rarity     int16   // [0-1000] : 0 = not rare, 1 = rarest, 1000 = most common
	MinRespawn int16   // minimum amount of time before mob can be hooked again
	MaxRespawn int16   // maximum amount of time before mob can be hooked again
	Level      byte    // level of monster (seem to be intervals of 10)
	Difficulty byte    // mob difficulty
	BaseDelay  byte    // base hook arrow delay
	BaseMove   byte    // base hook movement
	ReqBaitId  int16   // required bait
	AltBaitId  int16   // alternative required bait
	KeyItem    int16   // required key item
	Ranking    byte
	QuestOnly  byte // only fishable during quest
	Disabled   byte
}

func GetMobShort(w http.ResponseWriter, r *http.Request) {
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
	rows, err := db.Query("SELECT q.groupid, q.poolid, q.zoneid, q.name, q.maxLevel, a.name FROM mob_groups as q JOIN zone_settings as a ON q.zoneid = a.zoneid WHERE q.name LIKE ? ORDER BY q.name LIMIT 20", "%"+sID+"%")
	if err != nil {
		fmt.Println("error selecting mob group")
		panic(err.Error())
	}
	defer rows.Close()

	var items []*MobShort
	for rows.Next() {
		row := new(MobShort)
		if err := rows.Scan(&row.GroupId, &row.PoolId, &row.ZoneId, &row.Name, &row.MaxLevel, &row.ZoneName); err != nil {
			fmt.Printf("mob short error: %v", err)
		}

		frows, err := db.Query("SELECT q.level FROM fishing_mob as q WHERE q.name LIKE ? && q.zoneid = ?", "%"+sID+"%", row.ZoneId)
		if err != nil {
			fmt.Println("error selecting fishmob by fishid")
			panic(err.Error())
		}
		defer frows.Close()

		var fish FishMob
		for frows.Next() {
			row1 := new(FishMob)
			if err := frows.Scan(&row1.Level); err != nil {
				fmt.Printf("fish mob by id error: %v", err)
			}
			fish = *row1
		}
		if fish.Level > 0 { // No fishmob found
			row.Fish = 1
		}
		items = append(items, row)
	}

	jsonData, _ := json.Marshal(&items)
	w.Write(jsonData)
}

func GetMobGroupByID(w http.ResponseWriter, r *http.Request) {
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

	zID := -1
	if val, ok := pathParams["zID"]; ok {
		zID, err = strconv.Atoi(val)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "need a number"}`))
			return
		}
	}
	/*
		pID := -1
		if val, ok := pathParams["pID"]; ok {
			pID, err = strconv.Atoi(val)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message": "need a number"}`))
				return
			}
		} */
	items := GetMobDetails(sID, zID)
	jsonData, _ := json.Marshal(&items)
	w.Write(jsonData)
}

func GetMobDetails(sID int, zID int) []*MobGroup {
	var items []*MobGroup
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		fmt.Println("error validating sql.Open arguments")
		panic(err.Error())
	}
	defer db.Close()

	rows, err := db.Query("SELECT q.groupid, q.poolid, q.zoneid, q.name, q.respawntime, q.spawntype, q.dropid, q.HP, q.MP, q.minLevel, q.maxLevel, a.name, a.zonetype, b.familyid, b.mJob, b.sJob, b.cmbSkill, b.cmbDelay, b.cmbDmgMult, b.aggro, b.true_detection, b.links, b.mobType, b.immunity, b.name_prefix, b.flag, b.entityFlags, b.spellList, b.roamflag, b.skill_list_id, c.family, c.systemid, c.system, c.mobsize, c.speed, c.HP, c.MP, c.STR, c.DEX, c.VIT, c.AGI, c.INT, c.MND, c.CHR, c.ATT, c.DEF, c.ACC, c.EVA, c.Slash, c.Pierce, c.H2H, c.Impact, c.Fire, c.Ice, c.Wind, c.Earth, c.Lightning, c.Water, c.Light, c.Dark, c.Element, c.detects, c.charmable FROM mob_groups as q JOIN zone_settings as a ON q.zoneid = a.zoneid JOIN mob_pools as b ON q.poolid = b.poolid JOIN mob_family_system as c ON b.familyid = c.familyid WHERE q.groupid = ? AND q.zoneid = ?", sID, zID)
	if err != nil {
		fmt.Println("error selecting mob group")
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		row := new(MobGroup)
		if err := rows.Scan(&row.GroupId, &row.PoolId, &row.ZoneId, &row.Name, &row.Respawn, &row.SpawnType, &row.DropId, &row.HP, &row.MP, &row.MinLevel, &row.MaxLevel, &row.ZoneName, &row.ZoneType, &row.FamilyId, &row.MJob, &row.SJob, &row.CmbSkill, &row.CmbDelay, &row.CmbDmgMult, &row.Aggro, &row.TrueSight, &row.Links, &row.MobType, &row.Immunity, &row.NamePrefix, &row.Flag, &row.EntityFlags, &row.SpellList, &row.RoamFlag, &row.SkillList, &row.Family, &row.SystemId, &row.Genre, &row.MobSize, &row.Speed, &row.FHP, &row.FMP, &row.STR, &row.DEX, &row.VIT, &row.AGI, &row.INT, &row.MND, &row.CHR, &row.ATT, &row.DEF, &row.ACC, &row.EVA, &row.Slash, &row.Pierce, &row.H2H, &row.Impact, &row.Fire, &row.Ice, &row.Wind, &row.Earth, &row.Lightning, &row.Water, &row.Light, &row.Dark, &row.Element, &row.Detects, &row.Charmable); err != nil {
			fmt.Printf("mob by group error: %v", err)
		}

		//Get droplist
		hrows, err := db.Query("SELECT q.dropType, q.groupId, q.groupRate, q.itemId, q.itemRate, a.name FROM mob_droplist as q JOIN item_basic AS a ON q.itemId = a.itemid WHERE q.dropId = ?", row.DropId)
		if err != nil {
			fmt.Println("error selecting mob drop")
			panic(err.Error())
		}
		defer hrows.Close()
		for hrows.Next() {
			row7 := new(Drop)
			if err := hrows.Scan(&row7.DropType, &row7.GroupId, &row7.GroupRate, &row7.ItemId, &row7.ItemRate, &row7.ItemName); err != nil {
				fmt.Printf("mob drop error: %v", err)
			}
			row.Drops = append(row.Drops, row7)
		}

		// Start getting skills
		erows, err := db.Query("SELECT q.mob_skill_id FROM mob_skill_lists as q WHERE q.skill_list_id = ?", row.SkillList)
		if err != nil {
			fmt.Println("error selecting skill list")
			panic(err.Error())
		}
		defer erows.Close()

		for erows.Next() {
			row2 := new(ListIDs)
			if err := erows.Scan(&row2.SkillId); err != nil {
				fmt.Printf("skillid by error: %v", err)
			}

			srows, err := db.Query("SELECT q.mob_skill_id, q.mob_skill_name, q.mob_skill_aoe, q.mob_skill_distance, q.mob_prepare_time, q.mob_valid_targets, q.mob_skill_flag, q.mob_skill_param, q.knockback, q.primary_sc, q.secondary_sc, q.tertiary_sc FROM mob_skills as q WHERE q.mob_skill_id = ?", row2.SkillId)
			if err != nil {
				fmt.Println("error selecting mob skill")
				panic(err.Error())
			}
			defer srows.Close()
			for srows.Next() {
				row3 := new(AttSkill)
				if err := srows.Scan(&row3.SkillId, &row3.Name, &row3.AOE, &row3.Distance, &row3.PrepTime, &row3.ValidTargets, &row3.Flag, &row3.Param, &row3.Knockback, &row3.PrimSC, &row3.SecSC, &row3.TertSC); err != nil {
					fmt.Printf("mob skill error: %v", err)
				}
				row.Skills = append(row.Skills, row3)
			}
		}
		//Done getting skills

		// Start getting spells
		arows, err := db.Query("SELECT q.spell_id, q.min_level, q.max_level FROM mob_spell_lists as q WHERE q.spell_list_id = ? AND q.min_level <= ? AND q.max_level >= ?", row.SpellList, row.MaxLevel, row.MinLevel)
		if err != nil {
			fmt.Println("error selecting spell list")
			panic(err.Error())
		}
		defer arows.Close()

		for arows.Next() {
			row2 := new(SpellListIDs)
			if err := arows.Scan(&row2.SpellId, &row2.MinLevel, &row2.MaxLevel); err != nil {
				fmt.Printf("spellid2 by error: %v", err)
			}
			srows, err := db.Query("SELECT q.spellid, q.name, q.aoe, q.Jobs, q.spell_range, q.castTime, q.recastTime, q.skill, q.base, q.multiplier, q.CE, q.VE, q.validTargets, q.content_tag FROM spell_list as q WHERE q.spellid = ?", row2.SpellId)
			if err != nil {
				fmt.Println("error selecting mob spell")
				panic(err.Error())
			}
			defer srows.Close()
			for srows.Next() {
				row3 := new(Spell)
				if err := srows.Scan(&row3.SpellId, &row3.Name, &row3.AOE, &row3.Jobs, &row3.Range, &row3.CastTime, &row3.RecastTime, &row3.Skill, &row3.Base, &row3.Multiplier, &row3.CE, &row3.VE, &row3.ValidTarget, &row3.ContentTag); err != nil {
					fmt.Printf("mob spell error: %v", err)
				}
				row.Spells = append(row.Spells, row3)
			}
		}
		//Done getting spells

		// Get Mods
		mrows, err := db.Query("SELECT q.modid, q.value, q.is_mob_mod FROM mob_pool_mods as q WHERE q.poolid = ?", row.PoolId)
		if err != nil {
			fmt.Println("error selecting mods for pool")
			panic(err.Error())
		}
		defer mrows.Close()
		var mds []*Mods
		for mrows.Next() {
			mrow := new(Mods)
			if err := mrows.Scan(&mrow.Id, &mrow.Val, &mrow.IsMobMod); err != nil {
				fmt.Printf("mob pool mod error: %v", err)
			}
			if mrow.IsMobMod != 0 {
				mrow = GetMobMods(mrow)
			} else {
				mrow = GetMods(mrow)
			}
			mds = append(mds, mrow)
		}

		prows, err := db.Query("SELECT q.modid, q.value, q.is_mob_mod FROM mob_family_mods as q WHERE q.familyid = ?", row.FamilyId)
		if err != nil {
			fmt.Println("error selecting mods for group")
			panic(err.Error())
		}
		defer prows.Close()
		for prows.Next() {
			prow := new(Mods)
			if err := prows.Scan(&prow.Id, &prow.Val, &prow.IsMobMod); err != nil {
				fmt.Printf("mob fam mod error: %v", err)
			}
			if prow.IsMobMod != 0 {
				prow = GetMobMods(prow)
			} else {
				prow = GetMods(prow)
			}
			mds = append(mds, prow)
		}
		qrows, err := db.Query("SELECT q.modid, q.value, q.is_mob_mod FROM mob_spawn_mods as q WHERE q.mobid = ?", row.MobId)
		if err != nil {
			fmt.Println("error selecting spawn mods")
			panic(err.Error())
		}
		defer qrows.Close()
		for qrows.Next() {
			qrow := new(Mods)
			if err := qrows.Scan(&qrow.Id, &qrow.Val, &qrow.IsMobMod); err != nil {
				fmt.Printf("mob spawn mod error: %v", err)
			}
			if qrow.IsMobMod != 0 {
				qrow = GetMobMods(qrow)
			} else {
				qrow = GetMods(qrow)
			}
			mds = append(mds, qrow)
		}
		row.Modifs = mds
		row = GetStats(*row)
		// end getting Mods
		items = append(items, row)
	}
	return items
}

func GetFishMobByID(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", corStr)

	fID := pathParams["fID"]

	zID := -1
	var err error
	if val, ok := pathParams["zID"]; ok {
		zID, err = strconv.Atoi(val)
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

	frows, err := db.Query("SELECT q.level, q.ranking, q.difficulty, q.base_delay, q.base_move, q.log, q.quest, q.nm, q.nm_flags, q.areaid, q.rarity, q.min_respawn, q.max_respawn, q.required_baitid, q.alternative_baitid, q.required_keyitem, q.quest_only, q.disabled FROM fishing_mob as q WHERE q.name LIKE ? && q.zoneid = ?", "%"+fID+"%", zID)
	if err != nil {
		fmt.Println("error selecting fish by fishid")
		panic(err.Error())
	}
	defer frows.Close()

	var fish FishMob
	for frows.Next() {
		row1 := new(FishMob)
		if err := frows.Scan(&row1.Level, &row1.Ranking, &row1.Difficulty, &row1.BaseDelay, &row1.BaseMove, &row1.Log, &row1.Quest, &row1.NM, &row1.NmFlags, &row1.AreaId, &row1.Rarity, &row1.MinRespawn, &row1.MaxRespawn, &row1.ReqBaitId, &row1.AltBaitId, &row1.KeyItem, &row1.QuestOnly, &row1.Disabled); err != nil {
			fmt.Printf("fish by id error: %v", err)
		}
		fish = *row1
	}
	if fish.Level == 0 { // No fishmob found
		w.Write(nil)
		return
	}

	if fish.AreaId != 0 {
		db.QueryRow("SELECT q.name FROM fishing_area as q WHERE q.zoneid = ? && q.areaid = ?", zID, fish.AreaId).Scan(&fish.AreaName)
	}

	jsonData, _ := json.Marshal(fish)
	w.Write(jsonData)
}

var jobGrades = [23][9]int16{
	//HP,MP,STR,DEX,VIT,AGI,INT,MND,CHR
	{0, 0, 0, 0, 0, 0, 0, 0, 0}, //NON
	{2, 0, 1, 3, 4, 3, 6, 6, 5}, //WAR
	{1, 0, 3, 2, 1, 6, 7, 4, 5}, //MNK
	{5, 3, 4, 6, 4, 5, 5, 1, 3}, //WHM
	{6, 2, 6, 3, 6, 3, 1, 5, 4}, //BLM
	{4, 4, 4, 4, 5, 5, 3, 3, 4}, //RDM
	{4, 0, 4, 1, 4, 2, 3, 7, 7}, //THF
	{3, 6, 2, 5, 1, 7, 7, 3, 3}, //PLD
	{3, 6, 1, 3, 3, 4, 3, 7, 7}, //DRK
	{3, 0, 4, 3, 4, 6, 5, 5, 1}, //BST
	{4, 0, 4, 4, 4, 6, 4, 4, 2}, //BRD
	{5, 0, 5, 4, 4, 1, 5, 4, 5}, //RNG
	{2, 0, 3, 3, 3, 4, 5, 5, 4}, //SAM
	{4, 0, 3, 2, 3, 2, 4, 7, 6}, //NIN
	{3, 0, 2, 4, 3, 4, 6, 5, 3}, //DRG
	{7, 1, 6, 5, 6, 4, 2, 2, 2}, //SMN
	{4, 4, 5, 5, 5, 5, 5, 5, 5}, //BLU
	{4, 0, 5, 3, 5, 2, 3, 5, 5}, //COR
	{4, 0, 5, 2, 4, 3, 5, 6, 3}, //PUP
	{4, 0, 4, 3, 5, 2, 6, 6, 2}, //DNC
	{5, 4, 6, 4, 5, 4, 3, 4, 3}, //SCH
	{3, 2, 6, 4, 5, 4, 3, 3, 4}, //GEO
	{3, 6, 3, 4, 5, 2, 4, 4, 6}, //RUN
}

// MOST CALC STUFF COMES FROM MOBUTILS
func GetJobGrade(job byte, stat int16) int16 {
	return jobGrades[job][stat]
}

func GetBaseToRank(rank int16, lvl int) int {
	switch rank {
	case 1:
		return (5 + ((lvl-1)*50)/100) // A
	case 2:
		return (4 + ((lvl-1)*45)/100) // B
	case 3:
		return (4 + ((lvl-1)*40)/100) // C
	case 4:
		return (3 + ((lvl-1)*35)/100) // D
	case 5:
		return (3 + ((lvl-1)*30)/100) // E
	case 6:
		return (2 + ((lvl-1)*25)/100) // F
	case 7:
		return (2 + ((lvl-1)*20)/100) // G
	}
	return 0
}

func GetMLevel(PMob MobGroup) int {
	if PMob.MaxLevel >= PMob.MinLevel {
		return int(PMob.MinLevel) + int(PMob.MaxLevel-PMob.MinLevel) // get stats based on midway between
	} else {
		fmt.Println("min level > max level")
		return 10
	}
}
func GetStats(PMob MobGroup) *MobGroup {
	mLvl := GetMLevel(PMob)

	fSTR := GetBaseToRank(PMob.STR, mLvl)
	fDEX := GetBaseToRank(PMob.DEX, mLvl)
	fVIT := GetBaseToRank(PMob.VIT, mLvl)
	fAGI := GetBaseToRank(PMob.AGI, mLvl)
	fINT := GetBaseToRank(PMob.INT, mLvl)
	fMND := GetBaseToRank(PMob.MND, mLvl)
	fCHR := GetBaseToRank(PMob.CHR, mLvl)

	mSTR := GetBaseToRank(GetJobGrade(PMob.MJob, 2), mLvl)
	mDEX := GetBaseToRank(GetJobGrade(PMob.MJob, 3), mLvl)
	mVIT := GetBaseToRank(GetJobGrade(PMob.MJob, 4), mLvl)
	mAGI := GetBaseToRank(GetJobGrade(PMob.MJob, 5), mLvl)
	mINT := GetBaseToRank(GetJobGrade(PMob.MJob, 6), mLvl)
	mMND := GetBaseToRank(GetJobGrade(PMob.MJob, 7), mLvl)
	mCHR := GetBaseToRank(GetJobGrade(PMob.MJob, 8), mLvl)

	sSTR := GetBaseToRank(GetJobGrade(PMob.SJob, 2), mLvl)
	sDEX := GetBaseToRank(GetJobGrade(PMob.SJob, 3), mLvl)
	sVIT := GetBaseToRank(GetJobGrade(PMob.SJob, 4), mLvl)
	sAGI := GetBaseToRank(GetJobGrade(PMob.SJob, 5), mLvl)
	sINT := GetBaseToRank(GetJobGrade(PMob.SJob, 6), mLvl)
	sMND := GetBaseToRank(GetJobGrade(PMob.SJob, 7), mLvl)
	sCHR := GetBaseToRank(GetJobGrade(PMob.SJob, 8), mLvl)

	// subjob stat scaling

	if mLvl < 16 { // 1~15 is just 25%
		sSTR /= 4
		sDEX /= 4
		sAGI /= 4
		sINT /= 4
		sMND /= 4
		sCHR /= 4
		sVIT /= 4
	} else if mLvl < 45 { // 16~44 linear scaling 25% -> 50%
		coe := 0.25 + (float32(mLvl)-15.0)*(25.0/30.0)/100.0
		sSTR = (int)(float32(sSTR) * coe)
		sDEX = (int)(float32(sDEX) * coe)
		sAGI = (int)(float32(sAGI) * coe)
		sINT = (int)(float32(sINT) * coe)
		sMND = (int)(float32(sMND) * coe)
		sCHR = (int)(float32(sCHR) * coe)
		sVIT = (int)(float32(sVIT) * coe)
	} else { // 45+ is just 50%
		sSTR /= 2
		sDEX /= 2
		sAGI /= 2
		sINT /= 2
		sMND /= 2
		sCHR /= 2
		sVIT /= 2
	}

	PMob.STR = int16(fSTR + mSTR + sSTR)
	PMob.DEX = int16(fDEX + mDEX + sDEX)
	PMob.VIT = int16(fVIT + mVIT + sVIT)
	PMob.AGI = int16(fAGI + mAGI + sAGI)
	PMob.INT = int16(fINT + mINT + sINT)
	PMob.MND = int16(fMND + mMND + sMND)
	PMob.CHR = int16(fCHR + mCHR + sCHR)

	PMob.EVA = int16(GetEvasion(PMob))
	PMob.DEF = int16(GetBase(PMob, PMob.DEF))
	PMob.ATT = int16(GetBase(PMob, PMob.ATT))
	PMob.ACC = int16(GetBase(PMob, PMob.ACC))

	PMob.MEVA = int16(GetMagicEvasion(PMob))
	PMob.CmbDmgMult = int16(GetWeaponDamage(PMob))

	if PMob.MobType == 2 { //if NM (if multiple types, wont work)
		PMob.EVA += int16(mLvl / 5)
		PMob.ACC += int16(mLvl / 5)
		// PMob.MACC += int16(mLvl / 10);
		PMob.MEVA += int16(mLvl / 10)
		PMob.ATT += int16(mLvl / 4)
		PMob.STR += int16(mLvl / 4)
		PMob.CmbDmgMult = int16(GetWeaponDamage(PMob) + mLvl/4)
		PMob.DEF += 20
	}

	switch PMob.FamilyId {
	case 72: // colibri
		PMob.EVA += 5
	case 136: // goobbue
		PMob.ATT += 10
	case 208: // ram
		PMob.ATT += 10
	case 242: // tiger
		PMob.ATT += 10
	case 179: // manticore
		PMob.ATT += 10
	case 217: // scorp
		PMob.ATT += 20
	case 240: // tauri
		PMob.DEF -= 20
		PMob.EVA -= 10
	case 57: // buffalo
		PMob.DEF += 20
		PMob.ATT += 10
	case 58: // bugard
		PMob.DEF += 20
		PMob.ATT += 10
	case 26: // antlion
		PMob.DEF += 20
		PMob.ATT += 10
	case 357: // antlion
		PMob.DEF += 20
		PMob.ATT += 10
	case 188: // opo
		PMob.EVA += 25
	case 253: // wamoura
		PMob.DEF += 12
	case 176: // mamoolja
		PMob.EVA += 10
	case 177: // mamoolja
		PMob.EVA += 10
	case 285: // mamoolja
		PMob.EVA += 10
	case 233: // soulflayer
		PMob.DEF += 20
	case 74: // corse
		PMob.DEF += 20
	case 64: // chigoe
		PMob.ATT -= 50
		PMob.DEF -= 20
		PMob.EVA += 10
	case 180: // marid
		PMob.EVA -= 10
	case 371: // marid
		PMob.EVA -= 10
	case 59: // bugbear
		PMob.EVA -= 10
	}

	return &PMob
}

func GetMagicEvasion(PMob MobGroup) uint16 {
	mEvaRank := int16(3) // default for all mobs

	return GetBase(PMob, mEvaRank)
}

func GetWeaponDamage(PMob MobGroup) int {
	damage := GetMLevel(PMob)

	damage = int(float64(damage) * float64(PMob.CmbDmgMult) / 100.0)
	// if MOBMOD_WEAPON_BONUS
	for i := 0; i < len(PMob.Modifs); i++ {
		if PMob.Modifs[i].Id == 59 && PMob.Modifs[i].IsMobMod > 0 {
			damage = int(float64(damage) * float64(PMob.Modifs[i].Val) / 100.0)
			break
		}
	}
	return damage
}

func GetEvasion(PMob MobGroup) uint16 {
	evaRank := PMob.EVA

	// Mob evasion is based on job
	// but occasionally war mobs
	// might have a different rank
	switch PMob.MJob {
	case 6:
		evaRank = 1
	case 13:
		evaRank = 1
	case 2:
		evaRank = 2
	case 19:
		evaRank = 2
	case 12:
		evaRank = 2
	case 18:
		evaRank = 2
	case 22:
		evaRank = 2
	case 5:
		evaRank = 4
	case 10:
		evaRank = 4
	case 21:
		evaRank = 4
	case 17:
		evaRank = 4
	case 3:
		evaRank = 5
	case 20:
		evaRank = 5
	case 11:
		evaRank = 5
	case 15:
		evaRank = 5
	case 4:
		evaRank = 5
	default:
		break
	}

	return GetBase(PMob, evaRank)
}

func GetBase(PMob MobGroup, rank int16) uint16 {
	lvl := GetMLevel(PMob)

	if lvl > 50 {
		switch rank {
		case 1: // A
			return (uint16)(153 + float64(lvl-50)*5.0)
		case 2: // B
			return (uint16)(147 + float64(lvl-50)*4.9)
		case 3: // C
			return (uint16)(136 + float64(lvl-50)*4.8)
		case 4: // D
			return (uint16)(126 + float64(lvl-50)*4.7)
		case 5: // E
			return (uint16)(116 + float64(lvl-50)*4.5)
		case 6: // F
			return (uint16)(106 + float64(lvl-50)*4.4)
		case 7: // G
			return (uint16)(96 + float64(lvl-50)*4.3)
		}
	} else {
		switch rank {
		case 1:
			return uint16(6 + float64(lvl-1)*3.0)
		case 2:
			return uint16(5 + float64(lvl-1)*2.9)
		case 3:
			return uint16(5 + float64(lvl-1)*2.8)
		case 4:
			return uint16(4 + float64(lvl-1)*2.7)
		case 5:
			return uint16(4 + float64(lvl-1)*2.5)
		case 6:
			return uint16(3 + float64(lvl-1)*2.4)
		case 7:
			return uint16(3 + float64(lvl-1)*2.3)
		}
	}

	fmt.Printf("Mobutils::GetBase rank (%d) is out of bounds for mob (%d) \n", rank, PMob.GroupId)
	return 0
}

func GetMobMods(mods *Mods) *Mods {
	mods.Str = ""
	temp := ""
	if mods.Val > 0 {
		temp = " +"
	}
	switch mods.Id {
	case 1:
		mods.Str = "Minimum Gil Drop: " + strconv.Itoa(mods.Val)
	case 2:
		mods.Str = "Maximum Gil Drop: " + strconv.Itoa(mods.Val)
	case 3:
		mods.Str = "Base MP for non-Mage " + temp + strconv.Itoa(mods.Val)
	case 4:
		mods.Str = "Sight Range: " + strconv.Itoa(mods.Val)
	case 5:
		mods.Str = "Sound Range: " + strconv.Itoa(mods.Val)
	case 6:
		mods.Str = "Chance to buff: " + temp + strconv.Itoa(mods.Val) + "%"
	case 7:
		mods.Str = "Chance to cast -ga spell: " + temp + strconv.Itoa(mods.Val) + "%"
	case 8:
		mods.Str = "Chance to heal: " + temp + strconv.Itoa(mods.Val) + "%"
	case 9:
		mods.Str = "Can cure below HP : " + strconv.Itoa(mods.Val) + "%"
	case 10:
		mods.Str = "Sub link group: " + strconv.Itoa(mods.Val)
	case 11:
		mods.Str = "Link Radius: " + strconv.Itoa(mods.Val)
	case 12:
		mods.Str = "Draw-In"
	case 76:
		mods.Str = "Draw-In Range: " + strconv.Itoa(mods.Val)
	case 77:
		mods.Str = "Draw-In Max Reach: " + strconv.Itoa(mods.Val)
	case 78:
		mods.Str = "Draw-In Party/Alliance"
	case 79:
		mods.Str = "Draw_In When Mob Cannot Attack"
	case 13:
		mods.Str = "Severe Spell Chance (e.g. Death): " + strconv.Itoa(mods.Val) + "%"
	case 14:
		mods.Str = "Uses skill list: " + strconv.Itoa(mods.Val) // THIS MAY NEED TO BE INVESTIGATED
	case 15:
		mods.Str = "Gil for mugging: " + strconv.Itoa(mods.Val)
	case 16:
		mods.Str = "Enmity from Dmg: " + temp + strconv.Itoa(mods.Val) + "%"
	case 17:
		mods.Str = "Does not despawn"
	case 18:
		// skip
	case 20:
		mods.Str = "Chance to use TP: " + temp + strconv.Itoa(mods.Val) + "%"
	case 21:
		mods.Str = "Uses pet spell list: " + strconv.Itoa(mods.Val)
	case 22:
		mods.Str = "Chance to cast -na spell: " + temp + strconv.Itoa(mods.Val) + "%"
	case 23:
		mods.Str = "Immunity to: " + strconv.Itoa(mods.Val)
	case 24:
		mods.Str = "Gradually Rages" //NOT IMPLEMENTED YET
	case 25:
		mods.Str = "Builds resistance to: " + strconv.Itoa(mods.Val) //NOT IMPLEMENTED YET
	case 27:
		mods.Str = "Uses spell list: " + strconv.Itoa(mods.Val)
	case 28:
		mods.Str = "EXP Bonus: " + temp + strconv.Itoa(mods.Val)
	case 29:
		mods.Str = "Mobs will assist"
	case 30:
		mods.Str = "Special Skill: " + strconv.Itoa(mods.Val)
	case 31:
		mods.Str = "Distance allowed to roam: " + strconv.Itoa(mods.Val)
	case 33:
		mods.Str = "Cool down for special: " + strconv.Itoa(mods.Val)
	case 34:
		mods.Str = "Cool down for magic: " + strconv.Itoa(mods.Val)
	case 35:
		mods.Str = "Cool down for standing back: " + strconv.Itoa(mods.Val)
	case 36:
		mods.Str = "Cool down between roams: " + strconv.Itoa(mods.Val)
	case 37:
		mods.Str = "Aggro regardless of level"
	case 38:
		mods.Str = "No Drops"
	case 39:
		mods.Str = "Shares Pos. with another mob"
	case 40:
		mods.Str = "Cool down between teleports: " + strconv.Itoa(mods.Val)
	case 41:
		mods.Str = "Teleport Start Skill: " + strconv.Itoa(mods.Val)
	case 42:
		mods.Str = "Teleport Stop Skill: " + strconv.Itoa(mods.Val)
	case 43:
		switch mods.Val {
		case 1:
			mods.Str = "Teleports on cool down"
		case 2:
			mods.Str = "Teleports to close distance"
		default:
			mods.Str = "Teleport Type: " + strconv.Itoa(mods.Val)
		}
	case 44:
		mods.Str = "Dual Wields"
	case 45:
		mods.Str = "Additional effect on attacks: " + strconv.Itoa(mods.Val)
	case 46:
		mods.Str = "Additional effect when hit: " + strconv.Itoa(mods.Val)
	case 47:
		mods.Str = "Max distance from spawn point: " + strconv.Itoa(mods.Val)
	case 48:
		mods.Str = "Shares target with mob: " + strconv.Itoa(mods.Val)
	case 49:
		mods.Str = "Checks as an NM"
	case 51:
		mods.Str = "Maximum roam turns: " + strconv.Itoa(mods.Val)
	case 52:
		// not worth adding Roaming frequency. roam_cool - rand(roam_cool / (roam_rate / 10))
	case 53:
		mods.Str = "Added behavior: " + strconv.Itoa(mods.Val)
	case 54:
		mods.Str = "Gil Bonus"
	case 55:
		mods.Str = "Time to despawn when idle: " + strconv.Itoa(mods.Val)
	case 56:
		mods.Str = "Stands back with hp above: " + strconv.Itoa(mods.Val) + "%"
	case 57:
		mods.Str = "Time between casting: " + strconv.Itoa(mods.Val)
	case 58:
		mods.Str = "Time before first special: " + strconv.Itoa(mods.Val)
	case 59:
		mods.Str = "Weapon Damage Bonus " + temp + strconv.Itoa(mods.Val/100)
	case 60:
		mods.Str = "Animation sub reset: " + strconv.Itoa(mods.Val)
	case 61:
		mods.Str = "HP Multiplier: HP x " + strconv.Itoa(mods.Val/100)
	case 62:
		mods.Str = "Mob never stands back"
	case 63:
		mods.Str = "Uses attack skill list: " + strconv.Itoa(mods.Val)
	case 64:
		mods.Str = "Charmable"
	case 65:
		mods.Str = "Mob cannot move"
	case 66:
		mods.Str = "Mob can swing " + strconv.Itoa(mods.Val) + " times"
	case 67:
		mods.Str = "Mob may not aggro"
	case 68:
		mods.Str = "Range from target for alliance enmity: " + strconv.Itoa(mods.Val)
	case 69:
		mods.Str = "Mob may not link"
	case 70:
		mods.Str = "MOb cannot regain HP"
	case 71:
		mods.Str = "BCNM links by sound"
	case 72:
		mods.Str = "Claim Shield"
	case 73:
		mods.Str = "Family will link"
	case 74:
		mods.Str = "Mob will link with family in zone"
	case 75:
		mods.Str = "Encroaches player"
	case 84:
		mods.Str = "Sight detection angle: " + strconv.Itoa(mods.Val)
	case 85:
		mods.Str = "Only aggro if player or member has hate"
	case 86:
		mods.Str = "Cures and Raises players"
	default: //nothing
	}
	return mods
}
