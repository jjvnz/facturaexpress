# FACTURAEXPRESS

FACTURAEXPRESS es una API escrita en Go que utiliza el marco GIN y la base de datos PostgreSQL. La API permite a los usuarios crear y gestionar facturas, con campos para información de la empresa, servicios, valor total y detalles del operador. También hay una tabla de usuarios para almacenar información de inicio de sesión y registro, incluye la funcionalidad de generar reportes de facturas en formato PDF.

## Requisitos previos

- Go 1.20
- pgAdmin4
- PostgreSQL

## Instalación

1. Clona este repositorio en tu máquina local.
2. Navega hasta el directorio del proyecto.
3. Ejecuta el comando `go mod tidy` para descargar y verificar todas las dependencias necesarias.

## Configuración

Crea un archivo `config.json` en el directorio raíz del proyecto y agrega las siguientes variables:

```json
{
    "db": {
        "host": "Tu host de base de datos",
        "port": "Tu puerto de base de datos",
        "user": "Tu usuario de base de datos",
        "password": "Tu contraseña de base de datos",
        "dbname": "El nombre de tu base de datos"
    },
    "jwt": {
        "secret_key": "Tu clave secreta para firmar tokens JWT",
        "exp_time": "El tiempo de expiración para los tokens JWT (en segundos)"
    }
}
```

Asegúrate de reemplazar los valores con tus propios valores.

## Estructura del proyecto

La estructura escojida para el proyecto es la siguiente:


```
facturaexpress/
├── common/
│       └── constant.go
├── data/
│   └── db.go
├── font/
│   └── DejaVuSans.ttf
├── handlers/
│   ├── auth/
│   │       ├── login.go
│   │       ├── logout.go
│   │       └── registro.go
│   ├── invoice/
│   │       ├── createinvoice.go
│   │       ├── deleteinvoice.go
│   │       ├── generatepdf.go
│   │       ├── getinvoice.go
│   │       ├── listinvoices.go
│   │       └── updateinvoice.go
│   ├── role/
│   │       ├── assignrole.go
│   │       ├── listroles.go
│   │       └── updaterole.go
│   ├── user/
│   │       ├── createuser.go
│   │       ├── deleteuser.go
│   │       ├── getuserinfo.go
│   │       ├── listusers.go
│   │       └── updateuser.go
├── middlewares/
│   └── auth.go
├── models/
│   ├── claim.go
│   ├── db.go
│   ├── error.go
│   ├── invoice.go
│   ├── jwt.go
│   ├── role.go
│   └── user.go
├── routes/
|    └── router.go 
├── helpers/
|    ├── checkroleexists.go 
|    ├── checkusernameemail.go 
|    ├── formatdate.go
|    ├── generatejwttoken.go 
|    ├── getuseridfrominvoice.go 
|    ├── saveuser.go 
|    ├── saveuserrole.go 
|    ├── unmarshalservices.go 
|    ├── verifycredentials.go 
|    ├── verifyrole.go 
|    └── verifytoken.go 
├── interfaces/
|    └── database.go 
├── .gitignore 
├── README.md 
├── config.json 
├── go.mod 
├── go.sum 
└── main.go 
```
- La carpeta `common` contiene el archivo `constant.go` en él se definen constantes requeridas en el proyecto como "ADMIN" y "USER" etc.
- La carpeta `data` contiene el archivo `db.go` que interactúa con la base de datos.
- La carpeta `font` contiene el archivo de fuente `DejaVuSans.ttf`.
- La carpeta `handlers` contiene los controladores para las facturas, inicio de sesión, registro y roles.
- La carpeta `middlewares` contiene el middleware de autenticación.
- La carpeta `models` contiene las definiciones de modelos de datos para las reclamaciones, errores, facturas, roles y usuarios.
- La carpeta `routes` contiene el archivo `router.go` que define las rutas de la API.
- La carpeta `helpers` contiene funciones auxiliares para verificar roles, nombres de usuario y correos electrónicos, generar tokens JWT, guardar usuarios y roles, verificar credenciales y más.
- La carpeta `interfaces` contiene el archivo database. go que define la interfaz para interactuar con la base de datos.
- Los archivos `.gitignore`, `config.json`, `go.mod`, `go.sum`, `main.go` y `README.md` son archivos de configuración y código principal del proyecto.

## Esquema de base de datos

Aquí el esquema de base de datos para una tabla llamada `facturas`:

