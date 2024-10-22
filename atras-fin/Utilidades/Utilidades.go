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
