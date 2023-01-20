package ff

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Pet struct {
	Name      *string
	PoolId    int16
	PetId     int16
	MinLvl    byte
	MaxLvl    byte
	Time      int
	Elem      byte
	JugItemId int32
	JugName   *string
}

type AutomatonSpell struct {
	SkillLvl int16
	RemStr   string
	HeadStr  string
	Name     *string
}

type PupItem struct {
	ItemId  int16
	Slot    byte
	Elem    int
	Name    string
	Note    string
	ElemTot PupElems
	Mods    []PupMod
}

type PupElems struct {
	Fire      byte
	Ice       byte
	Wind      byte
	Earth     byte
	Lightning byte
	Water     byte
	Light     byte
	Dark      byte
}

type PupSkills struct {
	Melee  *ValByte
	Magic  *ValByte
	Ranged *ValByte
}

type PetDets struct {
	Mob    *MobGroup
	Skills []*PetSkill
}
type PetSkill struct {
	SkillId      int16
	Note         string
	Name         string
	Level        byte
	AddType      int16
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

func GetPetsByJob(w http.ResponseWriter, r *http.Request) {
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
	var seaStr string
	var pets []*Pet
	if jobsCapped[sID] == "BST" {
		seaStr = "element = 0 AND time != 0"
	} else if jobsCapped[sID] == "SMN" {
		seaStr = "element > 0 AND element <= 8"
	} else if jobsCapped[sID] == "DRG" {
		seaStr = "petid = 48"
	} else if jobsCapped[sID] == "PUP" {
		seaStr = "petid > 68 AND petid <= 72"
	} else {
		jsonData, _ := json.Marshal(&pets)
		w.Write(jsonData)
		return
	}

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		fmt.Println("error validating sql.Open arguments")
		panic(err.Error())
	}
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("SELECT petid, name, poolid, minLevel, maxLevel, time, element FROM pet_list WHERE %s AND minLevel <= %d ORDER BY minLevel", seaStr, lvlcap))
	if err != nil {
		fmt.Println("error selecting pet by job")
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		row3 := new(Pet)
		if err := rows.Scan(&row3.PetId, &row3.Name, &row3.PoolId, &row3.MinLvl, &row3.MaxLvl, &row3.Time, &row3.Elem); err != nil {
			fmt.Printf("job pet error: %v", err)
		}

		if jobsCapped[sID] == "BST" {
			db.QueryRow("SELECT itemId, name FROM item_weapon WHERE subskill = ?", row3.PetId).Scan(&row3.JugItemId, &row3.JugName)
		}

		pets = append(pets, row3)
	}

	jsonData, _ := json.Marshal(&pets)
	w.Write(jsonData)
}

func GetPetByID(w http.ResponseWriter, r *http.Request) {
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
	items := GetPetDetails(sID)
	jsonData, _ := json.Marshal(&items)
	w.Write(jsonData)
}

