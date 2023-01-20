package ff

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Character struct {
	CharID int32
	AccID  int32
	Name   string
	Nation byte
	// PosZone            int32
	// PosPrevZone        int32
	// PosRot             int16
	// PosX               float32
	// PosY               float32
	// PosZ               float32
	// Moghouse           int32
	// Boundary           int16
	// HomeZone           int16
	// HomeRot            int16
	// HomeX              float32
	// HomeY              float32
	// HomeZ              float32
	// Missions           []byte
	// Assault            []byte
	// Campaign           []byte
	// Quests             []byte
	// KeyItems           []byte
	// SetBlueSpells      []byte
	// Abilities          []byte
	// Weaponskills       []byte
	// Titles             []byte
	// Zones              []byte
	Playtime   int32
	GMLevel    byte
	Mentor     byte
	MJob       byte
	SJob       byte
	MLvl       byte
	SLvl       byte
	War        byte
	Mnk        byte
	Whm        byte
	Blm        byte
	Rdm        byte
	Thf        byte
	Pld        byte
	Drk        byte
	Bst        byte
	Brd        byte
	Rng        byte
	Sam        byte
	Nin        byte
	Drg        byte
	Smn        byte
	Blu        byte
	Cor        byte
	Pup        byte
	Dnc        byte
	Sch        byte
	Gil        int64
	RankSandy  int16
	RankBastok int16
	RankWindy  int16
	FameSandy  int16
	FameBastok int16
	FameWindy  int16
	FameJeuno  int16
	FameNorg   int16
	Face       byte
	Race       byte
	Equipment  []*CharEquip
	Skills     []*Skill
	LS         []*Linkshell
}

type CharEquip struct {
	SlotId  byte
	Name    string
	CharID  int32
	ESlotId byte
	ContId  byte
	ItemId  int16
}

type Skill struct {
	SkillId byte
	Value   int16
}

type Linkshell struct {
	ItemId int32
	Sig    *string
}

func GetCharByName(w http.ResponseWriter, r *http.Request) {
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

	rows, err := db.Query("SELECT q.charid, q.accid, q.charname, q.nation, q.playtime, q.gmlevel, q.mentor, a.mjob, a.sjob, a.mlvl, a.slvl, b.war, b.mnk, b.whm, b.blm, b.rdm, b.thf, b.pld, b.drk, b.bst, b.brd, b.rng, b.sam, b.nin, b.drg, b.smn, b.blu, b.cor, b.pup, b.dnc, b.sch, c.quantity, d.rank_sandoria, d.rank_bastok, d.rank_windurst, d.fame_sandoria, d.fame_bastok, d.fame_windurst, d.fame_norg, d.fame_jeuno, e.face, e.race FROM chars as q JOIN `char_stats` as a ON q.charid = a.charid JOIN `char_jobs` as b ON q.charid = b.charid JOIN `char_inventory` as c ON q.charid = c.charid JOIN `char_profile` as d ON q.charid = d.charid JOIN `char_look` as e ON q.charid = e.charid WHERE q.charname =? AND c.itemId = 65535", sID)
	if err != nil {
		fmt.Println("error selecting char")
		panic(err.Error())
	}
	defer rows.Close()

	var characters []*Character
	for rows.Next() {
		row := new(Character)
		if err := rows.Scan(&row.CharID, &row.AccID, &row.Name, &row.Nation, &row.Playtime, &row.GMLevel, &row.Mentor, &row.MJob, &row.SJob, &row.MLvl, &row.SLvl, &row.War, &row.Mnk, &row.Whm, &row.Blm, &row.Rdm, &row.Thf, &row.Pld, &row.Drk, &row.Bst, &row.Brd, &row.Rng, &row.Sam, &row.Nin, &row.Drg, &row.Smn, &row.Blu, &row.Cor, &row.Pup, &row.Dnc, &row.Sch, &row.Gil, &row.RankSandy, &row.RankBastok, &row.RankWindy, &row.FameSandy, &row.FameBastok, &row.FameWindy, &row.FameNorg, &row.FameJeuno, &row.Face, &row.Race); err != nil {
			fmt.Printf("chars by name error: %v", err)
		}
		characters = append(characters, row)
		erows, err := db.Query("SELECT q.charid, q.slotid, q.equipslotid, q.containerid, c.itemId, d.name FROM char_equip as q JOIN char_inventory as c ON q.charid = c.charid JOIN item_equipment as d ON c.itemId = d.itemId WHERE q.charid = ? AND q.equipslotid IN (0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15) AND q.containerid = c.location AND q.slotid = c.slot", row.CharID)
		if err != nil {
			fmt.Println("error selecting charequip")
			panic(err.Error())
		}
		defer erows.Close()

		for erows.Next() {
			row2 := new(CharEquip)
			if err := erows.Scan(&row2.CharID, &row2.SlotId, &row2.ESlotId, &row2.ContId, &row2.ItemId, &row2.Name); err != nil {
				fmt.Printf("charequip by error: %v", err)
			}
			row.Equipment = append(row.Equipment, row2)
		}

		srows, err := db.Query("SELECT q.skillid, q.value FROM char_skills as q WHERE q.charid = ?", row.CharID)
		if err != nil {
			fmt.Println("error selecting skills")
			panic(err.Error())
		}
		defer srows.Close()

		for srows.Next() {
			row3 := new(Skill)
			if err := srows.Scan(&row3.SkillId, &row3.Value); err != nil {
				fmt.Printf("skills by error: %v", err)
			}
			row.Skills = append(row.Skills, row3)
		}

		lrows, err := db.Query("SELECT signature, itemId FROM `char_inventory` WHERE charid = ? AND location = 0 AND slot = (SELECT slotid FROM `char_equip` WHERE charid = ? AND equipslotid = 16)", row.CharID, row.CharID)
		if err != nil {
			fmt.Println("error selecting ls")
			panic(err.Error())
		}
		defer lrows.Close()

		for lrows.Next() {
			row4 := new(Linkshell)
			if err := lrows.Scan(&row4.Sig, &row4.ItemId); err != nil {
				fmt.Printf("ls by error: %v", err)
			}
			row.LS = append(row.LS, row4)
		}
	}

	jsonData, _ := json.Marshal(&characters)
	w.Write(jsonData)
}
