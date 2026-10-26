package owtracker

import (
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/nfnt/resize"
)

// Auto rank (alpha): after a competitive match Overwatch shows a progress screen
// with the queued role on top ("DAMAGE", "УРОН") and the rank under the emblem
// ("EMERALD 3", "ИЗУМРУД 2"). Both are plain text, so Windows OCR reads them from
// two small crops in the middle of the 16:9 game area. The rank is absolute, so a
// missed match never puts the widget out of sync.

// RankRead is one look at the progress screen.
type RankRead struct {
	Role     string `json:"role"`     // tank | damage | support | open
	Tier     string `json:"tier"`     // Bronze … Champion
	Division int    `json:"division"` // 1 … 5
	Rank     int    `json:"rank"`     // widget ladder index (Bronze 5 = 0)
	RoleText string `json:"roleText,omitempty"`
	RankText string `json:"rankText,omitempty"`
	Lang     string `json:"lang,omitempty"`
}

// Valid reports a full read: a known role and a rank.
func (r RankRead) Valid() bool { return r.Role != "" && r.Tier != "" && r.Division >= 1 }

const (
	RankRoleOpen = "open" // open queue / 6v6: one rank for all roles
)

// Crops as fractions of the 16:9 game area (same on every resolution).
var (
	rankRoleCrop  = [4]float64{0.36, 0.165, 0.28, 0.07}
	rankLabelCrop = [4]float64{0.35, 0.595, 0.30, 0.075}
)

// rankOCRWidth: Windows OCR reads the stylized font best at this crop width
// (measured on 720p/1080p recordings; wider and narrower both lose letters).
const rankOCRWidth = 864

// Tier names per client language, in ladder order (accents stripped).
var rankTiers = []struct {
	Name  string
	Words []string
}{
	{"Bronze", []string{"BRONZE", "БРОНЗА", "BRONCE", "BRONZO", "BRAZ"}},
	{"Silver", []string{"SILVER", "СЕРЕБРО", "SILBER", "ARGENT", "PLATA", "PRATA", "ARGENTO", "SREBRO"}},
	{"Gold", []string{"GOLD", "ЗОЛОТО", "OR", "ORO", "OURO", "ZLOTO"}},
	{"Platinum", []string{"PLATINUM", "ПЛАТИНА", "PLATIN", "PLATINE", "PLATINO", "PLATINA", "PLATYNA"}},
	{"Emerald", []string{"EMERALD", "ИЗУМРУД", "SMARAGD", "EMERAUDE", "ESMERALDA", "SMERALDO", "SZMARAGD"}},
	{"Diamond", []string{"DIAMOND", "АЛМАЗ", "DIAMANT", "DIAMANTE", "DIAMENT"}},
	{"Master", []string{"MASTER", "МАСТЕР", "MEISTER", "MAITRE", "MAESTRO", "MESTRE", "MISTRZ"}},
	{"Grandmaster", []string{"GRANDMASTER", "ГРАНДМАСТЕР", "GROSSMEISTER", "GROßMEISTER", "GRANDMAITRE", "GRANMAESTRO", "GRAOMESTRE", "ARCYMISTRZ"}},
	{"Champion", []string{"CHAMPION", "ЧЕМПИОН", "CAMPEON", "CAMPEAO", "CAMPIONE", "CZEMPION"}},
}

// Role header words per client language (accents stripped).
var rankRoleWords = map[string][]string{
	"tank":       {"TANK", "ТАНК", "TANQUE", "CZOLG"},
	"damage":     {"DAMAGE", "УРОН", "SCHADEN", "DEGATS", "DANO", "DANNI", "OBRAZENIA"},
	"support":    {"SUPPORT", "ПОДДЕРЖКА", "UNTERSTUTZUNG", "SOUTIEN", "APOYO", "SUPORTE", "SUPPORTO", "WSPARCIE"},
	RankRoleOpen: {"OPENQUEUE", "ОТКРЫТАЯОЧЕРЕДЬ", "ОТКРЫТАЯ", "OFFENEWARTESCHLANGE", "FILEOUVERTE", "COLAABIERTA", "FILAABERTA", "CODAAPERTA", "OTWARTAKOLEJKA"},
}

// RankIndex maps tier + division onto the widget's Overwatch ladder.
func RankIndex(tier string, division int) int {
	for i, t := range rankTiers {
		if t.Name == tier && division >= 1 && division <= 5 {
			return i*5 + (5 - division)
		}
	}
	return -1
}

