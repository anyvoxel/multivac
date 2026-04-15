package domain

import "strings"

type Bucket string

const (
	BucketInbox        Bucket = "Inbox"
	BucketNextAction   Bucket = "NextAction"
	BucketWaitingFor   Bucket = "WaitingFor"
	BucketSomedayMaybe Bucket = "SomedayMaybe"
	BucketCompleted    Bucket = "Completed"
	BucketDropped      Bucket = "Dropped"
)

func (b Bucket) Valid() bool {
	switch b {
	case BucketInbox, BucketNextAction, BucketWaitingFor, BucketSomedayMaybe, BucketCompleted, BucketDropped:
		return true
	default:
		return false
	}
}

func ParseBucket(v string) (Bucket, bool) {
	switch strings.ToLower(strings.ReplaceAll(v, "_", "")) {
	case "inbox":
		return BucketInbox, true
	case "nextaction":
		return BucketNextAction, true
	case "waitingfor":
		return BucketWaitingFor, true
	case "somedaymaybe":
		return BucketSomedayMaybe, true
	case "completed":
		return BucketCompleted, true
	case "dropped":
		return BucketDropped, true
	default:
		return "", false
	}
}
