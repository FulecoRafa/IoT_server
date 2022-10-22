docker: door rain
	docker compose up --build -d

door:
	GOOS=linux GOARCH=amd64 go build -o cmd/door/door cmd/door/main.go

rain:
	GOOS=linux GOARCH=amd64 go build -o cmd/rain/rain cmd/rain/main.go