// matchRankRole finds the role word in the OCR text of the role header.
func matchRankRole(text string) string {
	role, _ := matchRankRoleBest(text)
	return role
}

// matchRankRoleBest is matchRankRole with a score (lower is better): an exact
// word in its own alphabet beats a misread letter, which beats look-alikes.
func matchRankRoleBest(text string) (string, int) {
	if role, d := matchRankRoleIn(text, false); role != "" {
		return role, d
	}
	if role, d := matchRankRoleIn(text, true); role != "" {
		return role, 10 + d
	}
	return "", 99
}

// matchRankRoleIn matches as read, or with look-alike letters folded ("ypoH" is
// УРОН read by the English OCR: a match, but not the client's language).
func matchRankRoleIn(text string, fold bool) (string, int) {
	conv := func(rs []rune) []rune { return rs }
	if fold {
		conv = foldRankGlyphs
	}
	toks := strings.Fields(text)
	// Single words, then neighbour pairs ("OPEN QUEUE").
	cands := make([]string, 0, len(toks)*2)
	for i, tk := range toks {
		cands = append(cands, normalizeBannerText(tk))
		if i+1 < len(toks) {
			cands = append(cands, normalizeBannerText(tk+toks[i+1]))
		}
	}
	best, bestDist := "", 99
	for _, c := range cands {
		got := conv([]rune(c))
		if len(got) < 2 {
			continue
		}
		for role, words := range rankRoleWords {
			for _, w := range words {
				wr := conv([]rune(w))
				if abs(len(got)-len(wr)) > 2 {
					continue
				}
				d := runeLevenshtein(got, wr)
				if d <= rankWordDistance(got, wr) && d < bestDist {
					best, bestDist = role, d
				}
			}
		}
	}
	return best, bestDist
}

// rankWordDistance is maxWordDistance, except that a 4-letter word ("GOLD",
// "TANK") may have one misread letter when the length matches ("G0LD", "TAMK").
// The first letter must match: "nano" (gameplay text) is not "DANO".
func rankWordDistance(got, word []rune) int {
	if len(word) == 4 && len(got) == 4 && got[0] == word[0] {
		return 1
	}
	return maxWordDistance(word)
}

// matchRankLabel parses "EMERALD 3" / "ИЗУМРУД 2" / "DIRMOND5" / "GOLD I" /
// "GOLDI" into tier + division.
func matchRankLabel(text string) (string, int) {
	div, body := rankDivision(text)
	// Digits inside the word are letters ("G0LD 3", "АЛМА3 2"), and Л may come
	// back as two Latin strokes ("AJlMA3 2").
	letters := foldRankGlyphs([]rune(normalizeBannerText(rankBodyFix.Replace(body))))
	if len(letters) < 2 || len(letters) > 16 {
		return "", 0
	}
	if div > 0 {
		tier, _ := matchRankTier(letters)
		if tier == "" {
			return "", 0
		}
		return tier, div
	}
	// No digit: the division may be glued on as a look-alike letter ("GOLDI",
	// "ИЗУМРУДЗ"). Only when the word without it is exact and the whole text is
	// itself a tier ("АЛМАЗ" with no division).
	d, ok := rankGluedGlyphs[letters[len(letters)-1]]
	if !ok || len(letters) < 4 {
		return "", 0
	}
	tier, dist := matchRankTier(letters[:len(letters)-1])
	if tier == "" || dist != 0 {
		return "", 0
	}
	if whole, wd := matchRankTier(letters); whole != "" && wd == 0 {
		return "", 0
	}
	return tier, d
}

// matchRankTier returns the closest tier to the (folded) letters and the distance.
func matchRankTier(letters []rune) (string, int) {
	best, bestDist := "", 99
	for _, t := range rankTiers {
		for _, w := range t.Words {
			wr := foldRankGlyphs([]rune(w))
			if abs(len(letters)-len(wr)) > 2 {
				continue
			}
			d := runeLevenshtein(letters, wr)
			if d <= rankWordDistance(letters, wr) && d < bestDist {
				best, bestDist = t.Name, d
			}
		}
	}
	return best, bestDist
}

var rankBodyFix = strings.NewReplacer("0", "O", "3", "З", "Jl", "Л", "JI", "Л", "J|", "Л")

