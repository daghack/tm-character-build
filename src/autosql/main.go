package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"skills"
	"strings"
)

func errH(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	dbmap, err := skills.InitDb("skills_db.bin")
	errH(err)
	defer dbmap.Db.Close()
	index, err := ioutil.ReadFile("index.html")
	errH(err)
	errH(skills.LoadCSV(dbmap, "skills.csv"))

	//shorts, err := skills.GetAbilityById(dbmap, 27)
	//errH(err)
	//fmt.Println(shorts)
	//fmt.Println(skills.ListReversePrerequisites(dbmap, shorts))
	//fmt.Println(skills.ListAbilitiesWithNoPrerequisites(dbmap))
	//fmt.Println(skills.ListAvailableAbilities(dbmap, []skills.SkillReference{skills.SkillReference{Id: shorts.Id, Count: 1}, skills.SkillReference{Id: 66, Count: 1}, skills.SkillReference{Id: 24, Count: 5}}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(index)
	})
	http.HandleFunc("/noprereqs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		abilities, _ := skills.ListAbilitiesWithNoPrerequisites(dbmap)
		bytes, _ := json.Marshal(abilities)
		w.Write(bytes)
	})
	http.HandleFunc("/prereqs", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		w.Header().Set("Access-Control-Allow-Origin", "*")
		skillValues := strings.Split(r.Form.Get("skills"), ",")
		skillList := []skills.SkillReference{}
		for _, sv := range skillValues {
			cut := strings.Split(sv, "x")
			skill := skills.SkillReference{Count: 1}
			fmt.Sscanf(cut[0], "%d", &skill.Id)
			if len(cut) > 1 {
				fmt.Sscanf(cut[1], "%d", &skill.Count)
			}
			skillList = append(skillList, skill)
		}
		fmt.Println(r.Form.Get("skills"), skillList)
		abilities, _ := skills.ListAvailableAbilities(dbmap, skillList)
		bytes, _ := json.Marshal(abilities)
		w.Write(bytes)
	})
	http.HandleFunc("/buildstr", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		fmt.Println("BUILDSTR CALLED")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		skillValues := strings.Split(r.Form.Get("skills"), ",")
		skillList := []skills.SkillReference{}
		towrite := []skills.AbilityPrereq{}
		for _, sv := range skillValues {
			cut := strings.Split(sv, "x")
			skill := skills.SkillReference{Count: 1}
			fmt.Sscanf(cut[0], "%d", &skill.Id)
			if len(cut) > 1 {
				fmt.Sscanf(cut[1], "%d", &skill.Count)
			}
			skillList = append(skillList, skill)
			ability, _ := skills.GetAbilityById(dbmap, skill.Id)
			towrite = append(towrite, skills.AbilityPrereq{Ability: *ability, Count: skill.Count})
		}
		fmt.Println(towrite)
		bytes, _ := json.Marshal(towrite)
		w.Write(bytes)
	})
	errH(http.ListenAndServe(":9090", nil))
}
