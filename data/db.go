package data

import (
	"database/sql"
	"encoding/json"
	interfaceDB "facturaexpress/interfaces"
	"facturaexpress/models"
	"fmt"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

type PostgresAdapter struct {
	conn *sql.DB
}

// implemeto la interfaz Database
var _ interfaceDB.Database = &PostgresAdapter{}

var instance *PostgresAdapter
var once sync.Once

func GetInstance() *PostgresAdapter {
	once.Do(func() {
		configFile, err := os.ReadFile("config.json")
		if err != nil {
			panic(fmt.Errorf("error al cargar el archivo de configuración: %v", err))
		}

		var config models.DBConfig
		err = json.Unmarshal(configFile, &config)
		if err != nil {
			panic(fmt.Errorf("error al leer el archivo de configuración: %v", err))
		}

		host := config.DB.Host
		port := config.DB.Port
		user := config.DB.User
		password := config.DB.Password
		dbname := config.DB.DBName

		psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)

		db, err := sql.Open("postgres", psqlInfo)
		if err != nil {
			panic(err)
		}

		err = db.Ping()
		if err != nil {
			panic(err)
		}

		instance = &PostgresAdapter{conn: db}
	})
	return instance
}

func (db *PostgresAdapter) Prepare(query string) (*sql.Stmt, error) {
	stmt, err := db.conn.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("error al preparar la consulta: %v", err)
	}
	return stmt, nil
}

func (db *PostgresAdapter) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar la consulta: %v", err)
	}
	return rows, nil
}

func (db *PostgresAdapter) Exec(query string, args ...interface{}) (sql.Result, error) {
	result, err := db.conn.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar la consulta: %v", err)
	}
	return result, nil
}

func (db *PostgresAdapter) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.conn.QueryRow(query, args...)
}

func (db *PostgresAdapter) Close() error {
	return db.conn.Close()
}
