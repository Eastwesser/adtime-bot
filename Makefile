.PHONY: migrate-up migrate-down migrate-status migrate-create

# Применить все миграции
migrate-up:
	goose -dir internal/storage/migrations postgres "user=postgres password=postgres dbname=adtime sslmode=disable" up

# Откатить последнюю миграцию
migrate-down:
	goose -dir internal/storage/migrations postgres "user=postgres password=postgres dbname=adtime sslmode=disable" down

# Показать статус миграций
migrate-status:
	goose -dir internal/storage/migrations postgres "user=postgres password=postgres dbname=adtime sslmode=disable" status

# Создать новую миграцию
migrate-create:
	@read -p "Enter migration name: " name; \
	goose -dir internal/storage/migrations postgres "user=postgres password=postgres dbname=adtime sslmode=disable" create $${name// /_} sql