# Requirements

- sqlite3 `brew install sqlite3`
- air `go install github.com/cosmtrek/air@latest`
- golangci-lint `brew install golangci-lint`
- mailhog ( For testing smtp features )

# Get started

After you have installed the above dependencies,
go into the project directory and simply run `npm --prefix ./frontend run dev` to run the frontend
and `air` to run the backend

## Migrations

To create a migration you should only create a new file with an increasing running number.
For example if goose is installed you can use following command

```sh
goose create my_new_table sql
```

To check project related migrations check following path:

```sh
database/migrations
```

from the root project directory

This will create in your actual folder a new file with a timestamp prefix.
Note: Any file prefix must differ and be in the correct order e.g. 1,2,3... or timestampMMHHss1...
