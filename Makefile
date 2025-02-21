goose-static-create:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-static-create <type_of_creation>_<name_of_migration> (eg. create_table_my_table)"; \
	else \
		goose -s --dir database/migrations/static create $(word 2, $(MAKECMDGOALS)) sql; \
	fi
goose-static-up:
	goose --dir database/migrations/static sqlite3 database/database.sqlite up
goose-dynamic-up:
	goose -no-versioning --dir database/migrations/dynamic sqlite3 database/database.sqlite up
goose-static-up-1:
	goose --dir database/migrations/static sqlite3 database/database.sqlite up-by-one
goose-dynamic-up-1:
	goose -no-versioning --dir database/migrations/dynamic sqlite3 database/database.sqlite up-by-one
goose-static-down-1:
	goose --dir database/migrations/static sqlite3 database/database.sqlite down
goose-dynamic-down-1:
	goose -no-versioning --dir database/migrations/dynamic sqlite3 database/database.sqlite down
goose-static-down-to-0:
	goose --dir database/migrations/static sqlite3 database/database.sqlite down-to 0
goose-dynamic-down-to-0:
	goose -no-versioning --dir database/migrations/dynamic sqlite3 database/database.sqlite down-to 0
goose-static-down-to:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-down-to <version_number>"; \
	else \
		goose --dir database/migrations/static sqlite3 database/database.sqlite down-to $(word 2, $(MAKECMDGOALS)); \
	fi
goose-dynamic-down-to:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-down-to <version_number>"; \
	else \
		goose -no-versioning --dir database/migrations/dynamic sqlite3 database/database.sqlite down-to $(word 2, $(MAKECMDGOALS)); \
	fi
goose-static-down-to-and-up:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-down-to-and-up <version_number>"; \
	else \
		goose --dir database/migrations/static sqlite3 database/database.sqlite down-to $(word 2, $(MAKECMDGOALS)); \
		goose --dir database/migrations/static sqlite3 database/database.sqlite up; \
	fi
goose-dynamic-down-to-and-up:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-down-to-and-up <version_number>"; \
	else \
		goose -no-versioning --dir database/migrations/dynamic sqlite3 database/database.sqlite down-to $(word 2, $(MAKECMDGOALS)); \
		goose -no-versioning --dir database/migrations/dynamic sqlite3 database/database.sqlite up; \
	fi
%:
	@: