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
insert into order_items (order_id, item_id, item_name, item_price, quantity)
select $1::uuid, rows.item_id, rows.item_name, rows.item_price, rows.quantity
from rows from (
    unnest($2::uuid[]),
    unnest($3::text[]),
    unnest($4::numeric[]),
    unnest($5::int[])
) as rows(item_id, item_name, item_price, quantity);
-- name: FetchOrder :one
select o.id, o.total_price, o.status, o.created_at,
    coalesce(
        (
            select json_agg(
                json_build_object(
                    'name', oi.item_name,
                    'price', oi.item_price,
                    'quantity', oi.quantity
                )
                order by oi.quantity desc
            )
            from order_items oi
            where oi.order_id = o.id
        ),
        '[]'
    )::jsonb as items
from orders o
where o.id = $1 and o.user_id = $2;

-- name: SelectOrderForUpdate :one
select id, user_id, status from orders
where id = $1 and user_id = $2 for update;
-- name: CancelOrder :exec
update orders
set status = 'cancelled'
where id = $1 and user_id = $2;
