package ff

/* These are all imported from JSON files and held in memory. They mostly contain either descriptions
for something in database or they are specially made by parsing scripts.
*/
var BCNMTreasure map[string][]BCTGroup
var AttachmentMods map[string][]PupMod
var ItemDes map[string]string
var KeyItems map[string]NameVal
var AbilityDes map[string]NameVal
var SpellDes map[string]string
var Titles map[string]string
var ZoneMapPaths map[string][]string

var Missions map[string]map[string]NameVal
var AssaultMissions map[string]Assault

const lvlcap = 75

const connStr = ""

const corStr = "*" //Used for CORS policy on each http header. Default is *. Change to limit entities that can call API.

var jobsCapped = [23]string{"NONE", "WAR", "MNK", "WHM", "BLM", "RDM", "THF", "PLD", "DRK", "BST", "BRD", "RNG", "SAM", "NIN", "DRG", "SMN", "BLU", "COR", "PUP", "DNC", "SCH", "GEO", "RUN"}

//Add the corresponding key from missionMap to show as a mission category in client
var expansionsOn = [9]string{"sandoria", "bastok", "windurst", "zilart", "cop", "toau", "assault", "wotg", "acp"}

// maxHp * mob_hp_multiplier
// maxHp * np_hp_multiplier     same thing for MP
/*
	Maps to af models for each job
*/
var afMap = map[int]string{
	1:  "(64, 65)",
	2:  "(66, 67)",
	3:  "(68, 69)",
	4:  "(70, 71)",
	5:  "(72, 73)",
	6:  "(74, 75)",
	7:  "(76, 77)",
	8:  "(78, 79)",
	9:  "(80, 81)",
	10: "(82, 83)",
	11: "(84, 85)",
	12: "(86, 87)",
	13: "(88, 89)",
	14: "(90, 91)",
	15: "(92, 93)",
	16: "(165, 166)",
	17: "(167, 168)",
	18: "(169, 170)",
	19: "(210, 211, 304)",
	20: "(214, 215)",
	21: "(308, 310)",
	22: "(337, 338)",
}

/*
	Used to convert between key abbreviations and Mission category names. Keys can be used
	to limit categories shown in the client based on expansions activated for server.
*/
var missionMap = map[string]string{
	"sandoria":  "Sandoria",
	"bastok":    "Bastok",
	"windurst":  "Windurst",
	"zilart":    "Rise of the Zilart",
	"cop":       "Chains of Promathia",
	"toau":      "Treasures of Ahu Urghan",
	"assault":   "Assault",
	"wotg":      "Wings of the Goddess",
	"campaign":  "Campaign",
	"acp":       "A Crystalline Prophecy",
	"amk":       "A Moogle Kupo d'Etat",
	"asa":       "Shantotto Ascension",
	"soa":       "Seekers of Adoulin",
	"rov":       "Rhapsodies of Vanadiel",
	"coalition": "Coalition",
}