func GetPetDetails(sID int) []*PetDets {
	var items []*PetDets
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		fmt.Println("error validating sql.Open arguments")
		panic(err.Error())
	}
	defer db.Close()

	rows, err := db.Query("SELECT b.name, b.familyid, b.mJob, b.sJob, b.cmbSkill, b.cmbDelay, b.cmbDmgMult, b.aggro, b.true_detection, b.links, b.mobType, b.immunity, b.name_prefix, b.flag, b.entityFlags, b.spellList, b.roamflag, b.skill_list_id, c.family, c.systemid, c.system, c.mobsize, c.speed, c.HP, c.MP, c.STR, c.DEX, c.VIT, c.AGI, c.INT, c.MND, c.CHR, c.ATT, c.DEF, c.ACC, c.EVA, c.Slash, c.Pierce, c.H2H, c.Impact, c.Fire, c.Ice, c.Wind, c.Earth, c.Lightning, c.Water, c.Light, c.Dark, c.Element, c.detects, c.charmable FROM mob_pools as b JOIN mob_family_system as c ON b.familyid = c.familyid WHERE b.poolid = ?", sID)
	if err != nil {
		fmt.Println("error selecting mob group")
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		row := new(PetDets)
		row.Mob = new(MobGroup)
		row.Mob.MinLevel = 75
		row.Mob.MaxLevel = 75
		if err := rows.Scan(&row.Mob.Name, &row.Mob.FamilyId, &row.Mob.MJob, &row.Mob.SJob, &row.Mob.CmbSkill, &row.Mob.CmbDelay, &row.Mob.CmbDmgMult, &row.Mob.Aggro, &row.Mob.TrueSight, &row.Mob.Links, &row.Mob.MobType, &row.Mob.Immunity, &row.Mob.NamePrefix, &row.Mob.Flag, &row.Mob.EntityFlags, &row.Mob.SpellList, &row.Mob.RoamFlag, &row.Mob.SkillList, &row.Mob.Family, &row.Mob.SystemId, &row.Mob.Genre, &row.Mob.MobSize, &row.Mob.Speed, &row.Mob.FHP, &row.Mob.FMP, &row.Mob.STR, &row.Mob.DEX, &row.Mob.VIT, &row.Mob.AGI, &row.Mob.INT, &row.Mob.MND, &row.Mob.CHR, &row.Mob.ATT, &row.Mob.DEF, &row.Mob.ACC, &row.Mob.EVA, &row.Mob.Slash, &row.Mob.Pierce, &row.Mob.H2H, &row.Mob.Impact, &row.Mob.Fire, &row.Mob.Ice, &row.Mob.Wind, &row.Mob.Earth, &row.Mob.Lightning, &row.Mob.Water, &row.Mob.Light, &row.Mob.Dark, &row.Mob.Element, &row.Mob.Detects, &row.Mob.Charmable); err != nil {
			fmt.Printf("pet by pool error: %v", err)
		}

		// Start getting skills
		erows, err := db.Query("SELECT q.mob_skill_id FROM mob_skill_lists as q WHERE q.skill_list_id = ?", row.Mob.SkillList)
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

			srows, err := db.Query("SELECT q.mob_skill_id, q.mob_skill_name, q.mob_skill_aoe, q.mob_skill_distance, q.mob_prepare_time, q.mob_valid_targets, q.mob_skill_flag, q.mob_skill_param, q.knockback, q.primary_sc, q.secondary_sc, q.tertiary_sc, d.level, d.addType FROM mob_skills as q JOIN abilities as d ON q.mob_skill_name = d.name WHERE q.mob_skill_id = ?", row2.SkillId)
			if err != nil {
				fmt.Println("error selecting pet skill")
				panic(err.Error())
			}
			defer srows.Close()
			for srows.Next() {
				row3 := new(PetSkill)
				if err := srows.Scan(&row3.SkillId, &row3.Name, &row3.AOE, &row3.Distance, &row3.PrepTime, &row3.ValidTargets, &row3.Flag, &row3.Param, &row3.Knockback, &row3.PrimSC, &row3.SecSC, &row3.TertSC, &row3.Level, &row3.AddType); err != nil {
					fmt.Printf("pet skill error: %v", err)
				}
				row3.Note = AbilityDes[row3.Name].Value
				if row3.Note != "" {
					row3.Name = AbilityDes[row3.Name].Name
				}
				row.Skills = append(row.Skills, row3)
			}
			// query can't sort joined data properly so di it manually
			sort.Slice(row.Skills[:], func(i, j int) bool {
				return row.Skills[i].Level < row.Skills[j].Level
			})
		}
		//Done getting skills

		// Start getting spells
		arows, err := db.Query("SELECT q.spell_id, q.min_level, q.max_level FROM mob_spell_lists as q WHERE q.spell_list_id = ? AND q.max_level >= ?", row.Mob.SpellList, lvlcap)
		if err != nil {
			fmt.Println("error selecting pet spell list")
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
				fmt.Println("error selecting pet spell")
				panic(err.Error())
			}
			defer srows.Close()
			for srows.Next() {
				row3 := new(Spell)
				if err := srows.Scan(&row3.SpellId, &row3.Name, &row3.AOE, &row3.Jobs, &row3.Range, &row3.CastTime, &row3.RecastTime, &row3.Skill, &row3.Base, &row3.Multiplier, &row3.CE, &row3.VE, &row3.ValidTarget, &row3.ContentTag); err != nil {
					fmt.Printf("pet spell error: %v", err)
				}
				row.Mob.Spells = append(row.Mob.Spells, row3)
			}
		}
		//Done getting spells

		// Get Mods
		mrows, err := db.Query("SELECT q.modid, q.value, q.is_mob_mod FROM mob_pool_mods as q WHERE q.poolid = ?", sID)
		if err != nil {
			fmt.Println("error selecting mods for pool")
			panic(err.Error())
		}
		defer mrows.Close()
		var mds []*Mods
		for mrows.Next() {
			mrow := new(Mods)
			if err := mrows.Scan(&mrow.Id, &mrow.Val, &mrow.IsMobMod); err != nil {
				fmt.Printf("pet pool mod error: %v", err)
			}
			if mrow.IsMobMod != 0 {
				mrow = GetMobMods(mrow)
			} else {
				mrow = GetMods(mrow)
			}
			mds = append(mds, mrow)
		}

		prows, err := db.Query("SELECT q.modid, q.value, q.is_mob_mod FROM mob_family_mods as q WHERE q.familyid = ?", row.Mob.FamilyId)
		if err != nil {
			fmt.Println("error selecting mods for group")
			panic(err.Error())
		}
		defer prows.Close()
		for prows.Next() {
			prow := new(Mods)
			if err := prows.Scan(&prow.Id, &prow.Val, &prow.IsMobMod); err != nil {
				fmt.Printf("pet fam mod error: %v", err)
			}
			if prow.IsMobMod != 0 {
				prow = GetMobMods(prow)
			} else {
				prow = GetMods(prow)
			}
			mds = append(mds, prow)
		}
		row.Mob.Modifs = mds
		row.Mob = GetStats(*row.Mob)
		// end getting Mods
		items = append(items, row)
	}
	return items
}

