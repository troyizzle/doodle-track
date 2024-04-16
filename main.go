package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const (
	JSON_FILE = "doodles.json"
)

type DoodleData struct {
	Dna    string
	Town   string
	Zone   string
	Traits []interface{}
	Cost   int
	Score  int
}

func main() {
	run()
}

func run() {
	doodlesFromFile := readDoodlesFromFile()

	url := "https://toonhq.org/api/doodles/1"

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching results: ", err)
	}
	defer resp.Body.Close()

	var data map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println("Error decoding json: ", err)
	}

	for zone, towns := range data {
		for town, doodles := range towns.(map[string]interface{}) {
			for _, doodle := range doodles.([]interface{}) {
				doodleData := doodle.(map[string]interface{})
				cost := doodleData["cost"].(float64)

				doodle := DoodleData{
					Dna:    doodleData["dna"].(string),
					Town:   town,
					Zone:   zone,
					Traits: doodleData["traits"].([]interface{}),
					Cost:   int(cost),
					Score:  calculateScore(doodleData["traits"].([]interface{})),
				}

				if _, exists := doodlesFromFile[doodle.Dna]; exists {
				} else {
					doodlesFromFile[doodle.Dna] = doodle
					if doodle.Score >= 16 {
						fmt.Println("New doodle found: ", doodle)
						fmt.Println("Score: ", doodle.Score)
					}
				}
			}
		}
	}

	writeResultsToFile(doodlesFromFile)
}

func readDoodlesFromFile() map[string]interface{} {
	file, err := os.Open(JSON_FILE)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{})
		}

		fmt.Println("Error opening file: ", err)
	}

	defer file.Close()

	var doodles map[string]interface{}

	if err := json.NewDecoder(file).Decode(&doodles); err != nil {
		fmt.Println("Error decoding json: ", err)
	}

	return doodles
}

func writeResultsToFile(data map[string]interface{}) {
	file, err := os.Create(JSON_FILE)
	if err != nil {
		fmt.Println("Error creating file: ", err)
	}

	defer file.Close()

	encoder := json.NewEncoder(file)

	if err := encoder.Encode(data); err != nil {
		fmt.Println("Error encoding json: ", err)
	}
}

func calculateScore(traits []interface{}) int {
	traitScoring := map[string]int{
		"fc": -3,
		"hc": -1,
		"gc": 1,
		"vc": 3,
		"bc": 6,
		"_c": 9,
	}

	traitMap := map[string][]string{
		"fc": {"Always Forgets", "Always Grumpy", "Always Lonely", "Often Bored", "Often Confused", "Often Grumpy", "Often Hungry", "Often Lonely", "Often Sad", "Often Tired", "Rarely Affectionate"},
		"hc": {"Sometimes Bored", "Sometimes Confused", "Sometimes Forgets", "Sometimes Grumpy", "Sometimes Hungry", "Sometimes Lonely", "Sometimes Sad", "Sometimes Tired"},
		"gc": {"Often Affectionate", "Pretty Calm", "Pretty Excitable", "Sometimes Affectionate"},
		"vc": {"Always Affectionate", "Rarely Bored", "Rarely Confused", "Rarely Forgets", "Rarely Grumpy", "Rarely Lonely", "Rarely Sad", "Very Excitable"},
		"bc": {"Rarely Tired"},
	}

	weightedTraits := make(map[string]int)

	for key, traits := range traitMap {
		for _, trait := range traits {
			weightedTraits[trait] = traitScoring[key]
		}
	}

	score := 0

	for _, trait := range traits {
		score += weightedTraits[trait.(string)]
	}

	score = min(score, 14)

	if len(traits) == 0 {
		return score
	}

	if traits[0] == "Rarely Tired" {
		score += traitScoring["_c"] - traitScoring["bc"]
	}

	return score
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
