package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// maxSearchDays ограничивает перебор для правил w и m,
// чтобы невыполнимое правило (например, m 31 2) не зациклило поиск.
const maxSearchDays = 366 * 10

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	date, err := time.Parse(DateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("invalid date %q: %w", dstart, err)
	}

	parts := strings.Split(repeat, " ")
	switch parts[0] {
	case "":
		return "", errors.New("empty repeat rule")
	case "d":
		return nextByDays(now, date, parts)
	case "y":
		return nextByYear(now, date, parts)
	case "w":
		return nextByWeekdays(now, date, parts)
	case "m":
		return nextByMonthDays(now, date, parts)
	default:
		return "", fmt.Errorf("unsupported repeat rule %q", repeat)
	}
}

func afterNow(date, now time.Time) bool {
	return truncateDay(date).After(truncateDay(now))
}

func truncateDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func nextByDays(now, date time.Time, parts []string) (string, error) {
	if len(parts) != 2 {
		return "", errors.New("rule d requires an interval")
	}
	interval, err := strconv.Atoi(parts[1])
	if err != nil || interval < 1 || interval > 400 {
		return "", fmt.Errorf("invalid interval %q", parts[1])
	}

	for {
		date = date.AddDate(0, 0, interval)
		if afterNow(date, now) {
			break
		}
	}
	return date.Format(DateFormat), nil
}

func nextByYear(now, date time.Time, parts []string) (string, error) {
	if len(parts) != 1 {
		return "", errors.New("rule y takes no arguments")
	}

	for {
		date = date.AddDate(1, 0, 0)
		if afterNow(date, now) {
			break
		}
	}
	return date.Format(DateFormat), nil
}

func nextByWeekdays(now, date time.Time, parts []string) (string, error) {
	if len(parts) != 2 {
		return "", errors.New("rule w requires weekdays")
	}

	var weekdays [7]bool
	for _, s := range strings.Split(parts[1], ",") {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 || n > 7 {
			return "", fmt.Errorf("invalid weekday %q", s)
		}
		weekdays[n%7] = true // в time.Weekday воскресенье — 0
	}

	return searchDay(now, date, func(d time.Time) bool {
		return weekdays[d.Weekday()]
	})
}

func nextByMonthDays(now, date time.Time, parts []string) (string, error) {
	if len(parts) != 2 && len(parts) != 3 {
		return "", errors.New("rule m requires days")
	}

	var days [32]bool // 1..31
	var lastDay, preLastDay bool
	for _, s := range strings.Split(parts[1], ",") {
		n, err := strconv.Atoi(s)
		switch {
		case err != nil:
			return "", fmt.Errorf("invalid month day %q", s)
		case n == -1:
			lastDay = true
		case n == -2:
			preLastDay = true
		case n >= 1 && n <= 31:
			days[n] = true
		default:
			return "", fmt.Errorf("invalid month day %q", s)
		}
	}

	var months [13]bool // 1..12
	if len(parts) == 3 {
		for _, s := range strings.Split(parts[2], ",") {
			n, err := strconv.Atoi(s)
			if err != nil || n < 1 || n > 12 {
				return "", fmt.Errorf("invalid month %q", s)
			}
			months[n] = true
		}
	} else {
		for i := range months {
			months[i] = true
		}
	}

	return searchDay(now, date, func(d time.Time) bool {
		if !months[d.Month()] {
			return false
		}
		// нулевой день следующего месяца — последний день текущего
		last := time.Date(d.Year(), d.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		return days[d.Day()] ||
			(lastDay && d.Day() == last) ||
			(preLastDay && d.Day() == last-1)
	})
}

// searchDay перебирает дни после большей из дат date и now
// и возвращает первый, подходящий под условие match.
func searchDay(now, date time.Time, match func(time.Time) bool) (string, error) {
	if afterNow(now, date) {
		date = now
	}
	date = truncateDay(date)

	for i := 0; i < maxSearchDays; i++ {
		date = date.AddDate(0, 0, 1)
		if match(date) {
			return date.Format(DateFormat), nil
		}
	}
	return "", errors.New("no matching date found")
}

func nextDayHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	if s := r.FormValue("now"); s != "" {
		var err error
		now, err = time.Parse(DateFormat, s)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid now %q", s), http.StatusBadRequest)
			return
		}
	}

	next, err := NextDate(now, r.FormValue("date"), r.FormValue("repeat"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write([]byte(next))
}
