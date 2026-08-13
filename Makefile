include app.env

postgres:
	docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=$(POSTGRES_USER) -e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) -d postgres:12-alpine 

createdb:
	docker exec -it postgres12 createdb --username=$(POSTGRES_USER) --owner=$(POSTGRES_USER) billi_bank

dropdb:
	docker exec -it postgres12 dropdb billi_bank

migrateup:
	migrate -path db/migration -database "$(DB_SOURCE)" -verbose up

migrateup1:
	migrate -path db/migration -database "$(DB_SOURCE)" -verbose up 1

migratedown:
	migrate -path db/migration -database "$(DB_SOURCE)" -verbose down

migratedown1:
	migrate -path db/migration -database "$(DB_SOURCE)" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/billiraheem/Billi-Bank/db/sqlc Store

myapp-image:
	docker build -t billibank:latest .

run-myimage:
	docker run --name billibank -p 8080:8080 billibank:latest

run-myimage2:
	docker run --name billibank -p 8080:8080 -e GIN_MODE=release billibank:latest

my-network:
	docker network create bank-network

connect-network:
	docker network connect bank-network postgres12

# this ensures that the billbank and postgres 12 are on the same network
run-myimage3:
	docker run --name billibank --network bank-network -p 8080:8080 -e GIN_MODE=release -e DB_SOURCE=$(DB_SOURCE_IMAGE) billibank:latest

executable-start:
	chmod +x start.sh

migrateup-aws:
	migrate -path db/migration -database "$(DB_SOURCE_AWS)" -verbose up

db_docs:
	dbdocs build docs/db.dbml

db_schema:
	dbml2sql --postgres -o docs/schema.sql docs/db.dbml

proto:
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
    proto/*.proto

# this for when you want to regenerate the proto files when we make changes
proto_2:
	rm -f pb/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
    proto/*.proto

# this commands allows us to serve both gRPC and http request
proto_gateway:
	rm -f pb/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
    proto/*.proto

evans:
	evans --host localhost --port 9090 -r repl --package pb --service BilliBank

.PHONY: postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 sqlc test server mockdb myapp-image run-myimage run-myimage2 run-myimage3 my-network connect-network executable-start migrateup-aws db_docs db_schema proto proto_2 evans proto_gateway