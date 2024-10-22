package Utilidades

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// ================================= Crear Archivo =================================
func CreateFile(name string) error {
	// Asignar directorio
	dir := filepath.Dir(name)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		fmt.Println("Error al crear Archivo, ya existe uno ==", err)
		return err
	}
	// Crear archivo
	if _, err := os.Stat(name); os.IsNotExist(err) {
		file, err := os.Create(name)
		if err != nil {
			fmt.Println("Error al crear Archivo ==", err)
			return err
		}
		defer file.Close()
	}
	return nil
}

// ================================= Abrir Archivo =================================
func OpenFile(name string) (*os.File, error) {
	file, err := os.OpenFile(name, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Error Abrir Archivo ==", err)
		return nil, err
	}
	return file, nil
}

// ================================= Eliminar Archivo =================================
func DeleteFile(nombre string) error {
	if _, err := os.Stat(nombre); os.IsNotExist(err) {
		println("Error: El archivo no existe.")
		fmt.Println("Error: El archivo no existe.")
		return err
	}
	err := os.Remove(nombre)
	if err != nil {
		println("Error al eliminar el archivo: ", err)
		fmt.Println("Error al eliminar el archivo: ", err)
		return err
	}

	return nil
}

// ================================= Escribir Objeto =================================
func WriteObject(file *os.File, data interface{}, position int64) error {
	file.Seek(position, 0)
	err := binary.Write(file, binary.LittleEndian, data)
	if err != nil {
		fmt.Println("Error Escriir Objeto ==", err)
		return err
	}
	return nil
}

// ================================= Leer Objeto =================================
func ReadObject(file *os.File, data interface{}, position int64) error {
	file.Seek(position, 0)
	err := binary.Read(file, binary.LittleEndian, data)
	if err != nil {
		fmt.Println("Error al leer objeto ==", err)
		return err
	}
	return nil
}

// Función para llenar el espacio con ceros (\0)
func FillWithZeros(file *os.File, start int32, size int32) error {
	// Posiciona el archivo al inicio del área que debe ser llenada
	file.Seek(int64(start), 0)

	// Crear un buffer lleno de ceros
	buffer := make([]byte, size)

	// Escribir los ceros en el archivo
	_, err := file.Write(buffer)
	if err != nil {
		fmt.Println("Error al llenar el espacio con ceros:", err)
		return err
	}

	fmt.Println("Espacio llenado con ceros desde el byte", start, "por", size, "bytes.")
	return nil
}

// Función para verificar que un bloque del archivo esté lleno de ceros
func VerifyZeros(file *os.File, start int32, size int32) {
	zeros := make([]byte, size)
	_, err := file.ReadAt(zeros, int64(start))
	if err != nil {
		fmt.Println("Error al leer la sección eliminada:", err)
		return
	}

	// Verificar si todos los bytes leídos son ceros
	isZeroFilled := true
	for _, b := range zeros {
		if b != 0 {
			isZeroFilled = false
			break
		}
	}

	if isZeroFilled {
		fmt.Println("La partición eliminada está completamente llena de ceros.")
	} else {
		fmt.Println("Advertencia: La partición eliminada no está completamente llena de ceros.")
	}
}
