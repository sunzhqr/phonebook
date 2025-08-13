package service

import (
	"context"
	"net/http"
	"strings"

	"github.com/sunzhqr/phonebook/internal/repository"
	"github.com/sunzhqr/phonebook/pkg/normalizer"
)

// CreateContact — нормализует вход, обеспечивает инварианты и делегирует репозиторию.
func (s *Service) CreateContact(ctx context.Context, in ContactCreateIn) (ContactOut, error) {
	if err := s.v.Struct(in); err != nil {
		return ContactOut{}, &Error{Code: http.StatusUnprocessableEntity, Message: err.Error()}
	}

	seen := make(map[string]struct{}, len(in.Phones))
	phones := make([]repository.PhoneInput, 0, len(in.Phones))
	hasPrimary := false

	for _, ph := range in.Phones {
		label := strings.TrimSpace(ph.Label)
		raw := strings.TrimSpace(ph.PhoneRaw)

		e164, digits, ok := normalizer.NormalizePhone(raw)
		if !ok {
			return ContactOut{}, &Error{Code: http.StatusUnprocessableEntity, Message: "invalid phone"}
		}

		// дубликаты по digits отбрасываем
		if _, dup := seen[digits]; dup {
			continue
		}
		seen[digits] = struct{}{}

		if ph.IsPrimary {
			hasPrimary = true
		}

		phones = append(phones, repository.PhoneInput{
			Label:       label,
			PhoneRaw:    raw,
			PhoneE164:   e164,
			PhoneDigits: digits,
			IsPrimary:   ph.IsPrimary,
		})
	}

	if len(phones) == 0 {
		return ContactOut{}, &Error{Code: http.StatusUnprocessableEntity, Message: "no valid phones"}
	}
	if !hasPrimary {
		phones[0].IsPrimary = true
	}

	c, err := s.repo.Create(ctx, repository.ContactInput{
		FirstName: strings.TrimSpace(in.FirstName),
		LastName:  strings.TrimSpace(in.LastName),
		Company:   strings.TrimSpace(in.Company),
		Phones:    phones,
	})
	if err != nil {
		return ContactOut{}, s.repoErr(err)
	}
	return toContactOut(c), nil
}

func (s *Service) GetContact(ctx context.Context, id int64) (ContactOut, error) {
	c, err := s.repo.Get(ctx, id)
	if err != nil {
		return ContactOut{}, s.repoErr(err)
	}
	return toContactOut(c), nil
}

func (s *Service) UpdateContact(ctx context.Context, id int64, in ContactUpdateIn) (ContactOut, error) {
	if err := s.v.Struct(in); err != nil {
		return ContactOut{}, &Error{Code: http.StatusUnprocessableEntity, Message: err.Error()}
	}

	var phones *[]repository.PhoneInput
	if in.Phones != nil {
		arr := make([]repository.PhoneInput, 0, len(*in.Phones))
		hasPrimary := false
		for _, ph := range *in.Phones {
			e164, digits, ok := normalizer.NormalizePhone(ph.PhoneRaw)
			if !ok {
				return ContactOut{}, &Error{Code: http.StatusUnprocessableEntity, Message: "invalid phone"}
			}
			if ph.IsPrimary {
				hasPrimary = true
			}
			arr = append(arr, repository.PhoneInput{Label: ph.Label, PhoneRaw: ph.PhoneRaw, PhoneE164: e164, PhoneDigits: digits, IsPrimary: ph.IsPrimary})
		}
		if !hasPrimary && len(arr) > 0 {
			arr[0].IsPrimary = true
		}
		phones = &arr
	}

	c, err := s.repo.Update(ctx, id, repository.ContactPatch{
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Company:   in.Company,
		Phones:    phones,
	})
	if err != nil {
		return ContactOut{}, s.repoErr(err)
	}
	return toContactOut(c), nil
}

func (s *Service) DeleteContact(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return s.repoErr(err)
	}
	return nil
}

func (s *Service) ListContacts(ctx context.Context, f ListFilter) (ListOut, error) {
	// Белый список сортировок для устойчивости API
	sort := map[string]string{"created_at": "created_at", "updated_at": "updated_at", "name": "name"}
	order := map[string]string{"asc": "asc", "desc": "desc"}
	sby, ok := sort[strings.ToLower(f.Sort)]
	if !ok {
		sby = "updated_at"
	}
	ord, ok := order[strings.ToLower(f.Order)]
	if !ok {
		ord = "desc"
	}

	res, next, err := s.repo.List(ctx, repository.ListFilter{
		FirstName: f.FirstName,
		LastName:  f.LastName,
		Company:   f.Company,
		Phone:     f.Phone,
		AfterID:   f.AfterID,
		Limit:     f.Limit,
		SortBy:    sby,
		Order:     ord,
	})
	if err != nil {
		return ListOut{}, s.repoErr(err)
	}

	items := make([]ContactOut, 0, len(res))
	for _, c := range res {
		items = append(items, toContactOut(c))
	}

	return ListOut{Items: items, Page: PageOut{NextAfterID: next, HasMore: next > 0, Limit: f.Limit}}, nil
}

func (s *Service) Search(ctx context.Context, q string, limit int) ([]ContactOut, error) {
	res, err := s.repo.Search(ctx, q, limit)
	if err != nil {
		return nil, s.repoErr(err)
	}
	out := make([]ContactOut, 0, len(res))
	for _, c := range res {
		out = append(out, toContactOut(c))
	}
	return out, nil
}
