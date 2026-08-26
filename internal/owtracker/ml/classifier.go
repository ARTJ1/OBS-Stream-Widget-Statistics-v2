package ml

import "image"

// Label is a banner screen classification.
type Label string

const (
	LabelNone Label = "none"
	LabelWin  Label = "win"
	LabelLoss Label = "loss"
)

// Result is one banner crop classification.
type Result struct {
	Label      Label
	Confidence float32 // 0..1 softmax max
	WinProb    float32
	LossProb   float32
	NoneProb   float32
}

// BannerClassifier classifies end-screen banner crops (win / loss / none).
type BannerClassifier interface {
	Ready() bool
	StatusNote() string
	Classify(img image.Image) (Result, error)
}

var global BannerClassifier = &noopClassifier{}

func Init(dataDir string) {
	if c := loadMLP(dataDir); c != nil && c.Ready() {
		global = c
		return
	}
	global = &noopClassifier{}
}

func Ready() bool { return global.Ready() }

func StatusNote() string { return global.StatusNote() }

func Classify(img image.Image) (Result, error) { return global.Classify(img) }

type noopClassifier struct{}

func (n *noopClassifier) Ready() bool              { return false }
func (n *noopClassifier) StatusNote() string       { return "ml model not loaded (run scripts/ml/train_banner.py)" }
func (n *noopClassifier) Classify(image.Image) (Result, error) {
	return Result{Label: LabelNone}, nil
}