/*
	Name is title of tutorial. Byte value is how many pictures it is. If wanting to remove pages
	shown, lower the pages count based on the number of pages that need to be removed. If tutorial
	needs to be removed completely, comment out that line.
*/
var tuts = []NameNum{
	//{Name: "Abyssea", Value: 1},
	{Name: "ACP_Missions", Value: 1},
	{Name: "Action_Commands", Value: 2},
	{Name: "Adventuring_Fellows", Value: 1},
	{Name: "Allegiance_Missions", Value: 1},
	//{Name: "Alluvion_Skirmish", Value: 2},
	{Name: "AMK_Missions", Value: 1},
	{Name: "ASA_Missions", Value: 1},
	{Name: "Assault", Value: 3},
	{Name: "Auctions", Value: 2},
	{Name: "Automatons", Value: 2},
	{Name: "Avatar_Battlefields", Value: 2},
	{Name: "Bazaars", Value: 2},
	{Name: "Besieged", Value: 2},
	{Name: "Camera", Value: 3},
	{Name: "Campaign", Value: 2},
	{Name: "Changing_Jobs", Value: 1},
	{Name: "Changing_Your_Equipment", Value: 3},
	{Name: "Chatting", Value: 3},
	{Name: "Chocobo_Digging", Value: 1},
	//{Name: "Coalition_Assignments", Value: 2},
	{Name: "Combat", Value: 3},
	{Name: "Communication", Value: 3},
	{Name: "Content_Level", Value: 1},
	{Name: "COP_Missions", Value: 1},
	{Name: "Creatures_Chart", Value: 1},
	{Name: "Crystal_Synthesis", Value: 3},
	{Name: "Delivering_Items", Value: 1},
	//{Name: "Delve", Value: 3},
	//{Name: "Domain_Invasion", Value: 2},
	{Name: "Elements_Chart", Value: 1},
	{Name: "Equipment_Storage", Value: 1},
	{Name: "Extra_Jobs", Value: 3},
	{Name: "Field_Manuals", Value: 2},
	{Name: "Fishing", Value: 3},
	{Name: "Forming_a_Party", Value: 3},
	{Name: "Furnishings", Value: 1},
	{Name: "Gardening", Value: 3},
	{Name: "Gathering", Value: 2},
	//{Name: "Geas_Fete", Value: 2},
	{Name: "Healing", Value: 1},
	//{Name: "Heroes_of_Abyssea", Value: 1},
	//{Name: "High-Tier_Missions_Battlefields", Value: 2},
	{Name: "Home_Points", Value: 1},
	//{Name: "Incursion", Value: 2},
	{Name: "Inventory_Expansion", Value: 1},
	{Name: "Item_Levels", Value: 1},
	{Name: "Job_Points", Value: 2},
	{Name: "Learning_Blue_Magic", Value: 1},
	{Name: "Level_Sync", Value: 1},
	{Name: "Limit_Break_Quests", Value: 2},
	{Name: "Linkshells", Value: 3},
	{Name: "Log_Windows", Value: 2},
	{Name: "Logging_Out", Value: 1},
	{Name: "Macros", Value: 3},
	{Name: "Magic", Value: 3},
	//{Name: "Meeble_Burrows", Value: 2},
	{Name: "Merit_Points", Value: 1},
	{Name: "Mog_Gardens", Value: 2},
	{Name: "Mog_Houses", Value: 3},
	{Name: "Mog_Storage", Value: 2},
	//{Name: "Monster_Rearing", Value: 2},
	//{Name: "Monstrosity", Value: 1},
	{Name: "Movement", Value: 3},
	{Name: "Nyzul_Isle_Investigation", Value: 2},
	//{Name: "Nyzul_Isle_Uncharted", Value: 2},
	{Name: "Orb_Battlefields", Value: 2},
	{Name: "Outpost_Teleportation", Value: 1},
	{Name: "PCs_and_NPCs", Value: 1},
	{Name: "Pet_Commands", Value: 2},
	{Name: "Phantom_Roll", Value: 1},
	{Name: "Porter_Moogles", Value: 2},
	{Name: "Quests", Value: 3},
	//{Name: "Records_of_Eminence", Value: 1},
	//{Name: "Reives", Value: 3},
	{Name: "Requesting_to_Join_Party", Value: 2},
	//{Name: "Rhapsodies_of_Vanadiel", Value: 1},
	{Name: "Rise_of_the_Zilart", Value: 1},
	{Name: "Salvage", Value: 2},
	//{Name: "Salvage_II", Value: 2},
	//{Name: "Scars_of_Abyssea", Value: 1},
	{Name: "Screenshots", Value: 1},
	//{Name: "Sinister_Reign", Value: 2},
	{Name: "Skillchains", Value: 2},
	{Name: "Skills", Value: 3},
	//{Name: "Skirmish", Value: 2},
	//{Name: "SoA_Missions", Value: 1},
	{Name: "Special_Equipment_and_Furnishing_Storage", Value: 1},
	{Name: "Survival_Guides", Value: 1},
	{Name: "Synergy", Value: 2},
	{Name: "Targeting", Value: 3},
	{Name: "Text_Commands", Value: 2},
	{Name: "ToAU_Missions", Value: 1},
	{Name: "Trading", Value: 1},
	{Name: "Transportation", Value: 3},
	{Name: "Treasure_Caskets", Value: 2},
	{Name: "Trust", Value: 1},
	//{Name: "Unity_Concord", Value: 2},
	//{Name: "Vagary", Value: 2},
	//{Name: "Visions_of_Abyssea", Value: 1},
	//{Name: "Walk_of_Echoes", Value: 2},
	//{Name: "Wanted_Objectives", Value: 2},
	//{Name: "Waypoints", Value: 2},
	{Name: "WotG_Missions", Value: 1},
}
