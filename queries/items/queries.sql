-- name: GetItem :one
select i.id, i.name, i.description, c.name as category_name from items i
join items_categories ic on i.id = ic.item_id
join categories c on ic.category_id = c.id
where i.id = $1;

-- name: GetItems :many
select i.id, i.name, i.description, c.name as category_name from items i
join items_categories ic on i.id = ic.item_id
join categories c on ic.category_id = c.id
where c.name = any(sqlc.slice('categories'))
group by i.id;
