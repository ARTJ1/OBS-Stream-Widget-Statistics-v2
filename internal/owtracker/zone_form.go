package owtracker

import (
	"strconv"
	"strings"
)

func ParseZoneForm(x, y, w, h string) Zone {
	return Zone{
		XPct: parsePct(x),
		YPct: parsePct(y),
		WPct: parsePct(w),
		HPct: parsePct(h),
	}
}

func parsePct(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}
