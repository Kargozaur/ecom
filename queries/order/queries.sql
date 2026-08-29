-- name: FetchUserOrders :many
select * from orders
where user_id = $1
order by created_at desc
limit $2
offset $3;

-- name: CreateOrder :one
insert into orders (user_id, total_price)
values ($1, $2)
returning id, total_price, status;

-- name: FetchOrder :one
with item_counts as (
    select o.id as order_id, o.total_price, o.status, o.created_at,
        i.id as item_id, i.name, i.price, count(*) as quantity
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
    ) as items
from item_counts
group by order_id, total_price, status, created_at;
