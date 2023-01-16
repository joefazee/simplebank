DB_URL=postgresql://root:secret@localhost:5455/simple_bank?sslmode=disable

createdb:
	docker exec -it simplebank-db createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it simplebank-db dropdb --username=root --owner=root simple_bank

postgres:
	docker run --name simplebank-db -e POSTGRES_USER=root --network bank-network -e POSTGRES_PASSWORD=secret -p 5455:5432 -d postgres

migrateup:
	 migrate -path db/migration -database "$(DB_URL)" -verbose up


migrateup1:
	 migrate -path db/migration -database "$(DB_URL)" -verbose up 1

psql:
	docker exec -it simplebank-db  psql -U root -d simple_bank


migratedown:
	 migrate -path db/migration -database "$(DB_URL)" -verbose down


migratedown1:
	 migrate -path db/migration -database "$(DB_URL)" -verbose down 1

docker-aws-login:
	aws ecr get-login-password | docker login --username AWS --password-stdin 346686984415.dkr.ecr.eu-west-1.amazonaws.com
	
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

proto:
	rm -f pb/*.go
	rm -f doc/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=simple_bank \
	proto/*.proto

evans:
	 evans --host localhost --port 4001  -r repl

.PHONY: createdb postgres dropdb migrateup migrateup1 migratedown migratedown1 sqlc server mock proto evans
