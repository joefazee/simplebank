createdb:
	docker exec -it simplebank-db createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it simplebank-db dropdb --username=root --owner=root simple_bank

postgres:
	docker run --name simplebank-db -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -p 5455:5432 -d postgres

migrateup:
	 migrate -path db/migration -database "postgresql://root:secret@localhost:5455/simple_bank?sslmode=disable" -verbose up


psql:
	docker exec -it simplebank-db  psql -U root -d simple_bank


migratedown:
	 migrate -path db/migration -database "postgresql://root:secret@localhost:5455/simple_bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

format:
	go fmt ./...

.PHONY: createdb postgres dropdb migrateup migratedown sqlc
