package bots

import "log"

type severity int

const (
	SevTrace severity = iota
	SevInfo
	SevWarn
	SevErr
)

func SeverityToPrefix(sev severity) string {
	switch sev {
	case SevTrace:
		return ":white_small_square:"
	case SevInfo:
		return ":information_source:"
	case SevWarn:
		return ":warning: @everyone"
	case SevErr:
		return ":rotating_light: @everyone"
	default:
		log.Fatalf("invalid severity %d\n", sev)
		return ""
	}
}