// rankGlyphFold merges letters Windows OCR swaps in the italic rank font
// (ПЛАТИНА -> ППЯТИНЯ, EMERALD -> FMFRALD), and Cyrillic letters with their Latin
// look-alikes, since either OCR language may read either alphabet ("ТАНК" as
// "TAHK"); applied to both sides of a match.
var rankGlyphFold = map[rune]rune{
	'Л': 'П', 'Я': 'A', 'F': 'E',
	'А': 'A', 'В': 'B', 'Е': 'E', 'К': 'K', 'М': 'M', 'Н': 'H', 'О': 'O',
	'Р': 'P', 'С': 'C', 'Т': 'T', 'Х': 'X', 'У': 'Y',
}

func foldRankGlyphs(rs []rune) []rune {
	out := make([]rune, len(rs))
	for i, r := range rs {
		if f, ok := rankGlyphFold[r]; ok {
			r = f
		}
		out[i] = r
	}
	return out
}

// rankDivisionGlyphs: letters Windows OCR returns for a division digit.
var rankDivisionGlyphs = map[string]int{
	"I": 1, "i": 1, "l": 1, "|": 1, "!": 1,
	"З": 3, "з": 3,
	"a": 4, "A": 4, "а": 4, "А": 4, "Ч": 4, "ч": 4, // the italic 4
	"S": 5, "s": 5,
}

// rankGluedGlyphs: the same look-alikes glued to the word (after normalizing).
var rankGluedGlyphs = map[rune]int{'З': 3, 'I': 1}

// rankDivision is the last 1–5 digit of the label (OCR sometimes glues it on)
// and the rest of the text, the tier word.
func rankDivision(text string) (int, string) {
	// A lone last token that only looks like a digit ("ИЗУМРУД З", "GOLD I").
	if toks := strings.Fields(text); len(toks) > 1 {
		if d, ok := rankDivisionGlyphs[toks[len(toks)-1]]; ok {
			return d, strings.Join(toks[:len(toks)-1], " ")
		}
	}
	rs := []rune(strings.TrimSpace(text))
	for i := len(rs) - 1; i >= 0; i-- {
		r := rs[i]
		if r >= '1' && r <= '5' {
			return int(r - '0'), string(rs[:i])
		}
		if r != '0' && unicode.IsDigit(r) {
			return 0, text // 6–9: not a division
		}
	}
	return 0, text
}

// rankCrop cuts a crop (fractions of the 16:9 area) and scales it for OCR.
func rankCrop(img image.Image, c [4]float64) image.Image {
	b := img.Bounds()
	x, y, w, h := bannerBounds(0, 0, b.Dx(), b.Dy())
	r := image.Rect(
		b.Min.X+x+int(c[0]*float64(w)), b.Min.Y+y+int(c[1]*float64(h)),
		b.Min.X+x+int((c[0]+c[2])*float64(w)), b.Min.Y+y+int((c[1]+c[3])*float64(h)),
	)
	sub, ok := img.(interface {
		SubImage(image.Rectangle) image.Image
	})
	if !ok || r.Dx() < 16 || r.Dy() < 4 {
		return nil
	}
	crop := sub.SubImage(r)
	nh := uint(float64(r.Dy()) * rankOCRWidth / float64(r.Dx()))
	return resize.Resize(rankOCRWidth, nh, crop, resize.Lanczos3)
}

// rankLabelShear uprights the italic rank font: Windows OCR misreads the slanted
// EN label ("EMERALD 3" -> "fMfRßlD3") and reads it cleanly once deskewed.
const rankLabelShear = 0.25

// deskew shifts each row right by shear*y, so text leaning right stands upright.
func deskew(src image.Image, shear float64) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	out := image.NewRGBA(image.Rect(0, 0, w+int(shear*float64(h))+1, h))
	draw.Draw(out, out.Bounds(), image.Black, image.Point{}, draw.Src)
	for y := 0; y < h; y++ {
		off := int(shear * float64(y))
		for x := 0; x < w; x++ {
			out.Set(x+off, y, src.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return out
}

// whiteTextGray turns bright, low-saturation text black on white (2nd attempt
// when the plain crop is unreadable over a bright emblem glow).
func whiteTextGray(src image.Image) *image.Gray {
	b := src.Bounds()
	g := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, gg, bb, _ := src.At(x, y).RGBA()
			r8, g8, b8 := int(r>>8), int(gg>>8), int(bb>>8)
			mx, mn := max(r8, g8, b8), min(r8, g8, b8)
			v := uint8(255)
			if mn > 165 && mx-mn < 60 {
				v = 0
			}
			g.SetGray(x, y, color.Gray{Y: v})
		}
	}
	return g
}

