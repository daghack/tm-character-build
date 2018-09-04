package skills

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

type Ability struct {
	Id        int
	Name      string
	Cost      int
	Stackable bool
	Tree      string
}

type AbilityPrereq struct {
	Ability
	Count int
}

type Prereq struct {
	PrereqName  string
	AbilityName string
	Count       int
}

type SkillReference struct {
	Id    int
	Count int
}

func ParseSkillsCsv(reader io.Reader) ([]*Ability, []*Prereq, error) {
	csvreader := csv.NewReader(reader)
	abilities := []*Ability{}
	prereqs := []*Prereq{}
	techTree := ""
	for {
		record, err := csvreader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		ability, abilityPrereqs, tree := parseCsvAbility(record)
		if tree != "" {
			techTree = tree
			continue
		}
		ability.Tree = techTree
		abilities = append(abilities, ability)
		prereqs = append(prereqs, abilityPrereqs...)
	}
	return abilities, prereqs, nil
}

func parseCsvAbility(fields []string) (*Ability, []*Prereq, string) {
	if fields[1] == "" {
		return nil, nil, fields[2]
	}
	toret := &Ability{Name: fields[1]}
	if strings.HasSuffix(fields[0], "*") {
		fields[0] = strings.TrimSuffix(fields[0], "*")
		toret.Stackable = true
	}
	fmt.Sscanf(fields[0], "%d", &toret.Cost)
	return toret, parseCsvPrereq(toret.Name, fields[2]), ""
}

func parseCsvPrereq(abilityName string, field string) []*Prereq {
	toret := []*Prereq{}
	if field == "—" || field == "" {
		return toret
	}
	prereqs := strings.Split(field, ", ")
	for _, prereqStr := range prereqs {
		prereq := &Prereq{AbilityName: abilityName}
		cut := strings.Split(prereqStr, " ×")
		if len(cut) < 2 {
			prereq.PrereqName = prereqStr
			prereq.Count = 1
		} else {
			prereq.PrereqName = cut[0]
			fmt.Sscanf(cut[1], "%d", &prereq.Count)
		}
		toret = append(toret, prereq)
	}
	return toret
}
