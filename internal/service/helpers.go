package service

import (
	"errors"
	"net/http"

	"github.com/sunzhqr/phonebook/internal/repository"
)

func (s *Service) repoErr(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return &Error{Code: http.StatusNotFound, Message: "not found"}
	case repository.IsBadRequest(err):
		return &Error{Code: http.StatusBadRequest, Message: err.Error()}
	default:
		return &Error{Code: http.StatusInternalServerError, Message: "internal"}
	}
}

func toContactOut(c repository.Contact) ContactOut {
	ph := make([]PhoneOut, 0, len(c.Phones))
	for _, p := range c.Phones {
		ph = append(ph, PhoneOut{Label: p.Label, PhoneRaw: p.PhoneRaw, PhoneE164: p.PhoneE164, IsPrimary: p.IsPrimary})
	}
	return ContactOut{ID: c.ID, FirstName: c.FirstName, LastName: c.LastName, Company: c.Company, Phones: ph, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}
}

//func rfc3339(t time.Time) string { return t.UTC().Format(time.RFC3339) }