func GetPupSpells(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", corStr)

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		fmt.Println("error validating sql.Open arguments")
		panic(err.Error())
	}
	defer db.Close()

	rows, err := db.Query("SELECT a.name, q.skilllevel, q.heads, q.removes FROM automaton_spells as q JOIN spell_list as a ON q.spellid = a.spellid ORDER BY q.skilllevel")
	if err != nil {
		fmt.Println("error selecting pup spell")
		panic(err.Error())
	}
	defer rows.Close()
	var spells []*AutomatonSpell
	for rows.Next() {
		var heads int
		var removes int
		row3 := new(AutomatonSpell)
		if err := rows.Scan(&row3.Name, &row3.SkillLvl, &heads, &removes); err != nil {
			fmt.Printf("pup spell error: %v", err)
		}
		for removes > 0 {
			str := GetRemoveVal(removes & 0xFF)
			row3.RemStr += str
			removes = removes >> 8
			if removes > 0 {
				row3.RemStr += ", "
			}
		}
		row3.HeadStr = GetHead(heads)
		spells = append(spells, row3)
	}

	jsonData, _ := json.Marshal(&spells)
	w.Write(jsonData)
}

func GetPupFrames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", corStr)

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		fmt.Println("error validating sql.Open arguments")
		panic(err.Error())
	}
	defer db.Close()

	rows, err := db.Query("SELECT q.itemid, q.name, q.slot, q.element FROM item_puppet as q WHERE itemid < 8449")
	if err != nil {
		fmt.Println("error selecting pup frame")
		panic(err.Error())
	}
	defer rows.Close()
	var frames []*PupItem
	for rows.Next() {
		row3 := new(PupItem)
		if err := rows.Scan(&row3.ItemId, &row3.Name, &row3.Slot, &row3.Elem); err != nil {
			fmt.Printf("pup frame error: %v", err)
		}
		row3.Note = ItemDes[strconv.Itoa(int(row3.ItemId))]

		for i := 0; i < 8; i++ {
			// fmt.Printf("%d : %d\n", i, row3.Elem>>(i*4)&0xF)
			switch i {
			case 0:
				row3.ElemTot.Fire = byte(row3.Elem >> (i * 4) & 0xF)
			case 1:
				row3.ElemTot.Ice = byte(row3.Elem >> (i * 4) & 0xF)
			case 2:
				row3.ElemTot.Wind = byte(row3.Elem >> (i * 4) & 0xF)
			case 3:
				row3.ElemTot.Earth = byte(row3.Elem >> (i * 4) & 0xF)
			case 4:
				row3.ElemTot.Lightning = byte(row3.Elem >> (i * 4) & 0xF)
			case 5:
				row3.ElemTot.Water = byte(row3.Elem >> (i * 4) & 0xF)
			case 6:
				row3.ElemTot.Light = byte(row3.Elem >> (i * 4) & 0xF)
			case 7:
				row3.ElemTot.Dark = byte(row3.Elem >> (i * 4) & 0xF)
			}
		}

		frames = append(frames, row3)
	}

	jsonData, _ := json.Marshal(&frames)
	w.Write(jsonData)
}

