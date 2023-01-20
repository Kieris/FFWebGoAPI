package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	ff "goAPI/ff"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	getItemGivers()
}

func GetMissingMods() []string {
	var retLines []string
	// open file
	f, err := os.Open("mods.txt")
	if err != nil {
		fmt.Println("no file found")
		f.Close()
		return retLines
	}
	// remember to close the file at the end of the program
	defer f.Close()

	// read the file line by line using scanner
	scanner := bufio.NewScanner(f)
	var str []string
	var str2 []string
	for scanner.Scan() {
		str = strings.Split(scanner.Text(), "=")
		if len(str) > 1 {
			modV := new(ff.Mods)
			str2 = strings.Split(str[1], ",")
			str2[0] = strings.ReplaceAll(str2[0], " ", "")

			modV.Id, err = strconv.Atoi(str2[0])
			modV.Val = 2
			if err != nil {
				fmt.Printf("err: %s", err)
			} else {
				ff.GetMods(modV)
				if modV.Str == "" {
					fmt.Printf("%s           : %d\n", str[0], modV.Id)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("file scan error")
	}
	f.Close()
	return retLines
}

func Convert() []string {
	var retLines []string
	// open file
	f, err := os.Open("mods.txt")
	if err != nil {
		fmt.Println("no file found")
		f.Close()
		return retLines
	}
	// remember to close the file at the end of the program
	defer f.Close()

	// read the file line by line using scanner
	scanner := bufio.NewScanner(f)
	var str []string
	var str2 []string
	for scanner.Scan() {
		str = strings.Split(scanner.Text(), "=")
		if len(str) > 1 {
			modV := new(ff.Mods)
			str2 = strings.Split(str[1], ",")
			str2[0] = strings.ReplaceAll(str2[0], " ", "")

			modV.Id, err = strconv.Atoi(str2[0])
			modV.Val = 2
			if err != nil {
				fmt.Printf("err: %s \n", err) // will print for outliers that don't match the rest
			} else {
				ff.GetMods(modV)
				if modV.Str == "" && modV.Weaponskill == "" {
					temp := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(str[0], " ", ""), "_", " "))
					modDes := strings.Title(temp)
					if len(str2) > 1 {
						if strings.Contains(str2[1], "%") || strings.Contains(str2[1], "perc") || strings.Contains(str[0], "PERC") {
							fmt.Printf("case %d:\n     mods.Str = \"%s\" + temp + mods.Val + \"%s\" %s\n", modV.Id, modDes, "%", str2[1])
						} else {
							fmt.Printf("case %d:\n     mods.Str = \"%s\" + temp + mods.Val %s\n", modV.Id, modDes, str2[1])
						}
					} else {
						fmt.Printf("case %d:\n     mods.Str = \"%s\" + temp + mods.Val\n", modV.Id, modDes)
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("file scan error")
	}
	f.Close()
	return retLines
}

type PupMod struct {
	Maneuver int
	Frame    string
	Mod      string
	Val      string
}

func getAttachmentJson() map[string][]PupMod { // "../FFXI_Events/BCNMLoots.json"
	mods := make(map[string][]PupMod)
	// open file
	f, err := os.Open("./Attachments.json")
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
	return mods
}

func getPupMods() {
	mods := make(map[string][]PupMod)
	filepath.Walk("../scripts/abilities/pets/attachments/",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.Contains(path, ".lua") {

				// open file
				f, err := os.Open(path)
				if err != nil {
					log.Fatal(err)
				}
				// remember to close the file at the end of the program
				defer f.Close()

				// read the file line by line using scanner
				scanner := bufio.NewScanner(f)
				lastLine := ""
				tempStr := strings.Split(path, "/")
				att := strings.ReplaceAll(tempStr[len(tempStr)-1], ".lua", "")
				lastManeuver := 0 // tracks how many of each maneuver is used. 0 means attachment is equipped but no maneuvers

				for scanner.Scan() {
					// do something with a line
					if strings.Contains(scanner.Text(), "pet:addMod(") {
						mod := new(PupMod)
						strInner := strings.Split(strings.Split(strings.ReplaceAll(strings.Trim(scanner.Text(), " "), "pet:addMod(tpz.mod.", ""), ")")[0], ",")
						mod.Maneuver = lastManeuver
						mod.Mod = strInner[0]
						mod.Val = strings.Trim(strInner[1], " ")
						if strings.Contains(lastLine, "local accbonus") { // val = target evasion * (.15 * maneuverCount)
							mod.Val = "targetEvasion * (.15 * " + strconv.Itoa(lastManeuver) + ") - prevAccBonus"
							mods[att] = append(mods[att], *mod)
						} else if strings.Contains(lastLine, "local skill") {
							mod.Val = "highestSkill * " + strings.ReplaceAll(mod.Val, "skill *", "")
							mods[att] = append(mods[att], *mod)
						} else if strings.Contains(lastLine, "if frame") {
							// line always ends with then
							strInner2 := strings.Split(lastLine, " ")
							str := strings.ReplaceAll(strInner2[len(strInner2)-2], "tpz.frames.", "")
							//add current mod to list
							mod.Frame = str
							mods[att] = append(mods[att], *mod)
						} else if strings.Contains(scanner.Text(), "amount") {
							//add current mod to list
							mod.Val = strconv.Itoa(lastManeuver)
							mods[att] = append(mods[att], *mod)
						} else {
							if strings.Contains(lastLine, "maneuvers") || strings.Contains(lastLine, "pet:addMod(") || strings.Contains(lastLine, "onEquip(pet)") || strings.Contains(lastLine, "updateModPerformance(pet, tpz.mod.STORETP,") {
								//ignore
							} else {
								fmt.Println(lastLine) //want to see it
							}
							//add current mod to list
							mods[att] = append(mods[att], *mod)
						}
					} else if strings.Contains(scanner.Text(), "updateModPerformance(pet, tpz.mod.") {
						strInner := strings.Split(strings.Split(strings.ReplaceAll(strings.Trim(scanner.Text(), " "), "updateModPerformance(pet, tpz.mod.", ""), ")")[0], ",")
						//add current mod to list
						mod := new(PupMod)
						mod.Maneuver = lastManeuver
						mod.Mod = strInner[0]
						mod.Val = strings.Trim(strInner[len(strInner)-1], " ")
						if mod.Val != "0" { // means attachment was unequiped and don't care about these currently
							mods[att] = append(mods[att], *mod)
						}
					}

					lastLine = scanner.Text()
					if strings.Contains(lastLine, "maneuvers") {
						if strings.Contains(lastLine, "1") {
							lastManeuver = 1
						} else if strings.Contains(lastLine, "2") {
							lastManeuver = 2
						} else if strings.Contains(lastLine, "3") {
							lastManeuver = 3
						} else if strings.Contains(lastLine, "0") {
							lastManeuver = 0
						}
					} else if strings.Contains(lastLine, "onEquip(pet)") {
						lastManeuver = 0
					}
				}
				if err := scanner.Err(); err != nil {
					log.Fatal(err)
				}
				f.Close()
			}
			return nil
		})
	//Can print out map values
	/*
		for k, v := range mods {
			fmt.Println("k:", k, "v:", v)
		}
	*/
	jsonData, _ := json.Marshal(mods)

	// write to JSON file
	jsonFile, err := os.Create("./Attachments.json")

	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	jsonFile.Write(jsonData)
	jsonFile.Close()
	fmt.Println("JSON data written to ", jsonFile.Name())
	if err != nil {
		log.Println(err)
	}
}

type BCTGroup struct {
	Group      int
	BCTreasure []*BCTreasure
}

type BCTreasure struct {
	ItemId int
	Name   string
	Rate   int
}

type ScriptItemDets struct {
	Items     []ScriptItem
	Zone      string
	NPC       string
	QuestReqs []*string
	TradeReqs []*ScriptItem
}

type ScriptItem struct {
	ItemId int
	Count  int
}

func getBCNMTreasure() {
	zoneList := map[string]struct{}{} // works like a HashSet
	bcnms := ff.GetBCList()
	count := 0
	for i := range bcnms {
		if bcnms[i].LootID != 0 {
			zoneList[bcnms[i].ZName] = struct{}{}
		}
	}
	treasure := make(map[string][]BCTGroup)
	for zone := range zoneList {
		// open file
		f, err := os.Open("../scripts/zones/" + zone + "/npcs/Armoury_Crate.lua")
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		lootStarted := false
		group := 1
		checkEnd := false
		loot := "0"
		var bctg *BCTGroup
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "] =") {
				loot = strings.Trim(strings.Split(scanner.Text(), "]")[0], "[ ")
				//fmt.Printf("new loot:    %s\n", loot)  //used to see if all loot IDs got added to map
				lootStarted = true
				group = 1
				bctg = new(BCTGroup)
				bctg.Group = group
			} else if strings.Contains(scanner.Text(), "=") && lootStarted && strings.Contains(scanner.Text(), "}") {
				strB := strings.Split(strings.Trim(strings.Split(strings.ReplaceAll(scanner.Text(), " ", ""), "}")[0], "{"), ",")
				if strings.HasPrefix(strB[0], "--") { // means item was commented out
					fmt.Println(scanner.Text()) //ignore but print
				} else {
					count++ // sanity check
					temp := new(BCTreasure)
					// populate name
					if strings.Contains(scanner.Text(), "amount") {
						amnt := strings.Split(strB[2], "=")[1]
						temp.Name = fmt.Sprintf("%s Gil", amnt)
					} else if strings.Contains(scanner.Text(), "--") {
						strTempArr := strings.Split(scanner.Text(), "--")
						temp.Name = strings.TrimPrefix(strTempArr[len(strTempArr)-1], " ")
					}
					item := strings.Split(strB[0], "=")[1]
					rate := strings.Split(strB[1], "=")[1]
					temp.ItemId, err = strconv.Atoi(item)
					if err != nil {
						fmt.Println(err.Error())
					}
					temp.Rate, err = strconv.Atoi(rate)
					if err != nil {
						fmt.Println(err.Error())
					}
					bctg.BCTreasure = append(bctg.BCTreasure, temp)
				}
				checkEnd = false // reset }, "count"
				//fmt.Printf("loot %s group %d item %s rate %s\n", loot, group, item, rate)
			} else if strings.Contains(scanner.Text(), "}") && lootStarted {
				if checkEnd { //checks for }, repeated
					checkEnd = false
					lootStarted = false
				} else {
					//fmt.Printf("add loot: %s\n", loot)//used to see if all loot IDs got added to map
					treasure[loot] = append(treasure[loot], *bctg)
					group++
					bctg = new(BCTGroup)
					bctg.Group = group
					checkEnd = true
				}
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
		f.Close()
	}
	countT := 0
	for k := range treasure {
		// fmt.Println("k:", k, "v:", v)
		for i := range treasure[k] {
			// fmt.Printf("loot: %s grp: %d\n", k, treasure[k][i].Group)
			for range treasure[k][i].BCTreasure {
				//fmt.Printf("loot: %s grp: %d item: %d\n", k, treasure[k][i].Group, treasure[k][i].BCTreasure[j].ItemId)
				countT++ //sanity check to match with earlier count
			}
		}
	}
	// These not matching means all loot did not get added to treasure somehow
	fmt.Println(count)
	fmt.Println(countT)

	jsonData, _ := json.Marshal(treasure)

	// write to JSON file
	jsonFile, err := os.Create("../FFXI_Events/BCNMLoots.json")

	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	jsonFile.Write(jsonData)
	jsonFile.Close()
	fmt.Println("JSON data written to ", jsonFile.Name())
	if err != nil {
		log.Println(err)
	}
}

func getItemGivers() {
	//retMap := make(map[string][]ScriptItemDets)
	var items []*ScriptItemDets
	count := 0
	err := filepath.Walk("../scripts/zones/",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.Contains(path, ".lua") {

				// open file
				f, err := os.Open(path)
				if err != nil {
					log.Fatal(err)
				}
				// remember to close the file at the end of the program
				defer f.Close()

				// read the file line by line using scanner
				scanner := bufio.NewScanner(f)
				//lastText := ""
				for scanner.Scan() {
					// do something with a line
					if strings.Contains(scanner.Text(), "giveItem(player,") { //"giveItem(player,"
						count++
						temp := new(ScriptItemDets)

						str1 := strings.Split(path, "../scripts/zones/")
						str1 = strings.Split(str1[1], "/")
						str2 := strings.Split(str1[len(str1)-1], ".lua")[0]

						str3 := strings.Split(scanner.Text(), "giveItem(player,")
						str3 = strings.Split(str3[len(str3)-1], ")")

						// fmt.Printf("zone:%s npc: %s item: %s\n", str1[0], str2, str3[0])
						temp.Items = getScriptItems(str3[0])
						temp.Zone = str1[0]
						temp.NPC = str2
						if len(temp.Items) == 0 {
							fmt.Printf("line: %s\n", str3[0])
						}
						//fmt.Printf("line: %v\n", temp)
						items = append(items, temp)
						/*fmt.Printf("last: %s\n", lastText)
						if strings.Contains(scanner.Text(), "player:getCharVar(") {
							fmt.Printf("line: %s\n", scanner.Text())
						} else if strings.Contains(lastText, "player:getCharVar(") {
							fmt.Printf("last: %s\n", lastText)
						}

						if strings.Contains(scanner.Text(), "npcUtil.tradeHas(") {
							fmt.Printf("line: %s\n", scanner.Text())
						} else if strings.Contains(lastText, "npcUtil.tradeHas(") {
							fmt.Printf("last: %s\n", lastText)
						}*/

					} else if strings.Contains(scanner.Text(), "giveItem") {
						fmt.Println(scanner.Text())
						count++
					}
					//lastText = scanner.Text()
				}
				if err := scanner.Err(); err != nil {
					log.Fatal(err)
				}

				// fmt.Println(path)
			}
			return nil
		})
	fmt.Println(len(items))
	fmt.Println(count)

	if err != nil {
		log.Println(err)
	}
}

func getScriptItems(str string) []ScriptItem {
	var items []ScriptItem
	str = strings.ReplaceAll(str, " ", "")
	id, err := strconv.Atoi(str)
	if err != nil {
		//look for multiple items, brackets, or count
		temp := new(ScriptItem)
		if strings.Contains(str, "}}") {
			strArr := strings.Split(str, "}}")
			if strings.Contains(strArr[0], "{") {
				//there is a second item
				strArr = strings.Split(strArr[0], "{")
				strArr2 := strings.Split(strArr[1], ",")
				if len(strArr2) > 1 && strArr2[1] != "" {
					id, err = strconv.Atoi(strings.ReplaceAll(strArr2[0], "}", ""))
					if err == nil {
						temp.ItemId = id
					} else {
						fmt.Println(err)
						fmt.Println(str)
					}
					if strArr2[1] != "" {
						cnt, err := strconv.Atoi(strings.ReplaceAll(strArr2[1], "}", ""))
						if err == nil {
							temp.Count = cnt
							items = append(items, *temp)
						} else {
							fmt.Println(err)
							fmt.Println(str)
						}
					}
				} else if len(strArr2) > 0 && strArr2[0] != "" {
					id, err = strconv.Atoi(strings.ReplaceAll(strArr2[0], "}", ""))
					if err == nil {
						temp.ItemId = id
						temp.Count = 1
						items = append(items, *temp)
					} else {
						fmt.Println(err)
						fmt.Println(strings.ReplaceAll(strArr2[0], "}", ""))
					}
				}
				temp = new(ScriptItem)
				strArr2 = strings.Split(strArr[2], ",")
				if len(strArr2) > 1 && strArr2[1] != "" {
					id, err = strconv.Atoi(strings.ReplaceAll(strArr2[0], "}", ""))
					if err == nil {
						temp.ItemId = id
					} else {
						fmt.Println(err)
						fmt.Println(str)
					}
					if strArr2[1] != "" {
						cnt, err := strconv.Atoi(strings.ReplaceAll(strArr2[1], "}", ""))
						if err == nil {
							temp.Count = cnt
							items = append(items, *temp)
						} else {
							fmt.Println(err)
							fmt.Println(strings.ReplaceAll(strArr2[1], "}", ""))
						}
					}
				} else if len(strArr2) > 0 && strArr2[0] != "" {
					id, err = strconv.Atoi(strings.ReplaceAll(strArr2[0], "}", ""))
					if err == nil {
						temp.ItemId = id
						temp.Count = 1
						items = append(items, *temp)
					} else {
						fmt.Println(err)
						fmt.Println(strings.ReplaceAll(strArr2[0], "}", ""))
					}
				}
			}
			//fmt.Println(strArr[1])
		} else if strings.Contains(str, ",") {
			str = strings.Trim(str, "{}")
			strArr := strings.Split(str, ",")
			id, err = strconv.Atoi(strArr[0])
			if err == nil {
				temp.ItemId = id
			} else {
				fmt.Println(err)
				fmt.Println(str)
			}
			cnt, err := strconv.Atoi(strArr[1])
			if err == nil {
				temp.Count = cnt
				items = append(items, *temp)
			} else {
				fmt.Println(err)
				fmt.Println(str)
			}
		}
	} else {
		temp := new(ScriptItem)
		temp.ItemId = id
		temp.Count = 1
		items = append(items, *temp)
	}

	return items
}
