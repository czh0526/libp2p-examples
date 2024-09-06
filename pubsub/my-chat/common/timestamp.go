package common

import "time"

func GetTimestampString(t time.Time) string {
	return t.Format(time.RFC3339)
}
