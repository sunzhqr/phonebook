package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sunzhqr/phonebook/internal/logger"
	"github.com/sunzhqr/phonebook/internal/repository"
	"github.com/sunzhqr/phonebook/internal/service"
)

type mockRepo struct {
	CreateFn func(context.Context, repository.ContactInput) (repository.Contact, error)
	GetFn    func(context.Context, int64) (repository.Contact, error)
	UpdateFn func(context.Context, int64, repository.ContactPatch) (repository.Contact, error)
	DeleteFn func(context.Context, int64) error
	ListFn   func(context.Context, repository.ListFilter) ([]repository.Contact, int64, error)
	SearchFn func(context.Context, string, int) ([]repository.Contact, error)
}

func (m *mockRepo) Create(ctx context.Context, in repository.ContactInput) (repository.Contact, error) {
	return m.CreateFn(ctx, in)
}
func (m *mockRepo) Get(ctx context.Context, id int64) (repository.Contact, error) {
	return m.GetFn(ctx, id)
}
func (m *mockRepo) Update(ctx context.Context, id int64, p repository.ContactPatch) (repository.Contact, error) {
	return m.UpdateFn(ctx, id, p)
}
func (m *mockRepo) Delete(ctx context.Context, id int64) error { return m.DeleteFn(ctx, id) }
func (m *mockRepo) List(ctx context.Context, f repository.ListFilter) ([]repository.Contact, int64, error) {
	return m.ListFn(ctx, f)
}
func (m *mockRepo) Search(ctx context.Context, q string, limit int) ([]repository.Contact, error) {
	return m.SearchFn(ctx, q, limit)
}

func TestService_CreateContact_Normalizes_And_Primary(t *testing.T) {
	lg := logger.New("dev")
	now := time.Now().UTC()
	mr := &mockRepo{
		CreateFn: func(_ context.Context, in repository.ContactInput) (repository.Contact, error) {
			if len(in.Phones) != 1 || !in.Phones[0].IsPrimary {
				return repository.Contact{}, errors.New("primary invariant failed")
			}
			return repository.Contact{
				ID: 1, FirstName: "Sanzhar", LastName: "Sanzharov",
				Phones:    []repository.Phone{{Label: in.Phones[0].Label, PhoneRaw: in.Phones[0].PhoneRaw, PhoneE164: in.Phones[0].PhoneE164, IsPrimary: true}},
				CreatedAt: now, UpdatedAt: now,
			}, nil
		},
	}
	svc := service.New(lg, mr)

	out, err := svc.CreateContact(context.Background(), service.ContactCreateIn{
		FirstName: " Sanzhar ", LastName: "Sanzharrov", Company: "",
		Phones: []service.PhoneIn{
			{Label: "work", PhoneRaw: "+7 (771) 123-45-67", IsPrimary: true},
			{Label: "dup", PhoneRaw: "7 771 123 45 67", IsPrimary: false}, // это дубликат
		},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out.ID != 1 || len(out.Phones) != 1 || !out.Phones[0].IsPrimary {
		t.Fatalf("bad out: %+v", out)
	}
}

func TestService_UpdateContact_Pointers_Semantics(t *testing.T) {
	lg := logger.New("dev")
	now := time.Now().UTC()
	mr := &mockRepo{
		UpdateFn: func(_ context.Context, _ int64, p repository.ContactPatch) (repository.Contact, error) {
			if p.Phones == nil {
				return repository.Contact{ID: 42, CreatedAt: now, UpdatedAt: now}, nil
			}
			if len(*p.Phones) != 0 {
				return repository.Contact{}, errors.New("expected clear-all phones")
			}
			return repository.Contact{ID: 42, Phones: nil, CreatedAt: now, UpdatedAt: now}, nil
		},
	}
	svc := service.New(lg, mr)

	out, err := svc.UpdateContact(context.Background(), 42, service.ContactUpdateIn{})
	if err != nil || out.ID != 42 {
		t.Fatalf("nil phones: err=%v out=%+v", err, out)
	}

	empty := []service.PhoneIn{}
	out, err = svc.UpdateContact(context.Background(), 42, service.ContactUpdateIn{Phones: &empty})
	if err != nil || out.ID != 42 {
		t.Fatalf("empty phones: err=%v out=%+v", err, out)
	}
}

func TestService_List_Maps_Repo(t *testing.T) {
	//lg := logger.New("dev")
	now := time.Now().UTC()
	mr := &mockRepo{
		ListFn: func(_ context.Context, _ repository.ListFilter) ([]repository.Contact, int64, error) {
			return []repository.Contact{
				{ID: 1, FirstName: "A", LastName: "A", CreatedAt: now, UpdatedAt: now},
				{ID: 2, FirstName: "B", LastName: "B", CreatedAt: now, UpdatedAt: now},
			}, 0, nil
		},
	}
	svc := service.New(logger.New("dev"), mr)
	out, err := svc.ListContacts(context.Background(), service.ListFilter{Limit: 2})
	if err != nil || len(out.Items) != 2 {
		t.Fatalf("list err=%v out=%+v", err, out)
	}
}
