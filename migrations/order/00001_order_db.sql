-- +goose Up
-- +goose StatementBegin
create table if not exists items (
    id uuid primary key default uuidv7(),
    name varchar(255) not null,
    description text not null,
    price decimal(10, 2) not null,
    created_at timestamp not null default (now() at time zone 'UTC'),
    updated_at timestamp not null default (now() at time zone 'UTC')
);
create type order_status as enum ('created', 'pending', 'completed', 'cancelled');
create table if not exists orders (
    id uuid primary key default uuidv7(),
    user_id uuid not null,
    total_price decimal(10, 2) not null,
    status order_status not null default 'created',
    created_at timestamp not null default (now() at time zone 'UTC'),
    updated_at timestamp not null default (now() at time zone 'UTC')
);
create table if not exists order_items (
    order_id uuid not null,
    item_id uuid not null,
    foreign key (order_id) references orders(id) on delete set null,
    foreign key (item_id) references items(id) on delete set null
);
create type event_type as enum ('payment_completed', 'payment_pending', 'payment_failed', 'payment_chargedback');
create table if not exists events (
    id uuid primary key default uuidv7(),
    order_id uuid not null,
    status order_status not null,
    event_type event_type not null,
    created_at timestamp not null default (now() at time zone 'UTC'),
    updated_at timestamp not null default (now() at time zone 'UTC')
);
create index if not exists idx_events_comp_events_status_created_at on events(status, created_at);
create index if not exists idx_order_items_order_id on order_items(order_id);
create index if not exists idx_orders_user_id on orders(user_id);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop table if exists events;
drop table if exists order_items;
drop table if exists orders;
drop type if exists event_type;
drop type if exists order_status;
--- +goose StatementEnd
