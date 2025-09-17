package utils

import (
	"time"

	"github.com/jinzhu/now"
)

// ParseTime parses time strings in local timezone.
func ParseTime(texts ...string) (time.Time, error) {
	return now.ParseInLocation(time.Local, texts...)
}
