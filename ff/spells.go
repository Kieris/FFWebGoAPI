package ff

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Spell struct {
	SpellId     int32
	Name        *string
	AOE         byte
	Jobs        []byte
	Level       byte
	CastTime    int16
	Element     byte
	RecastTime  int32
	Range       int16
	ValidTarget byte
	Cost        int16
	ContentTag  *string
	CE          int32
	VE          int32
	Skill       byte
	Base        int16
	Multiplier  float32
	Note        string
	Blue        *BlueSpell
}

type BlueSpell struct {
	SpellId        int32
	MobSkill       int16
	SetPoints      int16
	TraitCat       int16
	TraitCatWeight int16
	PrimarySC      int16
	SecondarySC    int16
	Modifs         []*Mods
	LearnFrom      []*string
	PointsNeeded   byte
	Mod            Mods
	Notes          []string
}

func GetSpellsByJob(w http.ResponseWriter, r *http.Request) {
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
	var spells []*Spell
	var grp int
	var str string
	grp = GetSpellGrpByJob(sID)
	if grp == 0 {
		jsonData, _ := json.Marshal(&spells)
		w.Write(jsonData)
		return
	} else if grp == 10 {
		str = "(2, 6)"
	} else {
		str = "(" + strconv.Itoa(grp) + ")"
	}

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		fmt.Println("error validating sql.Open arguments")
		panic(err.Error())
	}
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("%s%s", "SELECT q.spellid, q.name, q.aoe, q.element, q.mpCost, q.Jobs, q.spell_range, q.castTime, q.recastTime, q.skill, q.base, q.multiplier, q.CE, q.VE, q.validTargets, q.content_tag FROM spell_list as q WHERE q.group IN ", str))
	if err != nil {
		fmt.Println("error selecting spell by job")
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		row3 := new(Spell)
		if err := rows.Scan(&row3.SpellId, &row3.Name, &row3.AOE, &row3.Element, &row3.Cost, &row3.Jobs, &row3.Range, &row3.CastTime, &row3.RecastTime, &row3.Skill, &row3.Base, &row3.Multiplier, &row3.CE, &row3.VE, &row3.ValidTarget, &row3.ContentTag); err != nil {
			fmt.Printf("job spell error: %v", err)
		}
		row3.Level = GetSpellJobLevel(row3.Jobs, sID)
		row3.Note = SpellDes[strconv.Itoa(int(row3.SpellId))]

		if row3.Level > 0 && row3.Level <= lvlcap {
			spells = append(spells, row3)
		}
	}

	sort.Slice(spells[:], func(i, j int) bool {
		return spells[i].Level < spells[j].Level
	})

	jsonData, _ := json.Marshal(&spells)
	w.Write(jsonData)
}

// rdm,whm,sch,pld   grp 6
// blm, rdm, sch, drk, geo  grp 2
// smn 5
// nin 4
// brd 1
// blu 3
// trust 8
// This will narrow downs spell list before sorting out BLOB data
func GetSpellGrpByJob(jobID int) int {
	switch jobID {
	case 3:
		return 6
	case 7:
		return 6
	case 15:
		return 5
	case 10:
		return 1
	case 16:
		return 3
	case 4:
		return 2
	case 8:
		return 2
	case 21:
		return 2
	case 5:
		return 10
	case 20:
		return 10
	case 13:
		return 4
	default:
		return 0
	}
}

func GetSpellJobLevel(arr []byte, jobId int) byte {
	return arr[jobId-1]
}

func GetMiscBluNotes(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	InitHeader(w)
	jID := pathParams["jID"]

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		fmt.Println("error validating sql.Open arguments")
		panic(err.Error())
	}
	defer db.Close()
	var id int32
	rowb := new(BlueSpell)
	currMod := new(Mods)
	db.QueryRow("SELECT spellid FROM spell_list WHERE name = ?", jID).Scan(&id)

	db.QueryRow("SELECT q.mob_skill_id, q.set_points, q.trait_category, q.trait_category_weight, q.primary_sc, q.secondary_sc, a.trait_points_needed, a.modifier, a.value FROM blue_spell_list as q JOIN blue_traits as a ON q.trait_category = a.trait_category WHERE q.spellid = ?", id).Scan(&rowb.MobSkill, &rowb.SetPoints, &rowb.TraitCat, &rowb.TraitCatWeight, &rowb.PrimarySC, &rowb.SecondarySC, &rowb.PointsNeeded, &currMod.Id, &currMod.Val)
	if err != nil {
		fmt.Println("error selecting blue spell")
		panic(err.Error())
	}
	currMod = GetMods(currMod)
	rowb.Mod = *currMod
	// Get Mods
	mrows, err := db.Query("SELECT q.modid, q.value FROM blue_spell_mods as q WHERE q.spellid = ?", rowb.SpellId)
	if err != nil {
		fmt.Println("error selecting mods for bluespell")
		panic(err.Error())
	}
	defer mrows.Close()

	var mds []*Mods
	for mrows.Next() {
		mrow := new(Mods)
		if err := mrows.Scan(&mrow.Id, &mrow.Val); err != nil {
			fmt.Printf("blue mod error: %v", err)
		}
		mrow = GetMods(mrow)
		mds = append(mds, mrow)
	}

	// Get Mobs spell is learned from
	qrows, err := db.Query("SELECT q.skill_list_name FROM mob_skill_lists as q WHERE q.mob_skill_id = ?", rowb.MobSkill)
	if err != nil {
		fmt.Println("error selecting mobs for bluespell")
		panic(err.Error())
	}
	defer qrows.Close()

	var strs []*string
	for qrows.Next() {
		qrow := new(string)
		if err := qrows.Scan(&qrow); err != nil {
			fmt.Printf("blue mob error: %v", err)
		}
		strs = append(strs, qrow)
	}
	rowb.SpellId = id
	rowb.LearnFrom = strs
	rowb.Modifs = mds
	rowb.Notes = GetScriptDets("spells/bluemagic/" + jID)
	jsonData, _ := json.Marshal(rowb)
	w.Write(jsonData)
}
