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
		got := []rune(c)
		if len(got) < 2 {
			continue
		}
		for role, words := range rankRoleWords {
			for _, w := range words {
				wr := []rune(w)
				if abs(len(got)-len(wr)) > 2 {
					continue
				}
				d := runeLevenshtein(got, wr)
				if d <= maxWordDistance(wr) && d < bestDist {
					best, bestDist = role, d
				}
			}
		}
	}
	return best
}

// matchRankLabel parses "EMERALD 3" / "ИЗУМРУД 2" / "DIRMOND5" into tier + division.
func matchRankLabel(text string) (string, int) {
	letters := []rune(normalizeBannerText(text))
	if len(letters) < 2 || len(letters) > 16 {
		return "", 0
	}
	div := rankDivision(text)
	if div == 0 {
		return "", 0
	}
	letters = foldRankGlyphs(letters)
	best, bestDist := "", 99
	for _, t := range rankTiers {
		for _, w := range t.Words {
			wr := foldRankGlyphs([]rune(w))
			if abs(len(letters)-len(wr)) > 2 {
				continue
			}
			d := runeLevenshtein(letters, wr)
			if d <= maxWordDistance(wr) && d < bestDist {
				best, bestDist = t.Name, d
			}
		}
	}
	if best == "" {
		return "", 0
	}
	return best, div
}

// rankGlyphFold merges letters Windows OCR swaps in the italic rank font
// (ПЛАТИНА -> ППЯТИНЯ, EMERALD -> FMFRALD); applied to both sides of a match.
var rankGlyphFold = map[rune]rune{'Л': 'П', 'Я': 'А', 'F': 'E'}

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
var rankDivisionGlyphs = map[string]int{"З": 3, "з": 3, "I": 1, "l": 1, "|": 1}

// rankDivision is the last 1–5 digit of the label (OCR sometimes glues it on).
func rankDivision(text string) int {
	// A lone last token that only looks like a digit ("ИЗУМРУД З").
	if toks := strings.Fields(text); len(toks) > 1 {
		if d, ok := rankDivisionGlyphs[toks[len(toks)-1]]; ok {
			return d
		}
	}
	rs := []rune(strings.TrimSpace(text))
	for i := len(rs) - 1; i >= 0; i-- {
		r := rs[i]
		if r >= '1' && r <= '5' {
			return int(r - '0')
		}
		if unicode.IsDigit(r) {
			return 0 // 0, 6–9: not a division
		}
	}
	return 0
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
	read := func(img image.Image) (map[string]string, error) {
		if err := writePNG(path, img); err != nil {
			return nil, err
		}
		return ocr.Recognize(path)
	}

	role := rankCrop(frame, rankRoleCrop)
	if role == nil {
		return r, nil
	}
	texts, err := read(role)
	if err != nil {
		return r, err
	}
	for _, lang := range sortedLangs(texts) {
		if id := matchRankRole(texts[lang]); id != "" {
			r.Role, r.RoleText = id, strings.TrimSpace(texts[lang])
			break
		}
	}
	if r.Role == "" {
		return RankRead{}, nil // not the progress screen: skip the 2nd OCR
	}

	label := rankCrop(frame, rankLabelCrop)
	if label == nil {
		return r, nil
	}
	for _, img := range []image.Image{deskew(label, rankLabelShear), label, whiteTextGray(label)} {
		texts, err := read(img)
		if err != nil {
			return r, err
		}
		for _, lang := range sortedLangs(texts) {
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
