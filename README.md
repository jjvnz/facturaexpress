# FACTURAEXPRESS

## Instalación del marco GIN

Para instalar el marco GIN en su proyecto Go, siga estos pasos:

1. Cree un nuevo directorio para su proyecto y navegue hasta él:
```
mkdir facturaexpress && cd facturaexpress
```

2. Inicialice un nuevo módulo Go:
```
go mod init facturaexpress
```

3. Instale el paquete `gin-gonic/gin` usando el comando `go get`:
```
go get -u github.com/gin-gonic/gin
```

## Estructura del proyecto

La estructura escojida para el proyecto es la siguiente:


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

- La carpeta `api` contiene todo el código relacionado con la API, incluidos los controladores en la subcarpeta `handlers`, el middleware en la subcarpeta `middleware` y el enrutador en el archivo `router.go`.
- La carpeta `cmd` contiene subcarpetas para cada comando ejecutable, y cada subcarpeta contiene un archivo `main.go` que define el comando. En este caso, solo hay un comando llamado `server` que inicia el servidor de la API.
- La carpeta `pkg` contiene paquetes reutilizables que pueden ser importados por comandos en la carpeta `cmd`. En este ejemplo, hay dos paquetes: `models`, que contiene definiciones de modelos de datos, y `storage`, que contiene código para interactuar con la base de datos.
- Los archivos `go.mod` y `go.sum` definen las dependencias del proyecto.

## Esquema de base de datos

Aquí el esquema de base de datos para una tabla llamada `facturas`:

```sql
CREATE TABLE facturas (
    id SERIAL PRIMARY KEY,
    empresa_nombre TEXT NOT NULL,
    empresa_nit TEXT NOT NULL,
    fecha TIMESTAMP NOT NULL,
    servicios JSONB NOT NULL,
    valor_total NUMERIC NOT NULL,
    operador_nombre TEXT NOT NULL,
    operador_tipo_documento TEXT NOT NULL,
    operador_documento TEXT NOT NULL,
    operador_ciudad_expedicion_documento TEXT NOT NULL,
    operador_celular TEXT NOT NULL,
    operador_numero_cuenta_bancaria TEXT NOT NULL,
    operador_tipo_cuenta_bancaria TEXT NOT NULL,
    operador_banco TEXT NOT NULL,
    usuario_id INTEGER NOT NULL
);
```

Este esquema incluye columnas para los campos en las estructuras `Factura`, `Empresa` y `Operador`. El campo `Servicios` se almacena como una columna JSONB, lo que le permite almacenar una matriz de objetos `Servicio` en una sola fila. Se ha agregado la columna `usuario_id` para almacenar el ID del usuario asociado con cada factura.

También puede crear una nueva tabla llamada `usuarios` para almacenar datos de usuario. Aquí el esquema para la tabla `usuarios`:

```sql
CREATE TABLE usuarios (
    id SERIAL PRIMARY KEY,
    nombre_usuario TEXT NOT NULL,
    password TEXT NOT NULL,
    correo TEXT NOT NULL
);
```

Este esquema incluye columnas para los campos en la estructura `Usuario`.
