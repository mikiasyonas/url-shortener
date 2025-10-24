include .env

export $(shell sed 's/\=.*//' .env)

.PHONY:db/migrations/diff
db/migrations/diff:
	@echo 'Generating migration files...'
	atlas migrate diff --env gorm

.PHONY:db/migrations/apply
db/migrations/apply:
	@echo 'Applying migrations...'
	atlas migrate apply --env gorm

start:
	go run cmd/api/main.go