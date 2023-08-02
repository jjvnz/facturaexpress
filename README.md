# FACTURAEXPRESS

FACTURAEXPRESS es una API escrita en Go que utiliza el marco GIN y la base de datos PostgreSQL. La API permite a los usuarios crear y gestionar facturas, con campos para información de la empresa, servicios, valor total y detalles del operador. También hay una tabla de usuarios para almacenar información de inicio de sesión y registro.

## Requisitos previos

- Go 1.20
- pgAdmin4
- PostgreSQL

## Instalación

1. Clona este repositorio en tu máquina local.
2. Navega hasta el directorio del proyecto.
3. Ejecuta el comando `go mod tidy` para descargar y verificar todas las dependencias necesarias.

## Configuración

Crea un archivo `.env` en el directorio raíz del proyecto y agrega las siguientes variables de entorno:

```
DB_HOST= # Tu host de base de datos
DB_PORT= # Tu puerto de base de datos
DB_USER= # Tu usuario de base de datos
DB_PASSWORD= # Tu contraseña de base de datos
DB_NAME= # El nombre de tu base de datos
JWT_SECRET_KEY= # Tu clave secreta para firmar tokens JWT
JWT_EXP_TIME= # El tiempo de expiración para los tokens JWT (en segundos)
```

Asegúrate de reemplazar los valores con tus propios valores.

## Estructura del proyecto

La estructura escojida para el proyecto es la siguiente:


```
facturaexpress/
├── data/
│   └── db.go
├── font/
│   └── DejaVuSans.ttf
├── handlers/
│   ├── factura.go
│   ├── login.go
│   ├── registro.go
│   └── roles.go
├── middlewares/
│   └── auth.go
├── models/
│   ├── claims.go
│   ├── error.go
│   ├── factura.go
│   ├── roles.go
│   └── usuario.go
├── routes/
│   └── router.go
├── .env
├── .gitignore
├── go.mod
├── go.sum
├── main.go
└── README.md
```

- La carpeta `data` contiene el archivo `db.go` que interactúa con la base de datos.
- La carpeta `font` contiene el archivo de fuente `DejaVuSans.ttf`.
- La carpeta `handlers` contiene los controladores para las facturas, inicio de sesión, registro y roles.
- La carpeta `middlewares` contiene el middleware de autenticación.
- La carpeta `models` contiene las definiciones de modelos de datos para las reclamaciones, errores, facturas, roles y usuarios.
- La carpeta `routes` contiene el archivo `router.go` que define las rutas de la API.
- Los archivos `.env`, `.gitignore`, `go.mod`, `go.sum`, `main.go` y `README.md` son archivos de configuración y código principal del proyecto.

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

## Ejecución

Para ejecutar la aplicación, navega hasta el directorio del proyecto y ejecuta el comando `go run main.go`. Esto iniciará el servidor en el puerto especificado (por defecto es el puerto 8080).

## Uso

Una vez que el servidor esté en ejecución, puedes utilizar un cliente HTTP como Postman o cURL para enviar solicitudes a la API. Consulta la documentación de la API para obtener más información sobre los puntos finales disponibles y cómo utilizarlos.


## Manejo de roles y permisos

Tambien implementa un sistema de manejo de roles y permisos para controlar el acceso a ciertas funcionalidades de la API. Los usuarios pueden tener diferentes roles, como administrador o usuario regular, y cada rol tiene un conjunto de permisos asociados.

## Autenticación JWT

Se utiliza tokens JWT (JSON Web Tokens) para autenticar a los usuarios y proteger las rutas de la API. Cuando un usuario inicia sesión, se genera un token JWT que contiene información sobre el usuario y se envía al cliente. El cliente debe incluir este token antecedido por el prefijo 'Bearer '  y un espacio en las solicitudes posteriores para acceder a las rutas protegidas.