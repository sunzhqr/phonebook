create extension if not exists pg_trgm;

create table if not exists contacts (
    id          bigserial primary key,
    first_name  text not null,
    last_name   text not null,
    company     text,
    created_at  timestamptz not null default now(),
    updated_at  timestamptz not null default now()
);

create table if not exists contact_phones (
  id           bigserial primary key,
  contact_id   bigint not null references contacts(id) on delete cascade,
    label        text,
    phone_raw    text not null,
    phone_e164   text not null,
    phone_digits text not null,
    is_primary   boolean not null default false
);

create or replace function set_updated_at() returns trigger as $$
begin new.updated_at = now(); return new; end; $$ language plpgsql;
create trigger trg_contacts_updated before update on contacts for each row execute function set_updated_at();

create index if not exists idx_contacts_name_trgm on contacts using gin ((first_name || ' ' || last_name) gin_trgm_ops);
create index if not exists idx_phones_digits_trgm on contact_phones using gin (phone_digits gin_trgm_ops);
create unique index if not exists uq_contact_primary_phone on contact_phones (contact_id) where is_primary;