// brightTextGray is whiteTextGray relative to the crop: the label is the
// brightest near-gray thing in it whatever the tint (OBS color correction, a
// blue-tinted white over the Diamond glow, a dimmed screen). Black on white.
func brightTextGray(src image.Image) *image.Gray {
	b := src.Bounds()
	var hist [256]int
	n := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, gg, bb, _ := src.At(x, y).RGBA()
			hist[min(r, gg, bb)>>8]++
			n++
		}
	}
	// The 99th percentile of the darkest channel is the text's own level.
	top, acc := 255, 0
	for top > 0 && acc+hist[top] < n/100 {
		acc += hist[top]
		top--
	}
	thr := max(60, top*4/5)
	g := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, gg, bb, _ := src.At(x, y).RGBA()
			mx, mn := int(max(r, gg, bb)>>8), int(min(r, gg, bb)>>8)
			v := uint8(255)
			if mn >= thr && (mx-mn)*100 <= 35*mx {
				v = 0
			}
			g.SetGray(x, y, color.Gray{Y: v})
		}
	}
	return g
}

// rankOCR is the OCR step of ReadRankScreen (a Windows OCR process in production).
type rankOCR interface {
	Recognize(pngPath string) (map[string]string, error)
}

// ReadRankScreen looks for the progress screen in a full game frame. It returns
// a zero RankRead (no error) when the frame is something else.
func ReadRankScreen(ocr rankOCR, workDir string, frame image.Image) (RankRead, error) {
	var r RankRead
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return r, err
	}
	path := filepath.Join(workDir, "rank_ocr.png")
	defer os.Remove(path)
	read := func(img image.Image, lang string) (map[string]string, error) {
		if err := writePNG(path, img); err != nil {
			return nil, err
		}
		if lo, ok := ocr.(rankLangOCR); ok && lang != "" {
			return lo.RecognizeLang(path, lang)
		}
		return ocr.Recognize(path)
	}

	role := rankCrop(frame, rankRoleCrop)
	if role == nil {
		return r, nil
	}
	texts, err := read(role, "")
	if err != nil {
		return r, err
	}
	// The language that reads the word best (in its own alphabet) is the
	// client's: the label is read in it below. "ТАНК" read by the English OCR is
	// "TAHK", one letter off "TANK"; the Russian one reads it exactly.
	bestScore := 99
	for _, lang := range sortedLangs(texts) {
		if id, score := matchRankRoleBest(texts[lang]); id != "" && score < bestScore {
			r.Role, r.RoleText, r.Lang, bestScore = id, strings.TrimSpace(texts[lang]), lang, score
		}
	}
	if r.Role == "" {
		return RankRead{}, nil // not the progress screen: skip the 2nd OCR
	}

	label := rankCrop(frame, rankLabelCrop)
	if label == nil {
		return r, nil
	}
	// The label is in the language the role was read in: one OCR language per
	// try (about half the time with two languages installed). The other
	// languages get one last try.
	type try struct {
		img  image.Image
		lang string
	}
	upright := deskew(label, rankLabelShear)
	tries := []try{{upright, r.Lang}, {label, r.Lang}, {whiteTextGray(label), r.Lang},
		{brightTextGray(upright), r.Lang}}
	if lo, ok := ocr.(rankLangOCR); ok && len(lo.Langs()) > 1 {
		tries = append(tries, try{upright, ""})
	}
	for _, t := range tries {
		texts, err := read(t.img, t.lang)
		if err != nil {
			return r, err
		}
		for _, lang := range sortedLangs(texts) {
			if t.lang == "" && lang == r.Lang {
				continue // already read above
			}
			txt := strings.TrimSpace(texts[lang])
			if txt != "" && r.RankText == "" {
				r.RankText = txt
			}
			if tier, div := matchRankLabel(txt); tier != "" {
				r.Tier, r.Division, r.Rank = tier, div, RankIndex(tier, div)
				r.RankText, r.Lang = txt, lang
				return r, nil
			}
		}
	}
	return r, nil
}

// rankLangOCR runs a single OCR language (the Windows OCR process does).
type rankLangOCR interface {
	RecognizeLang(pngPath, lang string) (map[string]string, error)
	Langs() []string
}

// sortedLangs gives a stable order (map iteration is random).
func sortedLangs(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
