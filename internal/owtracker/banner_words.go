package owtracker

import (
	"strings"
	"unicode"
)

// Overwatch 2 end-of-match banner words per client language (accents stripped).
var bannerWords = map[Outcome][]string{
	OutcomeWin:  {"ПОБЕДА", "VICTORY", "SIEG", "VICTOIRE", "VICTORIA", "VITORIA", "VITTORIA", "ZWYCIESTWO"},
	OutcomeLoss: {"ПОРАЖЕНИЕ", "DEFEAT", "NIEDERLAGE", "DEFAITE", "DERROTA", "SCONFITTA", "PORAZKA"},
	OutcomeDraw: {"НИЧЬЯ", "DRAW", "UNENTSCHIEDEN", "EGALITE", "EMPATE", "PAREGGIO", "REMIS"},
}

// OutcomeDraw is recognized so a draw banner never counts as anything.
const OutcomeDraw Outcome = "draw"

// normalizeBannerText keeps letters only, uppercased, accents removed.
func normalizeBannerText(s string) string {
	var b strings.Builder
	for _, r := range s {
		if base, ok := accentFold[unicode.ToUpper(r)]; ok {
			r = base
		}
		if unicode.IsLetter(r) {
			b.WriteRune(unicode.ToUpper(r))
		}
	}
	return b.String()
}

// maxWordDistance: OCR on the stylized banner font misreads a letter or two.
func maxWordDistance(word []rune) int {
	switch n := len(word); {
	case n <= 4:
		return 0
	case n <= 5:
		return 1
	case n <= 8:
		return 2 // ПОБЕДА is often read as ПОВЕДЯ
	default:
		return 3
	}
}

// matchBannerWord maps OCR text to an outcome, or "" when it is not a banner word.
func matchBannerWord(text string) (Outcome, string) {
	got := []rune(normalizeBannerText(text))
	if len(got) < 4 || len(got) > 16 {
		return "", ""
	}
	best, bestWord, bestDist := Outcome(""), "", 99
	for outcome, words := range bannerWords {
		for _, w := range words {
			wr := []rune(w)
			if abs(len(got)-len(wr)) > 2 {
				continue
			}
			d := runeLevenshtein(got, wr)
			if d <= maxWordDistance(wr) && d < bestDist {
				best, bestWord, bestDist = outcome, w, d
			}
		}
	}
	return best, bestWord
}

func runeLevenshtein(a, b []rune) int {
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min3int(cur[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// accentFold covers the accented letters used by the banner words above.
var accentFold = map[rune]rune{
	'À': 'A', 'Á': 'A', 'Â': 'A', 'Ã': 'A', 'Ä': 'A', 'Ą': 'A',
	'Ç': 'C', 'Ć': 'C',
	'È': 'E', 'É': 'E', 'Ê': 'E', 'Ë': 'E', 'Ę': 'E',
	'Ì': 'I', 'Í': 'I', 'Î': 'I', 'Ï': 'I',
	'Ł': 'L', 'Ń': 'N', 'Ñ': 'N',
	'Ò': 'O', 'Ó': 'O', 'Ô': 'O', 'Õ': 'O', 'Ö': 'O',
	'Ś': 'S', 'Ù': 'U', 'Ú': 'U', 'Û': 'U', 'Ü': 'U',
	'Ź': 'Z', 'Ż': 'Z', 'Ё': 'Е',
}
