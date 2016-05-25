package parser

import (
	"errors"
	"github.com/itsabot/abot/shared/datatypes"
	"strconv"
	"strings"
	"time"
)

type timeStruct struct {
	Year    int
	Day     string
	Month   int
	Hour    int
	Minutes int
}

type timeValues struct {
	When     string // today, tomorrow, yesterday
	Meridiem string // am, pm
	Day      string // sunday, monday, tuesday, etc
	Amount   int    // days to move ahead or back
	Forward  int    // 1 (next), 0 (not set), -1 (last)
	Hour     int
	Minutes  int
}

func extractTime(in *dt.Msg) (time.Time, error) {
	tv, err := parseTime(in)
	if err != nil {
		return time.Now(), err
	}

	// construct time
	date := time.Now()

	/*
	* Date
	 */
	// Day
	if strings.EqualFold(tv.When, "today") {
		// don't care
	} else if strings.EqualFold(tv.When, "tomorrow") {
		date.AddDate(0, 0, 1)
	} else if strings.EqualFold(tv.When, "yesterday") {
		date.AddDate(0, 0, -1)
	}
	// Month

	// Year

	/*
	* Time
	 */

	return date, nil
}

func parseTime(in *dt.Msg) (*timeValues, error) {
	tv := &timeValues{}

	words := strings.Split(strings.ToLower(in.Sentence), " ")
	for i := 0; i < len(words); i++ {

		word := words[i]
		if strings.EqualFold(word, "today") {
			tv.When = "today"
		} else if strings.EqualFold(word, "tomorrow") {
			tv.When = "tomorrow"
		} else if strings.EqualFold(word, "yesterday") {
			tv.When = "yesterday"
		} else if strings.Contains(word, "sun") {
			tv.Day = "sunday"
		} else if strings.Contains(word, "mon") {
			tv.Day = "monday"
		} else if strings.Contains(word, "tues") {
			tv.Day = "tuesday"
		} else if strings.Contains(word, "wednes") {
			tv.Day = "wednesday"
		} else if strings.Contains(word, "thurs") {
			tv.Day = "thursday"
		} else if strings.Contains(word, "fri") {
			tv.Day = "friday"
		} else if strings.Contains(word, "satur") {
			tv.Day = "saturday"
		} else if strings.EqualFold(word, "morning") {
			tv.Hour = 8
		} else if strings.EqualFold(word, "afternoon") ||
			strings.EqualFold(word, "noon") {
			tv.Hour = 12
		} else if strings.EqualFold(word, "evening") {
			tv.Hour = 18
		} else if strings.EqualFold(word, "midnight") {
			tv.Hour = 0
		} else if strings.EqualFold(word, "week") {
			tv.When = "week"
		} else if strings.EqualFold(word, "weekend") {
			tv.When = "weekend"
		} else if strings.EqualFold(word, "next") {
			tv.Forward = 1
		} else if strings.EqualFold(word, "last") {
			tv.Forward = -1
		} else if strings.EqualFold(word, "fortnight") {
			tv.Amount = 14
		} else if strings.EqualFold(word, "score") {
			tv.Amount = tv.Amount * 20
		} else if strings.Contains(word, ":") {
			t := strings.Split(word, ":")

			tv.Hour, _ = strconv.Atoi(t[0])
			if strings.Contains(t[1], "m") {
				tv.Minutes, _ = strconv.Atoi(t[1][:2])

				m := t[1][2:]
				if strings.EqualFold(m, "pm") {
					if tv.Hour < 12 {
						tv.Hour += 12
					} else if tv.Hour == 12 {
						tv.Hour = 0
					}
				}
			} else {
				tv.Minutes, _ = strconv.Atoi(t[1])
			}
		} else if strings.EqualFold(word, "pm") {
			if tv.Hour < 12 {
				tv.Hour += 12
			} else if tv.Hour == 12 {
				tv.Hour = 0
			} // else time is before 12
		} else {
			return tv, errors.New("No time found")
		}

		return tv, nil
	}

	return nil, errors.New("could not parse time")
}
