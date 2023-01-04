createdb:
	docker exec -it simplebank-db createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it simplebank-db dropdb --username=root --owner=root simple_bank

postgres:
	docker run --name simplebank-db -e POSTGRES_USER=root --network bank-network -e POSTGRES_PASSWORD=secret -p 5455:5432 -d postgres

migrateup:
	 migrate -path db/migration -database "postgresql://root:secret@localhost:5455/simple_bank?sslmode=disable" -verbose up


migrateup1:
	 migrate -path db/migration -database "postgresql://root:secret@localhost:5455/simple_bank?sslmode=disable" -verbose up 1

psql:
	docker exec -it simplebank-db  psql -U root -d simple_bank


migratedown:
	 migrate -path db/migration -database "postgresql://root:secret@localhost:5455/simple_bank?sslmode=disable" -verbose down


migratedown1:
	 migrate -path db/migration -database "postgresql://root:secret@localhost:5455/simple_bank?sslmode=disable" -verbose down 1

docker-aws-login:
	aws ecr get-login-password | docker login --username AWS --password-stdin 567824320719.dkr.ecr.us-east-1.amazonaws.com
	
sqlc:
	sqlc generate

test:
	go test -v -cover ./...

format:
	go fmt ./...

mock:
	mockgen -destination db/mock/store.go --package=mockdb -source db/sqlc/store.go
	
server:
	go run .

keygen:
	@openssl rand -hex 64 | head -c 32

.PHONY: createdb postgres dropdb migrateup migrateup1 migratedown migratedown1 sqlc server mock