func GetPupAttachments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", corStr)
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		fmt.Println("error validating sql.Open arguments")
		panic(err.Error())
	}
	defer db.Close()

	rows, err := db.Query("SELECT q.itemid, q.name, q.slot, q.element FROM item_puppet as q WHERE itemid > 8448")
	if err != nil {
		fmt.Println("error selecting pup attachment")
		panic(err.Error())
	}
	defer rows.Close()
	var frames []*PupItem
	for rows.Next() {
		row3 := new(PupItem)
		if err := rows.Scan(&row3.ItemId, &row3.Name, &row3.Slot, &row3.Elem); err != nil {
			fmt.Printf("pup frame attachment: %v", err)
		}
		if AttachmentMods[row3.Name] != nil {
			row3.Mods = AttachmentMods[row3.Name]
		}
		row3.Note = ItemDes[strconv.Itoa(int(row3.ItemId))]
		for i := 0; i < 8; i++ {
			// fmt.Printf("%d : %d\n", i, row3.Elem>>(i*4)&0xF)
			switch i {
			case 0:
				row3.ElemTot.Fire = byte(row3.Elem >> (i * 4) & 0xF)
			case 1:
				row3.ElemTot.Ice = byte(row3.Elem >> (i * 4) & 0xF)
			case 2:
				row3.ElemTot.Wind = byte(row3.Elem >> (i * 4) & 0xF)
			case 3:
				row3.ElemTot.Earth = byte(row3.Elem >> (i * 4) & 0xF)
			case 4:
				row3.ElemTot.Lightning = byte(row3.Elem >> (i * 4) & 0xF)
			case 5:
				row3.ElemTot.Water = byte(row3.Elem >> (i * 4) & 0xF)
			case 6:
				row3.ElemTot.Light = byte(row3.Elem >> (i * 4) & 0xF)
			case 7:
				row3.ElemTot.Dark = byte(row3.Elem >> (i * 4) & 0xF)
			}
		}

		frames = append(frames, row3)
	}

	jsonData, _ := json.Marshal(&frames)
	w.Write(jsonData)
}

