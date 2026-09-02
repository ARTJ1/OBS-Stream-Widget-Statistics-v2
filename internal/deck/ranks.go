package deck

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/store"
)

type rankEntry struct {
	Type    string
	Level   int
	Roman   string
	ImgPath string // relative to /overlay/
}

func owRanks() []rankEntry {
	tiers := []struct {
		name string
		file string
	}{
		{"Bronze", "assets/uploads/Bronze.png"},
		{"Silver", "assets/uploads/Silver.png"},
		{"Gold", "assets/uploads/Gold.png"},
		{"Platinum", "assets/uploads/Platinum.png"},
		{"Emerald", "assets/uploads/Emerald.png"},
		{"Diamond", "assets/uploads/Diamond.png"},
		{"Master", "assets/uploads/Master.png"},
		{"Grandmaster", "assets/uploads/Grandmaster.png"},
		{"Champion", "assets/uploads/Champion.png"},
	}
	out := make([]rankEntry, 0, 45+500)
	for _, t := range tiers {
		for level := 5; level >= 1; level-- {
			out = append(out, rankEntry{Type: t.name, Level: level, ImgPath: t.file})
		}
	}
	for i := 0; i < 500; i++ {
		out = append(out, rankEntry{
			Type:    "Top",
			Level:   500 - i,
			ImgPath: "assets/uploads/Top_500.png",
		})
	}
	return out
}

var apexRoman = map[int]string{4: "IV", 3: "III", 2: "II", 1: "I"}

func apexRanks() []rankEntry {
	tiers := []string{"Rookie", "Bronze", "Silver", "Gold", "Platinum", "Diamond", "Master"}
	out := make([]rankEntry, 0, 28+750)
	for _, name := range tiers {
		file := "assets/apex/" + name + ".webp"
		for level := 4; level >= 1; level-- {
			out = append(out, rankEntry{
				Type:    name,
				Level:   level,
				Roman:   apexRoman[level],
				ImgPath: file,
			})
		}
	}
	for i := 0; i < 750; i++ {
		out = append(out, rankEntry{
			Type:    "Predator",
			Level:   750 - i,
			ImgPath: "assets/apex/Predator.webp",
		})
	}
	return out
}

var (
	cachedOW   = owRanks()
	cachedApex = apexRanks()
)

func ranksForGame(game string) []rankEntry {
	if game == store.GameApex {
		return cachedApex
	}
	return cachedOW
}

var owShortCode = map[string]string{
	"Bronze": "B", "Silver": "S", "Gold": "G", "Platinum": "P",
	"Emerald": "E", "Diamond": "D", "Master": "M",
	"Grandmaster": "GM", "Champion": "CH",
}

var apexShortCode = map[string]string{
	"Rookie": "R", "Bronze": "B", "Silver": "S", "Gold": "G",
	"Platinum": "P", "Diamond": "D", "Master": "M",
}

func shortRankCode(game string, r rankEntry) string {
	if r.Type == "Top" || r.Type == "Predator" {
		return "#" + strconv.Itoa(r.Level)
	}
	codeMap := owShortCode
	if game == store.GameApex {
		codeMap = apexShortCode
	}
	code := codeMap[r.Type]
	if code == "" {
		code = strings.ToUpper(r.Type[:1])
	}
	if r.Roman != "" {
		// Apex: G IV → G4-style digit for deck readability
		return code + strconv.Itoa(r.Level)
	}
	return code + strconv.Itoa(r.Level)
}

func resolveRank(game string, index int) (label, short, imgPath, tier string) {
	ranks := ranksForGame(game)
	if index < 0 || index >= len(ranks) {
		if game == store.GameApex {
			return "Unranked", "—", "assets/apex/Rookie.webp", "Unranked"
		}
		return "Unranked", "—", "assets/uploads/Bronze.png", "Unranked"
	}
	r := ranks[index]
	short = shortRankCode(game, r)
	switch {
	case r.Type == "Top" || r.Type == "Predator":
		label = "#" + strconv.Itoa(r.Level)
	case r.Roman != "":
		label = r.Type + " " + r.Roman
	default:
		label = fmt.Sprintf("%s %d", r.Type, r.Level)
	}
	return label, short, r.ImgPath, r.Type
}

func roleImagePath(role string) string {
	switch role {
	case store.RoleTank, store.RoleSupport, store.RoleDamage:
		return "assets/roles/" + role + ".png"
	default:
		return "assets/roles/tank.png"
	}
}
