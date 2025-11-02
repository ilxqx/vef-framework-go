package datetime

import "time"

var (
	// testLocation is the fixed timezone used for testing (UTC+8)
	testLocation = time.FixedZone("UTC+8", 8*60*60)
)

func testTime(year int, month time.Month, day, hour, min, sec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, 0, testLocation)
}

func testTimeUTC(year int, month time.Month, day, hour, min, sec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, 0, time.UTC)
}
