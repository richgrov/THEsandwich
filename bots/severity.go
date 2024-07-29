package bots

import "log"

type severity int

const (
	SevTrace severity = iota
	SevInfo
	SevWarn
	SevErr
)

func SeverityToEmote(sev severity) string {
	switch sev {
	case SevTrace:
		return ":white_small_square:"
	case SevInfo:
		return ":information_source:"
	case SevWarn:
		return ":warning:"
	case SevErr:
		return ":rotating_light:"
	default:
		log.Fatalf("invalid severity %d\n", sev)
		return ""
	}
}
