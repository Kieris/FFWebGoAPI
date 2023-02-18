package main

import (
	ffdb "goAPI/ff"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/abilities/{aID}", ffdb.GetAbility).Methods(http.MethodGet)
	api.HandleFunc("/zones", ffdb.GetZones).Methods(http.MethodGet)
	api.HandleFunc("/zones/mobs/{sID}", ffdb.GetZoneMobs).Methods(http.MethodGet)

	api.HandleFunc("/guildstore/{sID}", ffdb.GetStoreItems).Methods(http.MethodGet)
	api.HandleFunc("/jobs/{jID}/abilities", ffdb.GetAbilitiesByJob).Methods(http.MethodGet)
	api.HandleFunc("/weapons/{aID}", ffdb.GetWeapon).Methods(http.MethodGet)
	api.HandleFunc("/weapons/skill/{sID}", ffdb.GetWeaponsBySkill).Methods(http.MethodGet)
	api.HandleFunc("/weapons/type/{sID}", ffdb.GetWeaponsByDmgType).Methods(http.MethodGet)
	api.HandleFunc("/traits/{aID}", ffdb.GetTrait).Methods(http.MethodGet)
	api.HandleFunc("/traits/job/{sID}", ffdb.GetTraitsByJob).Methods(http.MethodGet)
	api.HandleFunc("/traits/level/{sID}", ffdb.GetTraitsByLevel).Methods(http.MethodGet)
	api.HandleFunc("/ah/items/{sID}", ffdb.GetAHByName).Methods(http.MethodGet)
	api.HandleFunc("/ah/buyer/{sID}", ffdb.GetAHByBuyer).Methods(http.MethodGet)
	api.HandleFunc("/ah/seller/{sID}", ffdb.GetAHBySeller).Methods(http.MethodGet)
	api.HandleFunc("/chars/{sID}", ffdb.GetCharByName).Methods(http.MethodGet)
	api.HandleFunc("/mobs/moblong/{sID}/{zID}/{pID}", ffdb.GetMobGroupByID).Methods(http.MethodGet)
	api.HandleFunc("/mobs/mob/{sID}", ffdb.GetMobShort).Methods(http.MethodGet)
	api.HandleFunc("/items/{sID}", ffdb.GetItemByName).Methods(http.MethodGet)
	api.HandleFunc("/items/item/{sID}", ffdb.GetItemByID).Methods(http.MethodGet)
	api.HandleFunc("/items/item/dets/{sID}", ffdb.GetItemDetsByID).Methods(http.MethodGet)
	api.HandleFunc("/skills/job/{sID}", ffdb.GetSkillRanksByJob).Methods(http.MethodGet)
	api.HandleFunc("/skills/job/{sID}/{rID}", ffdb.GetSkillRanksByLevel).Methods(http.MethodGet)
	api.HandleFunc("/events/zones/{aID}", ffdb.GetEventsByZone).Methods(http.MethodGet)
	api.HandleFunc("/events/zones/{aID}/{jID}", ffdb.GetEventsActorData).Methods(http.MethodGet)
	api.HandleFunc("/spells/job/{sID}", ffdb.GetSpellsByJob).Methods(http.MethodGet)
	api.HandleFunc("/armor/job/{sID}", ffdb.GetArmorByJob).Methods(http.MethodGet)
	api.HandleFunc("/ws/{sID}", ffdb.GetWSBySkillType).Methods(http.MethodGet)
	api.HandleFunc("/scmap/{sID}", ffdb.GetMapForSC).Methods(http.MethodGet)
	api.HandleFunc("/ws/sc/{sID}", ffdb.GetWSBySCType).Methods(http.MethodGet)
	api.HandleFunc("/merits/job/{sID}", ffdb.GetMeritsByJob).Methods(http.MethodGet)
	api.HandleFunc("/recipes/ingred/{sID}", ffdb.GetRecipesUsingItem).Methods(http.MethodGet)
	api.HandleFunc("/misc/{typ}/{jID}", ffdb.GetMiscNotes).Methods(http.MethodGet)
	api.HandleFunc("/misc/spells/bluemagic/{jID}", ffdb.GetMiscBluNotes).Methods(http.MethodGet)
	api.HandleFunc("/pets/job/{sID}", ffdb.GetPetsByJob).Methods(http.MethodGet)
	api.HandleFunc("/bcnms", ffdb.GetBCNMs).Methods(http.MethodGet)
	api.HandleFunc("/bcnms/{sID}", ffdb.GetBCNMDets).Methods(http.MethodGet)
	api.HandleFunc("/mappaths/{sID}", ffdb.GetZoneMaps).Methods(http.MethodGet)
	api.HandleFunc("/pets/dets/{sID}", ffdb.GetPetByID).Methods(http.MethodGet)
	api.HandleFunc("/tuts", ffdb.GetTuts).Methods(http.MethodGet)
	api.HandleFunc("/assaults", ffdb.GetAssaultList).Methods(http.MethodGet)

	api.HandleFunc("/crafts/recipes/{sID}/{l1ID}/{l2ID}", ffdb.GetRecipesByCraft).Methods(http.MethodGet)
	api.HandleFunc("/fishing/fish/{sID}/{l1ID}/{l2ID}/{incID}", ffdb.GetFishByLvl).Methods(http.MethodGet)
	api.HandleFunc("/fish/areas/{fID}", ffdb.GetFishAreasByID).Methods(http.MethodGet)
	api.HandleFunc("/fish/dets/{fID}", ffdb.GetFishDetsByID).Methods(http.MethodGet)
	api.HandleFunc("/fish/mob/{fID}/{zID}", ffdb.GetFishMobByID).Methods(http.MethodGet)
	api.HandleFunc("/pup/spells", ffdb.GetPupSpells).Methods(http.MethodGet)
	api.HandleFunc("/pup/frames", ffdb.GetPupFrames).Methods(http.MethodGet)
	api.HandleFunc("/pup/attach", ffdb.GetPupAttachments).Methods(http.MethodGet)
	api.HandleFunc("/missions", ffdb.GetMissions).Methods(http.MethodGet)
	api.HandleFunc("/missions/{sID}", ffdb.GetMissionList).Methods(http.MethodGet)
	api.HandleFunc("/missions/{sID}/{itemID}", ffdb.GetCatMission).Methods(http.MethodGet)

	api.HandleFunc("/pup/skillcap/{hID}/{fID}/{lID}", ffdb.GetPupSkillRanks).Methods(http.MethodGet)
	ffdb.InitJson()
	log.Fatal(http.ListenAndServe(":8080", r))//api listens on port 8080 of host machine
}