func GetPupSkillRanks(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", corStr)

	var err error
	lID := -1
	if val, ok := pathParams["lID"]; ok {
		lID, err = strconv.Atoi(val)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "need a number"}`))
			return
		}
	}

	hID := -1
	if val, ok := pathParams["hID"]; ok {
		hID, err = strconv.Atoi(val)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "need a number"}`))
			return
		}
	}

	fID := -1
	if val, ok := pathParams["fID"]; ok {
		fID, err = strconv.Atoi(val)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "need a number"}`))
			return
		}
	}
	skills := new(PupSkills)
	skills.Melee = GetPetSkillCap(hID, fID, 22, lID)[0]
	skills.Magic = GetPetSkillCap(hID, fID, 24, lID)[0]
	skills.Ranged = GetPetSkillCap(hID, fID, 23, lID)[0]

	jsonData, _ := json.Marshal(&skills)
	w.Write(jsonData)
}

// PSpell.heads & (1 << ((uint8)PCaster->getHead() - 1))
func GetHead(num int) string {
	str := ""
	if num&(1<<0) > 0 {
		str += "Harlequin, " //8193
	}
	if num&(1<<1) > 0 {
		str += "Valoredge, " //8194
	}
	if num&(1<<2) > 0 {
		str += "Sharpshot, " //8195
	}
	if num&(1<<3) > 0 {
		str += "Stormwaker, " //8196
	}
	if num&(1<<4) > 0 {
		str += "Soulsoother, " //8197
	}
	if num&(1<<5) > 0 {
		str += "Spiritreaver, " //8198
	}
	str = strings.TrimSuffix(str, ", ")
	return str
}

func GetRemoveVal(num int) string {
	switch num {
	case 19:
		return "Sleep II"
	case 2:
		return "Sleep"
	case 3:
		return "Poison"
	case 4:
		return "Paralyze"
	case 5:
		return "Blind"
	case 6:
		return "Silence"
	case 7:
		return "Petrification"
	case 8:
		return "Disease"
	case 9:
		return "Curse"
	case 20:
		return "Curse II"
	case 30:
		return "Bane"
	case 31:
		return "Plague"
	case 193:
		return "Lullaby"
	case 594974:
		return "Curse"
	case 2079:
		return "Virus"
	default:
		return ""
	}
}

func GetPetSkillCap(head int, frame int, skill int, level int) []*ValByte {
	rank := 0
	SKILL_AUTOMATON_MELEE := 22
	SKILL_AUTOMATON_MAGIC := 24
	SKILL_AUTOMATON_RANGED := 23
	FRAME_VALOREDGE := 8225
	FRAME_SHARPSHOT := 8226
	FRAME_STORMWAKER := 8227
	HEAD_VALOREDGE := 8194
	HEAD_SHARPSHOT := 8195
	HEAD_STORMWAKER := 8196
	HEAD_SOULSOOTHER := 8197
	HEAD_SPIRITREAVER := 8198
	switch frame {
	case FRAME_VALOREDGE:
		if skill == SKILL_AUTOMATON_MELEE {
			rank = 3
		}
	case FRAME_SHARPSHOT:
		if skill == SKILL_AUTOMATON_MELEE {
			rank = 6
		} else if skill == SKILL_AUTOMATON_RANGED {
			rank = 3
		}
	case FRAME_STORMWAKER:
		if skill == SKILL_AUTOMATON_MELEE {
			rank = 7
		} else if skill == SKILL_AUTOMATON_MAGIC {
			rank = 3
		}
	default: //case FRAME_HARLEQUIN:
		rank = 5
	}

	switch head {
	case HEAD_VALOREDGE:
		if skill == SKILL_AUTOMATON_MELEE {
			rank -= 1
		}
	case HEAD_SHARPSHOT:
		if skill == SKILL_AUTOMATON_RANGED {
			rank -= 1
		}
	case HEAD_STORMWAKER:
		if skill == SKILL_AUTOMATON_MELEE || skill == SKILL_AUTOMATON_MAGIC {
			rank -= 1
		}
	case HEAD_SOULSOOTHER:
		if skill == SKILL_AUTOMATON_MAGIC {
			rank -= 2
		}
	case HEAD_SPIRITREAVER:
		if skill == SKILL_AUTOMATON_MAGIC {
			rank -= 2
		}
	default:
		break
	}

	//only happens if a head gives bonus to a rank of 0 - making it G or F rank
	if rank < 0 {
		rank = 13 + rank
	}
	return GetSkillRanks(level, rank)
}
