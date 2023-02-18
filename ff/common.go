package ff

import (
	"bufio"
	"database/sql"
	"net/http"
	"fmt"
	"os"
	"strings"
)

func InitJson() {
	GetBCNMJson()
	GetAttachmentJson()
	GetDesJson()
	GetMissionsJson()
	GetZoneMapPaths()
	fmt.Println("JSON files loaded")
}

func Abs(value int) int {
	if value < 0 {
		return value * -1
	} else {
		return value
	}
}

func InitHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", corStr)
}

func romanic(val int) string {
	switch val {
	case 1:
		return "I"
	case 2:
		return "II"
	case 3:
		return "III"
	case 4:
		return "IV"
	case 5:
		return "V"
	case 6:
		return "VI"
	case 7:
		return "VI"
	case 8:
		return "VIII"
	case 9:
		return "IX"
	case 10:
		return "X"
	default:
		return ""
	}
}

func GetJobString(num byte) string {
	switch num {
	case 1:
		return "Warrior"
	case 2:
		return "Monk"
	case 3:
		return "White Mage"
	case 4:
		return "Black Mage"
	case 5:
		return "Red Mage"
	case 6:
		return "Thief"
	case 7:
		return "Paladin"
	case 8:
		return "Dark Knight"
	case 9:
		return "Beastmaster"
	case 10:
		return "Bard"
	case 11:
		return "Ranger"
	case 12:
		return "Samurai"
	case 13:
		return "Ninja"
	case 14:
		return "Dragoon"
	case 15:
		return "Summoner"
	case 16:
		return "Blue Mage"
	case 17:
		return "Corsair"
	case 18:
		return "Puppetmaster"
	case 19:
		return "Dancer"
	case 20:
		return "Scholar"
	case 21:
		return "Geomancer"
	case 22:
		return "Rune Fencer"
	default:
		return ""
	}
}

func IntPow(n, m int) int {
	if m == 0 {
		return 1
	}
	result := n
	for i := 2; i <= m; i++ {
		result *= n
	}
	return result
}

func GetPowersOfTwo(x int) []int {
	var powers []int
	i := 1
	for i <= x {
		if (i & x) > 0 {
			powers = append(powers, i)
		}
		i <<= 1
	}
	return powers
}

func GetJobsString(arr []int) string {
	str := ""
	for i := 0; i < len(arr); i++ {
		switch arr[i] {
		case 1:
			str += "WAR "
		case 2:
			str += "MNK "
		case 4:
			str += "WHM "
		case 8:
			str += "BLM "
		case 16:
			str += "RDM "
		case 32:
			str += "THF "
		case 64:
			str += "PLD "
		case 128:
			str += "DRK "
		case 256:
			str += "BST "
		case 512:
			str += "BRD "
		case 1024:
			str += "RNG "
		case 2048:
			str += "SAM "
		case 4096:
			str += "NIN "
		case 8192:
			str += "DRG "
		case 16384:
			str += "SMN "
		case 32768:
			str += "BLU "
		case 65536:
			str += "COR "
		case 131072:
			str += "PUP "
		case 262144:
			str += "DNC "
		case 524288:
			str += "SCH "
		case 1048576:
			str += "GEO "
		case 2097152:
			str += "RUN "
		default:
			str += ""
		}
	}
	return str
}

func GetWSName(val int) string {
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		fmt.Println("error validating sql.Open arguments")
		panic(err.Error())
	}
	defer db.Close()
	var wsname string
	err = db.QueryRow("SELECT q.name FROM weapon_skills as q WHERE q.weaponskillid = ?", val).Scan(&wsname)
	switch {
	case err == sql.ErrNoRows:
		fmt.Printf("error selecting ws name %d", val)
	case err != nil:
		fmt.Println("error selecting ws name")
	default:
		return wsname
	}
	return ""
}

func TransformName(str string) string {
	strT := strings.Split(str, "_")
	switch strT[len(strT)-1] {
	case "ii":
		strT[len(strT)-1] = "II"
	case "iii":
		strT[len(strT)-1] = "III"
	case "iv":
		strT[len(strT)-1] = "IV"
	case "vi":
		strT[len(strT)-1] = "VII"
	case "vii":
		strT[len(strT)-1] = "VII"
	case "viii":
		strT[len(strT)-1] = "VIII"
	case "ix":
		strT[len(strT)-1] = "IX"
	case "xi":
		strT[len(strT)-1] = "XI"
	case "xii":
		strT[len(strT)-1] = "XII"
	}
	strR := strings.Join(strT, " ")
	return strings.Title(strR)
}

func GetScriptDets(file string) []string {
	var retLines []string
	// open file
	f, err := os.Open("scripts/" + file + ".lua")
	if err != nil {
		// fmt.Println("no file found")
		return retLines
	}
	// remember to close the file at the end of the program
	defer f.Close()

	// read the file line by line using scanner
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "require") || strings.HasPrefix(scanner.Text(), "func") {
			break
		} else {
			retLines = append(retLines, scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("file scan error")
	}
	f.Close()
	return retLines
}

func GetScriptItemDets(file string) []string {
	var retLines []string
	// open file
	f, err := os.Open("scripts/items/" + file + ".lua")
	if err != nil {
		//	fmt.Println("no file found")
		return retLines
	}
	// remember to close the file at the end of the program
	defer f.Close()

	// read the file line by line using scanner
	scanner := bufio.NewScanner(f)
	s := strings.Split(file, "_")[0]
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "require") || strings.HasPrefix(scanner.Text(), "func") {
			break
		} else if strings.HasPrefix(scanner.Text(), "-- ID:") || strings.HasPrefix(scanner.Text(), "-- Item:") || strings.HasPrefix(strings.ToLower(strings.ReplaceAll(scanner.Text(), "'", "")), "-- "+s) || strings.HasPrefix(scanner.Text(), "----") {
			//keep going but don't add line
		} else {
			retLines = append(retLines, strings.ReplaceAll(scanner.Text(), "-", ""))
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("file scan error")
	}
	f.Close()
	return retLines
}
