# FACTURAEXPRESS

## Installation of the GIN framework

To install the GIN framework in your Go project, follow these steps:

1. Create a new directory for your project and navigate to it:
```
mkdir facturaexpress && cd facturaexpress
```

2. Initialize a new Go module:
```
go mod init facturaexpress
```

3. Install the `gin-gonic/gin` package using the `go get` command:
```
go get -u github.com/gin-gonic/gin
```

## Project Structure

The recommended structure for a Go API project using the GIN framework is as follows:


```
facturaexpress/
|── api/
|    |── handlers/
|    |    |── factura.go
|    |    |── login.go
|    |    |── registro.go
|    |    └── welcome.go
|    └── middleware/
|    |    └── auth.go
|    └── router.go
├── cmd/
|    └── server/
|        ├── .env
|        └── main.go
├── font/
|    └── DejaVuSans.ttf
├── pkg/
|    ├── models/
|    |    ├── claims.go
|    |    ├── factura.go
|    |    └── usuario.go
|    └── storage/
|        └── db.go
├── .gitignore
├── README.md
├── go.mod
└── go.sum
```

- The `api` folder contains all code related to the API, including controllers in the `handlers` subfolder, middleware in the `middleware` subfolder, and the router in the `router.go` file.
- The `cmd` folder contains subfolders for each executable command, and each subfolder contains a `main.go` file that defines the command. In this case, there is only one command called `server` that starts the API server.
- The `pkg` folder contains reusable packages that can be imported by commands in the `cmd` folder. In this example, there are two packages: `models`, which contains data model definitions, and `storage`, which contains code for interacting with the database.
- The `go.mod` and `go.sum` files define the project's dependencies.

## Database Schema

Here is an example of a database schema for a table named `facturas`:

```sql
CREATE TABLE facturas (
    id SERIAL PRIMARY KEY,
    nombre_operador TEXT NOT NULL,
    nit_operador TEXT NOT NULL,
    fecha TIMESTAMP NOT NULL,
    servicios JSONB NOT NULL,
    valor_total NUMERIC NOT NULL
);
```
