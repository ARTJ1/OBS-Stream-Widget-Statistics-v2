package owtracker

// Perceptual hashes of the Overwatch 2 center victory / defeat banners.
// Fill these after a calibrated screenshot, or capture from the admin panel
// (saved to data/ow_phash.json and used at runtime).
// Empty strings disable matching until templates exist.
const (
	WIN_HASH_TEMPLATE  = ""
	LOSS_HASH_TEMPLATE = ""

	// Hamming distance: <= 5 is typically ~95%+ visual similarity for 64-bit pHash.
	hashDistanceThreshold = 5
)
