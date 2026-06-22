package domain

import (
	"errors"
	"time"
)

type RecurrenceType string

const (
	RecurrenceDaily   RecurrenceType = "DAILY"
	RecurrenceMonthly RecurrenceType = "MONTHLY"
	RecurrenceDates   RecurrenceType = "DATES"
	RecurrenceEven    RecurrenceType = "EVEN"
	RecurrenceOdd     RecurrenceType = "ODD"
)

var (
	ErrDateReq   = errors.New("date is required")
	ErrStatusReq = errors.New("status is required")
	ErrWrongRec  = errors.New("recurrence invalid")
)

type TaskRecurrence struct {
	Type       RecurrenceType
	Interval   int         // for DAILY
	DayOfMonth int         // for MONTHLY
	Specifics  []time.Time // for DATES
}

// checks if reccuring task happens on targetData
func (r *TaskRecurrence) IsMatch(baseDueDate time.Time, targetDate time.Time) bool {
	start := time.Date(baseDueDate.Year(), baseDueDate.Month(), baseDueDate.Day(), 0, 0, 0, 0, baseDueDate.Location())
	target := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())

	if target.Before(start) {
		return false
	}

	switch r.Type {
	case RecurrenceDaily:
		if r.Interval <= 0 {
			return false
		}
		daysDiff := int(target.Sub(start).Hours() / 24)
		return daysDiff%r.Interval == 0

	case RecurrenceMonthly:
		return target.Day() == r.DayOfMonth

	case RecurrenceEven:
		return target.Day()%2 == 0

	case RecurrenceOdd:
		return target.Day()%2 != 0

	case RecurrenceDates:
		for _, specDate := range r.Specifics {
			if specDate.Year() == target.Year() && specDate.Month() == target.Month() && target.Day() == specDate.Day() {
				return true
			}
		}
		return false
	}

	return false
}

func (t Task) IsRecurring() bool {
	return t.Recurrence != nil
}
