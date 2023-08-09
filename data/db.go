package data

import (
	"database/sql"
	"encoding/json"
	"facturaexpress/models"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

func NewDB() (*DB, error) {
	// Carga los valores del archivo de configuración
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		return nil, fmt.Errorf("error al cargar el archivo de configuración: %v", err)
	}

	var config models.DBConfig
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return nil, fmt.Errorf("error al leer el archivo de configuración: %v", err)
	}

	host := config.DB.Host
	port := config.DB.Port
	user := config.DB.User
	password := config.DB.Password
	dbname := config.DB.DBName

	// Crea la cadena de conexión a la base de datos
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Abre una conexión a la base de datos
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	// Verifica que la conexión a la base de datos sea exitosa
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &DB{conn: db}, nil
}

func (db *DB) Prepare(query string) (*sql.Stmt, error) {
	stmt, err := db.conn.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("error al preparar la consulta: %v", err)
	}
	return stmt, nil
}

func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar la consulta: %v", err)
	}
	return rows, nil
}

func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	result, err := db.conn.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar la consulta: %v", err)
	}
	return result, nil
}

func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.conn.QueryRow(query, args...)
}

func (db *DB) Close() error {
	return db.conn.Close()
}
