package http

import (
	"taskTracker/internal/domain"
	"time"
)

type RecurrenceDTO struct {
	Type       string   `json:"type"`
	Interval   int      `json:"interval,omitempty"`
	DayOfMonth int      `json:"day_of_month,omitempty"`
	Specifics  []string `json:"specifics,omitempty"`
}

func (rd *RecurrenceDTO) ToDomain() *domain.TaskRecurrence {
	if rd == nil || rd.Type == "" {
		return nil
	}

	var specTimes []time.Time
	if len(rd.Specifics) > 0 {
		for _, s := range rd.Specifics {
			if t, err := time.Parse(time.RFC3339, s); err == nil {
				specTimes = append(specTimes, t)
			}
		}
	}

	return &domain.TaskRecurrence{
		Type:       domain.RecurrenceType(rd.Type),
		Interval:   rd.Interval,
		DayOfMonth: rd.DayOfMonth,
		Specifics:  specTimes,
	}
}

type RecordExecutionRequest struct {
	Date   time.Time         `json:"date"`
	Status domain.TaskStatus `json:"status"`
}

func (r *RecordExecutionRequest) Validate() error {
	if r.Date.IsZero() {
		return domain.ErrDateReq
	}
	if r.Status == "" {
		return domain.ErrStatusReq
	}
	return nil
}
