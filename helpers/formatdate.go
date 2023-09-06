package helpers

import (
	"fmt"
	"time"
)

func FormatDateInSpanish(date_string string) string {
	loc, err := time.LoadLocation("America/Bogota")
	if err != nil {
		fmt.Println(err)
	}
	t, err := time.Parse(time.RFC3339, date_string)
	if err != nil {
		fmt.Println(err)
	}
	t = t.In(loc)
	return fmt.Sprintf("%02d de %s de %d", t.Day(), Meses[t.Month()-1], t.Year())
}

var Meses = [...]string{
	"Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio",
	"Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre",
}
