# Usar la imagen oficial de Golang
FROM golang:1.20

# Establecer el directorio de trabajo dentro del contenedor
WORKDIR /app

# Copiar los archivos necesarios al contenedor
COPY . .

# Descargar todas las dependencias
RUN go get -d -v ./...

# Construir la aplicación
RUN go build -o main .

# Exponer el puerto en el que se ejecutará la aplicación
EXPOSE 8000

# Comando para ejecutar la aplicación
CMD ["./main"]
