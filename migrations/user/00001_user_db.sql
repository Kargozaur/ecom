-- +goose Up
-- +goose StatementBegin
create table if not exists users (
    id uuid primary key default uuidv7(),
    email text not null unique,
    name varchar(100) not null,
    password_hash text not null,
    created_at date default (now() at time zone 'UTC'),
    updated_at date default (now() at time zone 'UTC')
);

create table if not exists refresh_tokens (
    id uuid primary key default uuidv7(),
    user_id uuid,
    token_hash text,
    deleted bool default false,
    created_at timestamptz default (now() at time zone 'UTC'),
    foreign key (user_id) references users(id) on delete cascade
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists users;
drop table if exists refresh_tokens;
-- +goose StatementEnd
