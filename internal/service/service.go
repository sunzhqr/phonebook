package service

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/sunzhqr/phonebook/internal/logger"
	"github.com/sunzhqr/phonebook/internal/repository"
)

// ContactsService - интерфейс(контракт) для взаимодействия со слоем сервиса
type ContactsService interface {
	CreateContact(ctx context.Context, in ContactCreateIn) (ContactOut, error)
	GetContact(ctx context.Context, id int64) (ContactOut, error)
	UpdateContact(ctx context.Context, id int64, in ContactUpdateIn) (ContactOut, error)
	DeleteContact(ctx context.Context, id int64) error
	ListContacts(ctx context.Context, f ListFilter) (ListOut, error)
	Search(ctx context.Context, q string, limit int) ([]ContactOut, error)
}

type Service struct {
	lg   *logger.Logger
	repo repository.ContactsRepository
	v    *validator.Validate
}

func New(lg *logger.Logger, repo repository.ContactsRepository) *Service {
	v := validator.New(validator.WithRequiredStructEnabled())
	return &Service{
		lg:   lg,
		repo: repo,
		v:    v,
	}
}
