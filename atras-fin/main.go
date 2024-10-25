package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"proyecto1/Analizador"
	"proyecto1/ManejadorDisco"
)

type ReadMBRParams struct {
	Path string `json:"path"`
}

func HabilitarCors(respuesta *http.ResponseWriter) {
	(*respuesta).Header().Set("Access-Control-Allow-Origin", "*")
	(*respuesta).Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	(*respuesta).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func AnalizarEntrada(respuesta http.ResponseWriter, solicitud *http.Request) {
	HabilitarCors(&respuesta)
	if solicitud.Method == http.MethodPost {
		body, err := ioutil.ReadAll(solicitud.Body)
		if err != nil {
			http.Error(respuesta, "Error En La Solicitud.", http.StatusInternalServerError)
			return
		}
		fmt.Println("--------")
		fmt.Println(string(body))
		fmt.Println("--------")
		result := Analizador.Analizar(string(body))
		fmt.Fprint(respuesta, result)
		return
	}
	http.Error(respuesta, "Método No permitido", http.StatusMethodNotAllowed)
}

func ObtenerParticiones(respuesta http.ResponseWriter, solicitud *http.Request) {
	HabilitarCors(&respuesta)

	// Manejar la solicitud de verificación de CORS
	if solicitud.Method == http.MethodOptions {
		respuesta.WriteHeader(http.StatusOK)
		return
	}

	if solicitud.Method == http.MethodPost {
		var params ReadMBRParams
		// Decodificar el cuerpo JSON de la solicitud
		err := json.NewDecoder(solicitud.Body).Decode(&params)
		if err != nil {
			http.Error(respuesta, "Error al procesar la solicitud", http.StatusBadRequest)
			return
		}

		if params.Path == "" {
			http.Error(respuesta, "La ruta es requerida", http.StatusBadRequest)
			return
		}

		// Leer el MBR y obtener las particiones
		partitions, err := ManejadorDisco.ListPartitions(params.Path)
		if err != nil {
			http.Error(respuesta, fmt.Sprintf("Error al leer las particiones: %v", err), http.StatusInternalServerError)
			return
		}

		// Responder con las particiones en formato JSON
		json.NewEncoder(respuesta).Encode(partitions)
	} else {
		http.Error(respuesta, "Método No permitido", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/AnalizadorGo/ProcesarComando", AnalizarEntrada)
	http.HandleFunc("/AnalizadorGo/ObtenerParticiones", ObtenerParticiones)
	fmt.Println("-------------------------------------------")
	fmt.Println("Servidor corriendo en localhost:8080")
	fmt.Println("-------------------------------------------")
	http.ListenAndServe(":8080", nil)
}
