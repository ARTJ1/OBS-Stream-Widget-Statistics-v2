package owtracker

// Import-time: win/loss samples must differ by more than this.
const (
	hashDistanceThreshold = 5

	// Runtime default: streams/compression need a slightly higher bar.
	defaultMatchThreshold = 12
)

// WIN_HASH_TEMPLATE and LOSS_HASH_TEMPLATE alias the built-in defaults.
var (
	WIN_HASH_TEMPLATE  = defaultWinHash
	LOSS_HASH_TEMPLATE = defaultLossHash
)
