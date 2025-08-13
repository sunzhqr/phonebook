package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ContactsRepository - интерфейс(контракт) для взаимодействия со слоем репозитория
type ContactsRepository interface {
	Create(ctx context.Context, in ContactInput) (Contact, error)
	Get(ctx context.Context, id int64) (Contact, error)
	Update(ctx context.Context, id int64, patch ContactPatch) (Contact, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, f ListFilter) ([]Contact, int64, error)
	Search(ctx context.Context, q string, limit int) ([]Contact, error)
}

type Repos struct {
	Contacts ContactsRepository
}

func New(pool *pgxpool.Pool) *Repos {
	return &Repos{
		Contacts: &contactRepo{pool: pool},
	}
}
