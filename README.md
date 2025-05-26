# ADTIME TELEGRAM BOT (GOLANG)

To work with this bot use makefile.

## 1. Create a migration:
```bash
make create-migration
# enter migration name
```

## 2. Run the project:
```bash
make docker-up
```

## 3. Run migrations:
```bash
make migrate
```

## 4. Stop the project:
```bash
make docker-down
```

## TESTING

With docker-compose:
```bash
docker-compose run --rm migrator
```

Or just with makefiles:
```bash
make migrate-up
```

Create:
```bash
make migrate-create
# Type a migration name: "add_texture_description"
```

Check migration status:
```bash
make migrate-status
```

Revert the last migration:
```bash
make migrate-down
```
