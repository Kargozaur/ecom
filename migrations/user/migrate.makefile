.PHONY: migrate-user create-user-db

db_string=postgres://postgres:1234@localhost:5433/user_db
migrate-user:
	goose -dir ./migrations/user postgres $(db_string) up

create-user-db:
	docker exec -it db createdb -U postgres user_db
