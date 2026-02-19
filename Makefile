.PHONY: run run-redis clean

run:
	go run cmd/server/main.go

run-redis:
	docker-compose up -d redis

clean:
	docker-compose down
	rm -f ratex-demo

deps:
	go mod download
	go mod tidy