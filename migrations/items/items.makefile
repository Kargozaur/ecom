.PHONY: items-migrate items-createdb
db_string=postgres://postgres:1234@localhost:5433/items_db

items-migrate:
	goose -dir migrations/items postgres $(db_string) up

items-createdb:
	docker exec -it db createdb -U postgres items_db
