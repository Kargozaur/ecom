-- +goose Up
-- +goose StatementBegin
create table if not exists categories(
    id uuid primary key default uuidv7(),
    name text not null
);
create table if not exists items (
    id uuid primary key,
    name text not null,
    price numeric(10, 2) not null,
    description text,
    quantity int not null,
    created_at timestamptz default (now() at time zone 'UTC')
);
create table if not exists items_categories(
    item_id uuid primary key not null,
    category_id uuid,
    foreign key (item_id) references items(id) on delete cascade,
    foreign key(category_id) references categories(id) on delete set null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists categories;
drop table if exists items;
drop table if exists items_categories;
-- +goose StatementEnd
