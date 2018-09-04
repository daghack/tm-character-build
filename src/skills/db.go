package skills

import (
	"database/sql"
	"fmt"
	"github.com/go-gorp/gorp"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

const getPrereqs string = `
SELECT
	Id,
	Name,
	Cost,
	Stackable,
	Tree,
	Count
FROM prerequisites p
INNER JOIN abilities a ON p.PrereqName = a.Name
WHERE
	p.AbilityName = ?
`

const getReversePrereqs string = `
SELECT DISTINCT
	Id,
	Name,
	Cost,
	Stackable,
	Tree
FROM prerequisites p
INNER JOIN abilities a ON p.AbilityName = a.Name
WHERE
	p.PrereqName = ?
GROUP BY a.Id
`

const getNoPrereqs string = `
SELECT
	Id,
	Name,
	Cost,
	Stackable,
	Tree
FROM abilities a
LEFT JOIN prerequisites p ON p.AbilityName = a.Name
WHERE
	p.AbilityName IS NULL
`

const getAbilityById string = `SELECT * FROM abilities WHERE Id = ?`

func LoadCSV(dbmap *gorp.DbMap, filename string) error {
	err := dbmap.TruncateTables()
	if err != nil {
		return err
	}
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	abilities, prereqs, err := ParseSkillsCsv(file)
	if err != nil {
		return err
	}
	for _, ability := range abilities {
		err = dbmap.Insert(ability)
		if err != nil {
			return err
		}
	}
	for _, prereq := range prereqs {
		err = dbmap.Insert(prereq)
		if err != nil {
			return err
		}
	}
	return nil
}

func InitDb(dbfile string) (*gorp.DbMap, error) {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return nil, err
	}
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	dbmap.AddTableWithName(Ability{}, "abilities").SetKeys(true, "Id").AddIndex("SkillName", "Hash", []string{"Name"}).SetUnique(true)
	prereqMap := dbmap.AddTableWithName(Prereq{}, "prerequisites")
	prereqMap.AddIndex("PrereqName", "Hash", []string{"PrereqName"}).SetUnique(false)
	prereqMap.AddIndex("AbilityName", "Hash", []string{"AbilityName"}).SetUnique(false)
	err = dbmap.CreateTablesIfNotExists()
	if err != nil {
		return nil, err
	}
	return dbmap, nil
}

func ListPrerequisites(dbmap *gorp.DbMap, ability *Ability) ([]AbilityPrereq, error) {
	toret := []AbilityPrereq{}
	_, err := dbmap.Select(&toret, getPrereqs, ability.Name)
	if err != nil {
		return nil, err
	}
	return toret, nil
}

func ListReversePrerequisites(dbmap *gorp.DbMap, ability *Ability) ([]Ability, error) {
	toret := []Ability{}
	_, err := dbmap.Select(&toret, getReversePrereqs, ability.Name)
	if err != nil {
		return nil, err
	}
	return toret, nil
}

func GetAbilityById(dbmap *gorp.DbMap, abilityId int) (*Ability, error) {
	toret := Ability{}
	err := dbmap.SelectOne(&toret, getAbilityById, abilityId)
	if err != nil {
		return nil, err
	}
	return &toret, nil
}

func ListAbilitiesWithNoPrerequisites(dbmap *gorp.DbMap) ([]Ability, error) {
	toret := []Ability{}
	_, err := dbmap.Select(&toret, getNoPrereqs)
	if err != nil {
		return nil, err
	}
	return toret, nil
}

func ListAvailableAbilities(dbmap *gorp.DbMap, skills []SkillReference) ([]Ability, error) {
	toret := []Ability{}
	noPrereqs, err := ListAbilitiesWithNoPrerequisites(dbmap)
	if err != nil {
		return nil, err
	}
	haveskill := map[int]int{}
	toretskill := map[int]bool{}
	potentiallyAvailable := []Ability{}
	for _, v := range skills {
		haveskill[v.Id] = v.Count
		ability, err := GetAbilityById(dbmap, v.Id)
		if err != nil {
			return nil, err
		}
		if ability.Stackable {
			toretskill[ability.Id] = true
			toret = append(toret, *ability)
		}
		localpotentials, err := ListReversePrerequisites(dbmap, ability)
		fmt.Println(localpotentials)
		if err != nil {
			return nil, err
		}
		potentiallyAvailable = append(potentiallyAvailable, localpotentials...)
	}
	for _, np := range noPrereqs {
		if (haveskill[np.Id] == 0 || np.Stackable) && !toretskill[np.Id] {
			toretskill[np.Id] = true
			toret = append(toret, np)
		}
	}
	for _, p := range potentiallyAvailable {
		if haveskill[p.Id] > 0 || toretskill[p.Id] {
			continue
		}
		prereqs, err := ListPrerequisites(dbmap, &p)
		if err != nil {
			return nil, err
		}
		requirementMet := true
		for _, prereq := range prereqs {
			if haveskill[prereq.Id] < prereq.Count {
				requirementMet = false
				break
			}
		}
		if requirementMet {
			toretskill[p.Id] = true
			toret = append(toret, p)
		}
	}
	return toret, nil
}
