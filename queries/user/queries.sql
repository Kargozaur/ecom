-- name: CreateUser :one
insert into users (email, name, password_hash)
values($1, $2, $3)
returning id, name, created_at;

-- name: FetchProfile :one
select id, email, password_hash, name, created_at from users
where id = $1;

-- name: FetchPassword :one
select id, email, password_hash from users
where email = $1;

-- name: CreateToken :exec
insert into refresh_tokens (user_id, token_hash)
values($1, $2);

-- name: GetTokenForUpdate :one
select token_hash, user_id, deleted from refresh_tokens
where token_hash = $1 and deleted is false for update;

-- name: DeleteToken :exec
update refresh_tokens
set deleted = true
where token_hash = $1 and user_id = $2 and deleted is false;
