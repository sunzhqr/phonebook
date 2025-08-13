package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contactRepo struct {
	pool *pgxpool.Pool
}

func (r *contactRepo) Create(ctx context.Context, in ContactInput) (Contact, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Contact{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		id        int64
		createdAt time.Time
		updatedAt time.Time
	)
	if err := tx.QueryRow(ctx,
		`insert into contacts(first_name, last_name, company)
         values ($1, $2, $3)
         returning id, created_at, updated_at`,
		in.FirstName, in.LastName, in.Company,
	).Scan(&id, &createdAt, &updatedAt); err != nil {
		return Contact{}, err
	}

	// гарантируем единственный primary
	hasPrimary := false
	for i := range in.Phones {
		if in.Phones[i].IsPrimary {
			hasPrimary = true
			break
		}
	}
	if !hasPrimary && len(in.Phones) > 0 {
		in.Phones[0].IsPrimary = true
	}

	// пакетная вставка телефонов
	if len(in.Phones) > 0 {
		var b pgx.Batch
		for _, p := range in.Phones {
			b.Queue(
				`insert into contact_phones(contact_id, label, phone_raw, phone_e164, phone_digits, is_primary)
                 values ($1, $2, $3, $4, $5, $6)`,
				id, p.Label, p.PhoneRaw, p.PhoneE164, p.PhoneDigits, p.IsPrimary,
			)
		}
		if br := tx.SendBatch(ctx, &b); br != nil {
			if err := br.Close(); err != nil {
				return Contact{}, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Contact{}, err
	}
	phones, _ := r.getPhones(ctx, id)

	return Contact{
		ID:        id,
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Company:   in.Company,
		Phones:    phones,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func (r *contactRepo) Get(ctx context.Context, id int64) (Contact, error) {
	var c Contact
	err := r.pool.QueryRow(ctx,
		`select id, first_name, last_name, coalesce(company,''), created_at, updated_at
         from contacts where id=$1`,
		id,
	).Scan(&c.ID, &c.FirstName, &c.LastName, &c.Company, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Contact{}, ErrNotFound
		}
		return Contact{}, err
	}
	c.Phones, _ = r.getPhones(ctx, id)
	return c, nil
}

func (r *contactRepo) Update(ctx context.Context, id int64, p ContactPatch) (Contact, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Contact{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// частичное обновление скалярных полей
	if p.FirstName != nil || p.LastName != nil || p.Company != nil {
		set := make([]string, 0, 3)
		args := make([]any, 0, 4)
		idx := 1

		if p.FirstName != nil {
			set = append(set, fmt.Sprintf("first_name=$%d", idx))
			args = append(args, strings.TrimSpace(*p.FirstName))
			idx++
		}
		if p.LastName != nil {
			set = append(set, fmt.Sprintf("last_name=$%d", idx))
			args = append(args, strings.TrimSpace(*p.LastName))
			idx++
		}
		if p.Company != nil {
			set = append(set, fmt.Sprintf("company=$%d", idx))
			args = append(args, strings.TrimSpace(*p.Company))
			idx++
		}

		args = append(args, id)
		query := "update contacts set " + strings.Join(set, ",") + " where id=$" + fmt.Sprint(idx)
		if _, err := tx.Exec(ctx, query, args...); err != nil {
			return Contact{}, err
		}
	}

	// полная замена набора телефонов
	if p.Phones != nil {
		if _, err := tx.Exec(ctx, `delete from contact_phones where contact_id=$1`, id); err != nil {
			return Contact{}, err
		}
		phones := *p.Phones

		hasPrimary := false
		for i := range phones {
			if phones[i].IsPrimary {
				hasPrimary = true
				break
			}
		}
		if !hasPrimary && len(phones) > 0 {
			phones[0].IsPrimary = true
		}

		if len(phones) > 0 {
			var b pgx.Batch
			for _, ph := range phones {
				b.Queue(
					`insert into contact_phones(contact_id, label, phone_raw, phone_e164, phone_digits, is_primary)
                     values ($1, $2, $3, $4, $5, $6)`,
					id, ph.Label, ph.PhoneRaw, ph.PhoneE164, ph.PhoneDigits, ph.IsPrimary,
				)
			}
			if br := tx.SendBatch(ctx, &b); br != nil {
				if err := br.Close(); err != nil {
					return Contact{}, err
				}
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Contact{}, err
	}
	return r.Get(ctx, id)
}

func (r *contactRepo) Delete(ctx context.Context, id int64) error {
	ct, err := r.pool.Exec(ctx, `delete from contacts where id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *contactRepo) List(ctx context.Context, f ListFilter) ([]Contact, int64, error) {
	var sb strings.Builder
	args := make([]any, 0, 8)
	idx := 1

	sb.WriteString(`
select
  c.id,
  c.first_name,
  c.last_name,
  coalesce(c.company,''),
  c.created_at,
  c.updated_at
from contacts c
`)

	where := make([]string, 0, 4)
	if f.FirstName != "" {
		where = append(where, fmt.Sprintf("c.first_name ilike $%d", idx))
		args = append(args, "%"+f.FirstName+"%")
		idx++
	}
	if f.LastName != "" {
		where = append(where, fmt.Sprintf("c.last_name ilike $%d", idx))
		args = append(args, "%"+f.LastName+"%")
		idx++
	}
	if f.Company != "" {
		where = append(where, fmt.Sprintf("c.company ilike $%d", idx))
		args = append(args, "%"+f.Company+"%")
		idx++
	}
	if f.Phone != "" {
		sb.WriteString("join contact_phones p on p.contact_id = c.id\n")
		where = append(where, fmt.Sprintf("p.phone_digits ilike $%d", idx))
		args = append(args, "%"+digitsOnly(f.Phone)+"%")
		idx++
	}

	if len(where) > 0 {
		sb.WriteString("where " + strings.Join(where, " and ") + "\n")
	}

	// keyset
	if f.AfterID > 0 {
		cond := fmt.Sprintf("c.id > $%d", idx)
		if len(where) > 0 {
			sb.WriteString("and " + cond + "\n")
		} else {
			sb.WriteString("where " + cond + "\n")
		}
		args = append(args, f.AfterID)
		idx++
	}

	order := strings.ToLower(f.Order)
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	switch f.SortBy {
	case "name":
		sb.WriteString("order by c.last_name " + order + ", c.first_name " + order + ", c.id asc\n")
	case "created_at", "updated_at":
		sb.WriteString("order by c." + f.SortBy + " " + order + ", c.id asc\n")
	default:
		sb.WriteString("order by c.updated_at " + order + ", c.id asc\n")
	}

	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	sb.WriteString("limit " + fmt.Sprint(limit+1))

	rows, err := r.pool.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]Contact, 0, limit)
	var lastID int64
	for rows.Next() {
		var c Contact
		if err := rows.Scan(&c.ID, &c.FirstName, &c.LastName, &c.Company, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, c)
		lastID = c.ID
	}

	hasMore := len(list) > limit
	if hasMore {
		list = list[:limit]
	}

	// подтягиваем телефоны (MVP; можно оптимизировать батчем)
	for i := range list {
		ph, _ := r.getPhones(ctx, list[i].ID)
		list[i].Phones = ph
	}

	var next int64
	if hasMore {
		next = lastID
	}
	return list, next, nil
}

func (r *contactRepo) Search(ctx context.Context, q string, limit int) ([]Contact, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	q = strings.TrimSpace(q)
	if q == "" {
		return []Contact{}, nil
	}

	isPhone := isDigitsOrPlus(q)

	if isPhone {
		digits := digitsOnly(q)
		sql := `
select
  c.id,
  c.first_name,
  c.last_name,
  coalesce(c.company,''),
  c.created_at,
  c.updated_at,
  p.label,
  p.phone_raw,
  p.phone_e164,
  p.phone_digits,
  p.is_primary
from contacts c
join contact_phones p on p.contact_id = c.id
where p.phone_digits ilike '%' || $1 || '%'
order by c.updated_at desc, c.id asc
limit $2`
		rows, err := r.pool.Query(ctx, sql, digits, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		// Группируем совпавшие телефоны по контакту (и не подтягиваем прочие)
		byID := make(map[int64]*Contact)
		order := make([]int64, 0, limit)

		for rows.Next() {
			var (
				id int64
				c  Contact
				ph Phone
			)
			if err := rows.Scan(
				&id, &c.FirstName, &c.LastName, &c.Company, &c.CreatedAt, &c.UpdatedAt,
				&ph.Label, &ph.PhoneRaw, &ph.PhoneE164, &ph.PhoneDigits, &ph.IsPrimary,
			); err != nil {
				return nil, err
			}
			node, ok := byID[id]
			if !ok {
				c.ID = id
				c.Phones = []Phone{ph}
				byID[id] = &c
				order = append(order, id)
			} else {
				node.Phones = append(node.Phones, ph)
			}
		}

		out := make([]Contact, 0, len(order))
		for _, id := range order {
			out = append(out, *byID[id])
		}
		return out, nil
	}

	sql := `
select
  c.id, c.first_name, c.last_name, coalesce(c.company,''), c.created_at, c.updated_at
from contacts c
where (c.first_name || ' ' || c.last_name) ilike $1
order by similarity(c.first_name || ' ' || c.last_name, $1) desc, c.updated_at desc
limit $2`
	like := "%" + q + "%"
	rows, err := r.pool.Query(ctx, sql, like, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]Contact, 0, limit)
	for rows.Next() {
		var c Contact
		if err := rows.Scan(&c.ID, &c.FirstName, &c.LastName, &c.Company, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.Phones, _ = r.getPhones(ctx, c.ID)
		res = append(res, c)
	}
	return res, nil
}

func (r *contactRepo) getPhones(ctx context.Context, contactID int64) ([]Phone, error) {
	rows, err := r.pool.Query(ctx,
		`select label, phone_raw, phone_e164, phone_digits, is_primary
         from contact_phones
         where contact_id = $1
         order by is_primary desc, id asc`,
		contactID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Phone, 0, 4)
	for rows.Next() {
		var p Phone
		if err := rows.Scan(&p.Label, &p.PhoneRaw, &p.PhoneE164, &p.PhoneDigits, &p.IsPrimary); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}
