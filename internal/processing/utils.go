package processing

import "time"

func isMarketHoliday(date time.Time, loc *time.Location) bool {
	if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		return true
	}

	return false
}
