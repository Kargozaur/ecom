-- name: GetItem :one
select i.id, i.name, i.description, i.price,
    array_agg(c.name)::text[] as category_names
from items i
join items_categories ic on i.id = ic.item_id
join categories c on ic.category_id = c.id
where i.id = $1;

-- name: GetItems :many
select i.id, i.name, i.description, i.price,
    array_agg(c.name)::text[] as category_names
from items i
join items_categories ic on i.id = ic.item_id
join categories c on ic.category_id = c.id
where c.name = any(sqlc.slice('categories'))
group by i.id;
