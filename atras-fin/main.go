package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"proyecto1/Analizador"
)

func analizarTexto(respuesta http.ResponseWriter, solicitud *http.Request) {
	habilitarCors(&respuesta)
	if solicitud.Method == http.MethodPost {
		body, err := ioutil.ReadAll(solicitud.Body)
		if err != nil {
			http.Error(respuesta, "Error al leer el cuerpo de la solicitud", http.StatusInternalServerError)
			return
		}

		// Obtén el resultado del análisis
		result := Analizador.Analizar(string(body))

		// Envía el resultado al frontend
		fmt.Fprint(respuesta, result)
		return
	}
	http.Error(respuesta, "Método No permitido", http.StatusMethodNotAllowed)
}

func habilitarCors(respuesta *http.ResponseWriter) {
	(*respuesta).Header().Set("Access-Control-Allow-Origin", "*")
	(*respuesta).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func main() {
	http.HandleFunc("/analizar", analizarTexto)
	fmt.Println("Servidor escuchando en el puerto 8080...")
	http.ListenAndServe(":8080", nil)
}

/*
import (
	"fmt"
	"proyecto1/Analizador"
)
	func main() {
		fmt.Println("===Start===")
		Analizador.Analizar("mkdisk -size=3000 -unit=K -path=/home/pablo03r/discosp1/prueba1.mia ")
		fmt.Println("===End===")
		}

*/
