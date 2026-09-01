-- name: FetchUserOrders :many
select id, total_price, status, created_at from orders
where user_id = $1
order by created_at desc
limit $2
offset $3;

-- name: CreateOrder :one
insert into orders (user_id, total_price)
values ($1, $2)
returning id, total_price, status;

-- name: CreateEvent :one
insert into events(order_id, status, event_type)
values ($1, $2, $3)
returning id, status;

-- name: SelectEventForUpdate :many
select id, status from events
where status = 'payment_pending'
order by created_at
limit $1
for update skip locked;

-- name: UpdateEvent :exec
update events
set status = $2
where id = $1;

-- name: CreateOrderItems :exec
insert into order_items (order_id, item_id)
values ($1, $2);

-- name: FetchOrder :one
with item_counts as (
    select o.id as order_id, o.total_price, o.status, o.created_at,
        i.id as item_id, i.name, i.price, count(oi.item_id) as quantity
    from orders o
    left join order_items oi on o.id = oi.order_id
    left join items i on oi.item_id = i.id
    where o.id = $1 and o.user_id = $2
    group by o.id, i.id, i.name, i.price
)
select
    order_id as id,
    total_price,
    status,
    created_at,
    coalesce(
        json_agg(
            json_build_object('name', name, 'price', price, 'quantity', quantity)
            order by quantity desc
        ) filter (where item_id is not null),
        '[]'
    )::jsonb as items
from item_counts
group by order_id, total_price, status, created_at;

-- name: SelectOrderForUpdate :one
select id, user_id, status from orders
where id = $1 and user_id = $2 for update;
-- name: CancelOrder :exec
update orders
set status = 'cancelled'
where id = $1 and user_id = $2;
