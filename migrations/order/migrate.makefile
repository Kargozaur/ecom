.PHONY: migrate-order create-order-db

db_string=postgres://postgres:1234@localhost:5433/user_db

migrate-order:
	goose -dir ./migrations/order postgres $(db_string) up

create-order-db:
	docker exec -it db createdb -U postgres order_db
