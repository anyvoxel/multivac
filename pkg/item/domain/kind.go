package domain

import "strings"

type Kind string

const (
	KindInbox        Kind = "Inbox"
	KindTask         Kind = "Task"
	KindWaitingFor   Kind = "WaitingFor"
	KindSomedayMaybe Kind = "SomedayMaybe"
)

func (k Kind) Valid() bool {
	switch k {
	case KindInbox, KindTask, KindWaitingFor, KindSomedayMaybe:
		return true
	default:
		return false
	}
}

func ParseKind(v string) (Kind, bool) {
	switch strings.ToLower(strings.ReplaceAll(v, "_", "")) {
	case "inbox":
		return KindInbox, true
	case "task":
		return KindTask, true
	case "waitingfor":
		return KindWaitingFor, true
	case "somedaymaybe":
		return KindSomedayMaybe, true
	default:
		return "", false
	}
}
