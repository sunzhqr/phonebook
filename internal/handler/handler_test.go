package handler_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sunzhqr/phonebook/internal/handler"
	"github.com/sunzhqr/phonebook/internal/logger"
	"github.com/sunzhqr/phonebook/internal/service"
)

type mockSvc struct {
	CreateFn func(context.Context, service.ContactCreateIn) (service.ContactOut, error)
	GetFn    func(context.Context, int64) (service.ContactOut, error)
	UpdateFn func(context.Context, int64, service.ContactUpdateIn) (service.ContactOut, error)
	DeleteFn func(context.Context, int64) error
	ListFn   func(context.Context, service.ListFilter) (service.ListOut, error)
	SearchFn func(context.Context, string, int) ([]service.ContactOut, error)
}

func (m *mockSvc) CreateContact(ctx context.Context, in service.ContactCreateIn) (service.ContactOut, error) {
	return m.CreateFn(ctx, in)
}
func (m *mockSvc) GetContact(ctx context.Context, id int64) (service.ContactOut, error) {
	return m.GetFn(ctx, id)
}
func (m *mockSvc) UpdateContact(ctx context.Context, id int64, in service.ContactUpdateIn) (service.ContactOut, error) {
	return m.UpdateFn(ctx, id, in)
}
func (m *mockSvc) DeleteContact(ctx context.Context, id int64) error {
	return m.DeleteFn(ctx, id)
}
func (m *mockSvc) ListContacts(ctx context.Context, f service.ListFilter) (service.ListOut, error) {
	return m.ListFn(ctx, f)
}
func (m *mockSvc) Search(ctx context.Context, q string, limit int) ([]service.ContactOut, error) {
	return m.SearchFn(ctx, q, limit)
}

func router(h *handler.Handler) http.Handler {
	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/contacts", h.ListContacts)
		r.Get("/contacts/search", h.Search)
		r.Get("/contacts/{id}", h.GetContact)
		r.Post("/contacts", h.CreateContact)
		r.Put("/contacts/{id}", h.UpdateContact)
		r.Delete("/contacts/{id}", h.DeleteContact)
	})
	return r
}

func Test_Create_Get_Update_Delete_List_Search(t *testing.T) {
	now := time.Now().UTC()
	lg := logger.New("dev")
	ms := &mockSvc{
		CreateFn: func(_ context.Context, in service.ContactCreateIn) (service.ContactOut, error) {
			return service.ContactOut{ID: 1, FirstName: in.FirstName, LastName: in.LastName,
				Phones:    []service.PhoneOut{{Label: "work", PhoneRaw: in.Phones[0].PhoneRaw, PhoneE164: "+77711234567", IsPrimary: true}},
				CreatedAt: now, UpdatedAt: now}, nil
		},
		GetFn: func(_ context.Context, id int64) (service.ContactOut, error) {
			return service.ContactOut{ID: id, FirstName: "Sanzhar", LastName: "Sanzharov", CreatedAt: now, UpdatedAt: now}, nil
		},
		UpdateFn: func(_ context.Context, id int64, in service.ContactUpdateIn) (service.ContactOut, error) {
			company := ""
			if in.Company != nil {
				company = *in.Company
			}
			return service.ContactOut{ID: id, FirstName: "Sanzhar", LastName: "Sanzharov", Company: company, CreatedAt: now, UpdatedAt: now}, nil
		},
		DeleteFn: func(_ context.Context, _ int64) error { return nil },
		ListFn: func(_ context.Context, f service.ListFilter) (service.ListOut, error) {
			return service.ListOut{
				Items: []service.ContactOut{{ID: 1}, {ID: 2}},
				Page:  service.PageOut{NextAfterID: 2, HasMore: false, Limit: f.Limit},
			}, nil
		},
		SearchFn: func(_ context.Context, q string, _ int) ([]service.ContactOut, error) {
			// возвращаем только совпавший номер
			return []service.ContactOut{{ID: 10, Phones: []service.PhoneOut{{PhoneE164: "+77711234567"}}}}, nil
		},
	}
	h := handler.New(lg, ms)
	ts := httptest.NewServer(router(h))
	defer ts.Close()

	// create
	body := []byte(`{"first_name":"Sanzhar","last_name":"Sanzharov","phones":[{"label":"work","phone_raw":"+77711234567","is_primary":true}]}`)
	res, err := http.Post(ts.URL+"/api/v1/contacts", "application/json", bytes.NewReader(body))
	if err != nil || res.StatusCode != http.StatusCreated {
		t.Fatalf("create status=%v err=%v", res.StatusCode, err)
	}

	// get
	if res, _ := http.Get(ts.URL + "/api/v1/contacts/1"); res.StatusCode != http.StatusOK {
		t.Fatalf("get %v", res.Status)
	}

	// update
	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/api/v1/contacts/1", bytes.NewReader([]byte(`{"company":"ForteBank"}`)))
	req.Header.Set("Content-Type", "application/json")
	if res, _ := http.DefaultClient.Do(req); res.StatusCode != http.StatusOK {
		t.Fatalf("update %v", res.Status)
	}

	// list
	if res, _ := http.Get(ts.URL + "/api/v1/contacts?limit=2"); res.StatusCode != http.StatusOK {
		t.Fatalf("list %v", res.Status)
	}

	// search by phone
	if res, _ := http.Get(ts.URL + "/api/v1/contacts/search?q=%2B7771"); res.StatusCode != http.StatusOK {
		t.Fatalf("search %v", res.Status)
	}

	// delete
	req, _ = http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/contacts/1", nil)
	if res, _ := http.DefaultClient.Do(req); res.StatusCode != http.StatusNoContent {
		t.Fatalf("delete %v", res.Status)
	}
}