```sql
CREATE TABLE facturas (
    id SERIAL PRIMARY KEY,
    empresa_nombre TEXT NOT NULL,
    empresa_nit TEXT NOT NULL,
    fecha DATE NOT NULL,
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

## Ejecución

Para ejecutar la aplicación, navega hasta el directorio del proyecto y ejecuta el comando `go run main.go`. Esto iniciará el servidor en el puerto especificado (por defecto es el puerto 8080).

## Uso

Una vez que el servidor esté en ejecución, puedes utilizar un cliente HTTP como Postman o cURL para enviar solicitudes a la API. Consulta la [Documentación de Postman](https://documenter.getpostman.com/view/23764700/2s9Xy5LAAk) de la API para obtener más información sobre los puntos finales disponibles y cómo utilizarlos.


## Manejo de roles y permisos

Tambien implementa un sistema de manejo de roles y permisos para controlar el acceso a ciertas funcionalidades de la API. Los usuarios pueden tener diferentes roles, como administrador o usuario regular, y cada rol tiene un conjunto de permisos asociados.

## Autenticación JWT

Se utiliza tokens JWT (JSON Web Tokens) para autenticar a los usuarios y proteger las rutas de la API. Cuando un usuario inicia sesión, se genera un token JWT que contiene información sobre el usuario y se envía al cliente. El cliente debe incluir este token antecedido por el prefijo 'Bearer '  y un espacio en las solicitudes posteriores para acceder a las rutas protegidas.


## Códigos de error


Los códigos de error se definen en el archivo common/constant.go y se utilizan en todo el proyecto para mejorar la legibilidad y la gestión de los códigos de error. Aquí están las constantes de error que se utilizan actualmente:

```go
const (
	ErrInvalidAuthHeader          = "INVALID_AUTH_HEADER"
	ErrInvalidToken               = "INVALID_TOKEN"
	ErrDBError                    = "DB_ERROR"
	ErrInsuficientRole            = "INSUFFICIENT_ROLE"
	ErrBadRequest                 = "BAD_REQUEST"
	ErrEmailNotFound              = "EMAIL_NOT_FOUND"
	ErrIncorrectPassword          = "INCORRECT_PASSWORD"
	ErrJWTGenerationError         = "JWT_GENERATION_ERROR"
	ErrJWTStorageError            = "JWT_STORAGE_ERROR"
	ErrTokenAlreadyBlacklisted    = "TOKEN_ALREADY_BLACKLISTED"
	ErrPasswordHashingFailed      = "PASSWORD_HASHING_FAILED"
	ErrJSONBindingFailed          = "JSON_BINDING_FAILED"
	ErrNoPermission               = "NO_PERMISSION"
	ErrInvalidData                = "INVALID_DATA"
	ErrServicesMarshalError       = "SERVICES_MARSHAL_ERROR"
	ErrServicesUnMarshalError     = "SERVICES_UNMARSHAL_ERROR"
	ErrInvalidID                  = "INVALID_ID"
	ErrInvoiceNotFound            = "INVOICE_NOT_FOUND"
	ErrNotFound                   = "NOT_FOUND"
	ErrInvalidPageParam           = "INVALID_PAGE_PARAM"
	ErrInvalidLimitParam          = "INVALID_LIMIT_PARAM"
	ErrLimitTooHigh               = "LIMIT_TOO_HIGH"
	ErrMissingFields              = "MISSING_FIELDS"
	ErrUserNotFound               = "USER_NOT_FOUND"
	ErrInvalidUserID              = "INVALID_USER_ID"
	ErrInvalidRoleID              = "INVALID_ROLE_ID"
	ErrQueryPreparationFailed     = "QUERY_PREPARATION_FAILED"
	ErrUserVerificationFailed     = "USER_VERIFICATION_FAILED"
	ErrRoleVerificationFailed     = "ROLE_VERIFICATION_FAILED"
	ErrRoleNotFound               = "ROLE_NOT_FOUND"
	ErrUserRoleVerificationFailed = "USER_ROLE_VERIFICATION_FAILED"
	ErrUserAlreadyHasRole         = "USER_ALREADY_HAS_ROLE"
	ErrUserRoleUpdateFailed       = "USER_ROLE_UPDATE_FAILED"
	ErrJWTTokenRetrievalFailed    = "JWT_TOKEN_RETRIEVAL_FAILED"
	ErrJWTTokenBlacklistingFailed = "JWT_TOKEN_BLACKLISTING_FAILED"
	ErrQueryFailed                = "QUERY_FAILED"
	ErrErrorGeneratingToken       = "ERROR_GENERATING_TOKEN"
	ErrDatabaseSaveFailed         = "DATABASE_SAVE_FAILED"
	ErrRoleIDRetrievalFailed      = "ROLE_ID_RETRIEVAL_FAILED"
)
```
