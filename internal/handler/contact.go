package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sunzhqr/phonebook/internal/logger"
	"github.com/sunzhqr/phonebook/internal/service"
)

type Handler struct {
	lg  *logger.Logger
	svc service.ContactsService
}

func New(lg *logger.Logger, svc service.ContactsService) *Handler { return &Handler{lg: lg, svc: svc} }

func (h *Handler) CreateContact(w http.ResponseWriter, r *http.Request) {
	var dto ContactCreateDTO
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&dto); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	in := service.ContactCreateIn{FirstName: dto.FirstName, LastName: dto.LastName, Company: dto.Company}
	in.Phones = make([]service.PhoneIn, 0, len(dto.Phones))
	for _, p := range dto.Phones {
		in.Phones = append(in.Phones, service.PhoneIn{Label: p.Label, PhoneRaw: p.PhoneRaw, IsPrimary: p.IsPrimary})
	}
	res, err := h.svc.CreateContact(r.Context(), in)
	if err != nil {
		writeSvcErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, res)
}

func (h *Handler) GetContact(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	res, err := h.svc.GetContact(r.Context(), id)
	if err != nil {
		writeSvcErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) UpdateContact(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	var dto ContactUpdateDTO
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&dto); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	in := service.ContactUpdateIn{
		FirstName: dto.FirstName,
		LastName:  dto.LastName,
		Company:   dto.Company,
	}

	if dto.Phones != nil { // ← защита от nil
		arr := make([]service.PhoneIn, 0, len(*dto.Phones))
		for _, p := range *dto.Phones {
			arr = append(arr, service.PhoneIn{
				Label: p.Label, PhoneRaw: p.PhoneRaw, IsPrimary: p.IsPrimary,
			})
		}
		in.Phones = &arr
	}

	res, err := h.svc.UpdateContact(r.Context(), id, in)
	if err != nil {
		writeSvcErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) DeleteContact(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteContact(r.Context(), id); err != nil {
		writeSvcErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListContacts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	afterID, _ := strconv.ParseInt(q.Get("after_id"), 10, 64)
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	filter := service.ListFilter{FirstName: q.Get("first_name"), LastName: q.Get("last_name"), Company: q.Get("company"), Phone: q.Get("phone"), AfterID: afterID, Limit: limit, Sort: q.Get("sort"), Order: q.Get("order")}
	res, err := h.svc.ListContacts(r.Context(), filter)
	if err != nil {
		writeSvcErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	res, err := h.svc.Search(r.Context(), q, limit)
	if err != nil {
		writeSvcErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeSvcErr(w http.ResponseWriter, err error) {
	if se, ok := err.(*service.Error); ok {
		http.Error(w, se.Message, se.Code)
		return
	}
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
