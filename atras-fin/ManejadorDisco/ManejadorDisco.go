package ManejadorDisco

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"proyecto1/Estructura"
	"proyecto1/Utilidades"
	"strings"
	"time"
)

// Estructura para representar una partición montada
type PartitionMounted struct {
	Path     string
	Name     string
	ID       string
	Status   byte // 0: no montada, 1: montada
	LoggedIn bool // true: usuario ha iniciado sesión, false: no ha iniciado sesión
}

// Mapa para almacenar las particiones montadas, organizadas por disco
var mountedPartitions = make(map[string][]PartitionMounted)

type RutaDisco struct {
	Ruta string
}

var ListaRutasDiscos []RutaDisco

func AgregarRutaDisco(ruta string) {
	ListaRutasDiscos = append(ListaRutasDiscos, RutaDisco{Ruta: ruta})
}

func EliminarRutaDisco(ruta_disco string) {
	for i, ruta := range ListaRutasDiscos {
		if ruta.Ruta == ruta_disco {
			ListaRutasDiscos = append(ListaRutasDiscos[:i], ListaRutasDiscos[i+1:]...)
			break
		}
	}
}

func ObtenerRutaDiscos(buffer *bytes.Buffer) {
	for _, ruta := range ListaRutasDiscos {
		fmt.Fprintf(buffer, "%s\n", ruta.Ruta)
	}
}

// Función para imprimir las particiones montadas
func PrintMountedPartitions(buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "Particiones montadas: \n")

	if len(mountedPartitions) == 0 {
		fmt.Fprintf(buffer, "No hay particiones montadas. \n")
		return
	}

	for diskID, partitions := range mountedPartitions {
		fmt.Printf("Disco ID: %s\n", diskID)
		for _, partition := range partitions {
			loginStatus := "No"
			if partition.LoggedIn {
				loginStatus = "Sí"
			}
			fmt.Fprintf(buffer, " - Partición Name: %s, ID: %s, Path: %s, Status: %c, LoggedIn: %s\n",
				partition.Name, partition.ID, partition.Path, partition.Status, loginStatus)
		}
	}
	fmt.Fprintln(buffer, "")
}

// Función para eliminar particiones
func DeletePartition(path string, name string, delete_ string, buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "======Start DELETE PARTITION======")
	fmt.Fprintf(buffer, "Path:"+path)
	fmt.Fprintf(buffer, "Name:"+name)
	fmt.Fprintf(buffer, "Delete type:"+delete_)

	// Abrir el archivo binario en la ruta proporcionada
	file, err := Utilidades.OpenFile(path)
	if err != nil {
		fmt.Fprintf(buffer, "Error: Could not open file at path:"+path)
		return
	}

	var TempMBR Estructura.MRB
	// Leer el objeto desde el archivo binario
	if err := Utilidades.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Fprintf(buffer, "Error: Could not read MBR from file")
		return
	}

	// Buscar la partición por nombre
	found := false
	for i := 0; i < 4; i++ {
		// Limpiar los caracteres nulos al final del nombre de la partición
		partitionName := strings.TrimRight(string(TempMBR.MRBPartitions[i].PART_Name[:]), "\x00")
		if partitionName == name {
			found = true

			// Si es una partición extendida, eliminar las particiones lógicas dentro de ella
			if TempMBR.MRBPartitions[i].PART_Type[0] == 'e' {
				fmt.Fprintf(buffer, "Eliminando particiones lógicas dentro de la partición extendida...")
				ebrPos := TempMBR.MRBPartitions[i].PART_Start
				var ebr Estructura.EBR
				for {
					err := Utilidades.ReadObject(file, &ebr, int64(ebrPos))
					if err != nil {
						fmt.Print(buffer, "Error al leer EBR:", err)
						break
					}
					// Detener el bucle si el EBR está vacío
					if ebr.EBRStart == 0 && ebr.EBRSize == 0 {
						fmt.Fprintf(buffer, "EBR vacío encontrado, deteniendo la búsqueda.")
						break
					}
					// Depuración: Mostrar el EBR leído
					fmt.Fprintf(buffer, "EBR leído antes de eliminar:")
					Estructura.PrintEBR(buffer, ebr)

					// Eliminar partición lógica
					if delete_ == "fast" {
						ebr = Estructura.EBR{}                           // Resetear el EBR manualmente
						Utilidades.WriteObject(file, ebr, int64(ebrPos)) // Sobrescribir el EBR reseteado
					} else if delete_ == "full" {
						Utilidades.FillWithZeros(file, ebr.EBRStart, ebr.EBRSize)
						ebr = Estructura.EBR{}                           // Resetear el EBR manualmente
						Utilidades.WriteObject(file, ebr, int64(ebrPos)) // Sobrescribir el EBR reseteado
					}

					// Depuración: Mostrar el EBR después de eliminar
					fmt.Fprintf(buffer, "EBR después de eliminar:")
					Estructura.PrintEBR(buffer, ebr)

					if ebr.EBRNext == -1 {
						break
					}
					ebrPos = ebr.EBRNext
				}
			}

			// Proceder a eliminar la partición (extendida, primaria o lógica)
			if delete_ == "fast" {
				// Eliminar rápido: Resetear manualmente los campos de la partición
				TempMBR.MRBPartitions[i] = Estructura.Partition{} // Resetear la partición manualmente
				fmt.Fprintf(buffer, "Partición eliminada en modo Fast.")
			} else if delete_ == "full" {
				// Eliminar completamente: Resetear manualmente y sobrescribir con '\0'
				start := TempMBR.MRBPartitions[i].PART_Start
				size := TempMBR.MRBPartitions[i].PART_Size
				TempMBR.MRBPartitions[i] = Estructura.Partition{} // Resetear la partición manualmente
				// Escribir '\0' en el espacio de la partición en el disco
				Utilidades.FillWithZeros(file, start, size)
				fmt.Fprintf(buffer, "Partición eliminada en modo Full.")

				// Leer y verificar si el área está llena de ceros
				Utilidades.VerifyZeros(file, start, size)
			}
			break
		}
	}

	if !found {
		// Buscar particiones lógicas si no se encontró en el MBR
		fmt.Fprintf(buffer, "Buscando en particiones lógicas dentro de las extendidas...")
		for i := 0; i < 4; i++ {
			if TempMBR.MRBPartitions[i].PART_Type[0] == 'e' { // Solo buscar dentro de particiones extendidas
				ebrPos := TempMBR.MRBPartitions[i].PART_Start
				var ebr Estructura.EBR
				for {
					err := Utilidades.ReadObject(file, &ebr, int64(ebrPos))
					if err != nil {
						fmt.Print(buffer, "Error al leer EBR:", err)
						break
					}

					// Depuración: Mostrar el EBR leído
					fmt.Fprintf(buffer, "EBR leído:")
					Estructura.PrintEBR(buffer, ebr)

					logicalName := strings.TrimRight(string(ebr.EBRName[:]), "\x00")
					if logicalName == name {
						found = true
						// Eliminar la partición lógica
						if delete_ == "fast" {
							ebr = Estructura.EBR{}                           // Resetear el EBR manualmente
							Utilidades.WriteObject(file, ebr, int64(ebrPos)) // Sobrescribir el EBR reseteado
							fmt.Fprintf(buffer, "Partición lógica eliminada en modo Fast.")
						} else if delete_ == "full" {
							Utilidades.FillWithZeros(file, ebr.EBRStart, ebr.EBRSize)
							ebr = Estructura.EBR{}                           // Resetear el EBR manualmente
							Utilidades.WriteObject(file, ebr, int64(ebrPos)) // Sobrescribir el EBR reseteado
							Utilidades.VerifyZeros(file, ebr.EBRStart, ebr.EBRSize)
							fmt.Fprintf(buffer, "Partición lógica eliminada en modo Full.")
						}
						break
					}

					if ebr.EBRNext == -1 {
						break
					}
					ebrPos = ebr.EBRNext
				}
			}
			if found {
				break
			}
		}
	}

	if !found {
		fmt.Print(buffer, "Error: No se encontró la partición con el nombre:", name)
		return
	}

	// Sobrescribir el MBR
	if err := Utilidades.WriteObject(file, TempMBR, 0); err != nil {
		fmt.Fprintf(buffer, "Error: Could not write MBR to file")
		return
	}

	// Leer el MBR actualizado y mostrarlo
	fmt.Fprintf(buffer, "MBR actualizado después de la eliminación:")
	Estructura.PrintMBR(buffer, TempMBR)

	// Si es una partición extendida, mostrar los EBRs actualizados
	for i := 0; i < 4; i++ {
		if TempMBR.MRBPartitions[i].PART_Type[0] == 'e' {
			fmt.Fprintf(buffer, "Imprimiendo EBRs actualizados en la partición extendida:")
			ebrPos := TempMBR.MRBPartitions[i].PART_Start
			var ebr Estructura.EBR
			for {
				err := Utilidades.ReadObject(file, &ebr, int64(ebrPos))
				if err != nil {
					fmt.Print(buffer, "Error al leer EBR:", err)
					break
				}
				// Detener el bucle si el EBR está vacío
				if ebr.EBRStart == 0 && ebr.EBRSize == 0 {
					fmt.Fprintf(buffer, "EBR vacío encontrado, deteniendo la búsqueda.")
					break
				}
				// Depuración: Imprimir cada EBR leído
				fmt.Fprintf(buffer, "EBR leído después de actualización:")
				Estructura.PrintEBR(buffer, ebr)
				if ebr.EBRNext == -1 {
					break
				}
				ebrPos = ebr.EBRNext
			}
		}
	}

	// Cerrar el archivo binario
	defer file.Close()

	fmt.Fprintf(buffer, "======FIN DELETE PARTITION======")
}

func ModifyPartition(path string, name string, add int, unit string, buffer *bytes.Buffer) error {
	fmt.Fprintf(buffer, "======Start MODIFY PARTITION======")
	if add == 0 {
		fmt.Fprintf(buffer, "Error FDISK ADD: El tamaño a agregar debe ser distinto que 0.\n")
		return nil
	}
	// Abrir el archivo binario en la ruta proporcionada
	file, err := Utilidades.OpenFile(path)
	if err != nil {
		fmt.Fprintf(buffer, "Error: Could not open file at path:"+path)
		return err
	}
	defer file.Close()

	// Leer el MBR
	var TempMBR Estructura.MRB
	if err := Utilidades.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Fprintf(buffer, "Error: Could not read MBR from file")
		return err
	}

	// Imprimir MBR antes de modificar
	fmt.Fprintf(buffer, "MBR antes de la modificación:")
	Estructura.PrintMBR(buffer, TempMBR)

	// Buscar la partición por nombre
	var foundPartition *Estructura.Partition
	var partitionType byte

	// Revisar si la partición es primaria o extendida
	for i := 0; i < 4; i++ {
		partitionName := strings.TrimRight(string(TempMBR.MRBPartitions[i].PART_Name[:]), "\x00")
		if partitionName == name {
			foundPartition = &TempMBR.MRBPartitions[i]
			partitionType = TempMBR.MRBPartitions[i].PART_Type[0]
			break
		}
	}

	// Si no se encuentra en las primarias/extendidas, buscar en las particiones lógicas
	if foundPartition == nil {
		for i := 0; i < 4; i++ {
			if TempMBR.MRBPartitions[i].PART_Type[0] == 'e' {
				ebrPos := TempMBR.MRBPartitions[i].PART_Start
				var ebr Estructura.EBR
				for {
					if err := Utilidades.ReadObject(file, &ebr, int64(ebrPos)); err != nil {
						fmt.Fprintf(buffer, "Error al leer EBR: %v", err)
						return err
					}

					ebrName := strings.TrimRight(string(ebr.EBRName[:]), "\x00")
					if ebrName == name {
						partitionType = 'l' // Partición lógica
						foundPartition = &Estructura.Partition{
							PART_Start: ebr.EBRStart,
							PART_Size:  ebr.EBRSize,
						}
						break
					}

					// Continuar buscando el siguiente EBR
					if ebr.EBRNext == -1 {
						break
					}
					ebrPos = ebr.EBRNext
				}
				if foundPartition != nil {
					break
				}
			}
		}
	}

	// Verificar si la partición fue encontrada
	if foundPartition == nil {
		fmt.Fprintf(buffer, "Error: No se encontró la partición con el nombre:"+name)
		return nil // Salir si no se encuentra la partición
	}

	// Convertir unidades a bytes
	var addBytes int
	if unit == "k" {
		addBytes = add * 1024
	} else if unit == "m" {
		addBytes = add * 1024 * 1024
	} else {
		fmt.Fprintf(buffer, "Error: Unidad desconocida, debe ser 'k' o 'm'")
		return nil // Salir si la unidad no es válida
	}

	// Flag para saber si continuar o no
	var shouldModify = true

	// Comprobar si es posible agregar o quitar espacio
	if add > 0 {
		// Agregar espacio: verificar si hay suficiente espacio libre después de la partición
		nextPartitionStart := foundPartition.PART_Start + foundPartition.PART_Size
		if partitionType == 'l' {
			// Para particiones lógicas, verificar con el siguiente EBR o el final de la partición extendida
			for i := 0; i < 4; i++ {
				if TempMBR.MRBPartitions[i].PART_Type[0] == 'e' {
					extendedPartitionEnd := TempMBR.MRBPartitions[i].PART_Start + TempMBR.MRBPartitions[i].PART_Size
					if nextPartitionStart+int32(addBytes) > extendedPartitionEnd {
						fmt.Fprintf(buffer, "Error: No hay suficiente espacio libre dentro de la partición extendida")
						shouldModify = false
					}
					break
				}
			}
		} else {
			// Para primarias o extendidas
			if nextPartitionStart+int32(addBytes) > TempMBR.MRBSize {
				fmt.Fprintf(buffer, "Error: No hay suficiente espacio libre después de la partición")
				shouldModify = false
			}
		}
	} else {
		// Quitar espacio: verificar que no se reduzca el tamaño por debajo de 0
		if foundPartition.PART_Size+int32(addBytes) < 0 {
			fmt.Fprintf(buffer, "Error: No es posible reducir la partición por debajo de 0")
			shouldModify = false
		}
	}

	// Solo modificar si no hay errores
	if shouldModify {
		foundPartition.PART_Size += int32(addBytes)
	} else {
		fmt.Fprintf(buffer, "No se realizaron modificaciones debido a un error.")
		return nil // Salir si hubo un error
	}

	// Si es una partición lógica, sobrescribir el EBR
	if partitionType == 'l' {
		ebrPos := foundPartition.PART_Start
		var ebr Estructura.EBR
		if err := Utilidades.ReadObject(file, &ebr, int64(ebrPos)); err != nil {
			fmt.Fprintf(buffer, "Error al leer EBR: %v", err)
			return err
		}

		// Actualizar el tamaño en el EBR y escribirlo de nuevo
		ebr.EBRSize = foundPartition.PART_Size
		if err := Utilidades.WriteObject(file, ebr, int64(ebrPos)); err != nil {
			fmt.Fprintf(buffer, "Error al escribir el EBR actualizado: %v", err)
			return err
		}

		// Imprimir el EBR modificado
		fmt.Fprintf(buffer, "EBR modificado:")
		Estructura.PrintEBR(buffer, ebr)
	}

	// Sobrescribir el MBR actualizado
	if err := Utilidades.WriteObject(file, TempMBR, 0); err != nil {
		fmt.Println(buffer, "Error al escribir el MBR actualizado: %v", err)
		return err
	}

	// Imprimir el MBR modificado
	fmt.Fprintf(buffer, "MBR después de la modificación:")
	Estructura.PrintMBR(buffer, TempMBR)

	fmt.Fprintf(buffer, "======END MODIFY PARTITION======")
	return nil
}

// Estructura para representar una partición en JSON
type PartitionInfo struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Start  int32  `json:"start"`
	Size   int32  `json:"size"`
	Status string `json:"status"`
}

// Función para leer el MBR desde un archivo binario y devolver las particiones
func ListPartitions(path string) ([]PartitionInfo, error) {
	// Abrir el archivo binario
	file, err := Utilidades.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close()

	// Crear una variable para almacenar el MBR
	var mbr Estructura.MRB

	// Leer el MBR desde el archivo
	err = Utilidades.ReadObject(file, &mbr, 0) // Leer desde la posición 0
	if err != nil {
		return nil, fmt.Errorf("error al leer el MBR: %v", err)
	}

	// Crear una lista de particiones basada en el MBR
	var partitions []PartitionInfo
	for _, partition := range mbr.MRBPartitions {
		if partition.PART_Size > 0 { // Solo agregar si la partición tiene un tamaño
			// Limpiar el nombre para eliminar caracteres nulos
			partitionName := strings.TrimRight(string(partition.PART_Name[:]), "\x00")

			partitions = append(partitions, PartitionInfo{
				Name:   partitionName,
				Type:   strings.TrimRight(string(partition.PART_Type[:]), "\x00"),
				Start:  partition.PART_Start,
				Size:   partition.PART_Size,
				Status: strings.TrimRight(string(partition.PART_Status[:]), "\x00"),
			})
		}
	}

	return partitions, nil
}

func Unmount(id string, buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "Desmontando partición con ID:"+id)

	// Buscar la partición montada por ID
	var partitionFound *PartitionMounted
	var diskID string
	var partitionIndex int

	for disk, partitions := range mountedPartitions {
		for i, partition := range partitions {
			if partition.ID == id {
				partitionFound = &partitions[i]
				diskID = disk
				partitionIndex = i
				break
			}
		}
		if partitionFound != nil {
			break
		}
	}

	// Si no se encuentra la partición, mostrar un error
	if partitionFound == nil {
		fmt.Fprintf(buffer, "Error: No se encontró una partición montada con el ID proporcionado:"+id)
		return
	}

	// Abrir el archivo del disco correspondiente
	file, err := Utilidades.OpenFile(partitionFound.Path)
	if err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo abrir el archivo en la ruta:"+partitionFound.Path)
		return
	}
	defer file.Close()

	// Leer el MBR
	var TempMBR Estructura.MRB
	if err := Utilidades.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo leer el MBR desde el archivo")
		return
	}

	// Buscar la partición en el MBR utilizando el nombre
	nameBytes := [16]byte{}
	copy(nameBytes[:], []byte(partitionFound.Name))
	partitionUpdated := false

	for i := 0; i < 4; i++ {
		if bytes.Equal(TempMBR.MRBPartitions[i].PART_Name[:], nameBytes[:]) {
			// Cambiar el estado de la partición de montada ('1') a desmontada ('0')
			TempMBR.MRBPartitions[i].PART_Status[0] = '0'
			// Borrar el ID de la partición
			copy(TempMBR.MRBPartitions[i].PART_Id[:], "")
			partitionUpdated = true
			break
		}
	}

	if !partitionUpdated {
		fmt.Fprintf(buffer, "Error: No se pudo encontrar la partición en el MBR para desmontar")
		return
	}

	// Sobrescribir el MBR actualizado al archivo
	if err := Utilidades.WriteObject(file, TempMBR, 0); err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo sobrescribir el MBR en el archivo")
		return
	}

	// Eliminar la partición de la lista de particiones montadas
	mountedPartitions[diskID] = append(mountedPartitions[diskID][:partitionIndex], mountedPartitions[diskID][partitionIndex+1:]...)

	// Si ya no hay particiones montadas en este disco, eliminar el disco de la lista
	if len(mountedPartitions[diskID]) == 0 {
		delete(mountedPartitions, diskID)
	}

	fmt.Fprintf(buffer, "Partición desmontada con éxito.")
	PrintMountedPartitions(buffer) // Mostrar las particiones montadas restantes
}

// Funcion para eliminar una particion montada
func EliminarDiscoPorRuta(ruta string, buffer *bytes.Buffer) {
	discoID := generateDiskID(ruta)
	if _, existe := mountedPartitions[discoID]; existe {
		delete(mountedPartitions, discoID)
		fmt.Fprintf(buffer, "El disco con ruta '%s' y sus particiones asociadas han sido eliminados.\n", ruta)
	}
}

// Función para obtener las particiones montadas
func GetMountedPartitions() map[string][]PartitionMounted {
	return mountedPartitions
}

// Función para marcar una partición como logueada
func MarkPartitionAsLoggedIn(id string) {
	for diskID, partitions := range mountedPartitions {
		for i, partition := range partitions {
			if partition.ID == id {
				mountedPartitions[diskID][i].LoggedIn = true
				fmt.Println("Partición con ID marcada como logueada.")
				return
			}
		}
	}
	fmt.Printf("No se encontró la partición con ID %s para marcarla como logueada.\n", id)
}

func MarkPartitionAsLoggedOut(id string) {
	for DiscoID, partitions := range mountedPartitions {
		for i, Particion := range partitions {
			if Particion.ID == id {
				mountedPartitions[DiscoID][i].LoggedIn = false
				return
			}
		}
	}
}

// Función para obtener el ID del último disco montado
func getLastDiskID() string {
	var lastDiskID string
	for diskID := range mountedPartitions {
		lastDiskID = diskID
	}
	return lastDiskID
}

func generateDiskID(path string) string {
	return strings.ToLower(path)
}

// YA REVISADO
func Mkdisk(size int, fit string, unit string, path string, buffer *bytes.Buffer) {

	fmt.Fprintln(buffer, "=-=-=-=-=-=-=INICIO MKDISK=-=-=-=-=-=-=")
	println("Size:", size)
	println("Fit:", fit)
	println("Unit:", unit)
	println("Path:", path)
	// Validar fit bf/ff/wf

	// Validar fit bf/ff/wf
	if fit != "bf" && fit != "wf" && fit != "ff" {
		fmt.Fprintf(buffer, "Error: Fit debe ser bf, wf o ff")
		return
	}

	// Validar size > 0
	if size <= 0 {
		fmt.Fprintf(buffer, "Error: Size debe ser mayor a 0")
		return
	}

	// Validar unit k - m
	if unit != "k" && unit != "m" {
		fmt.Fprintf(buffer, "Error: Las unidades válidas son k o m")
		return
	}

	// Crear el archivo
	err := Utilidades.CreateFile(path)
	if err != nil {
		fmt.Print(buffer, "Error:", err)
		return
	}

	// Asignar tamaño en bytes
	if unit == "k" {
		size = size * 1024
	} else {
		size = size * 1024 * 1024
	}

	// Abrir el archivo binario
	file, err := Utilidades.OpenFile(path)
	if err != nil {
		return
	}

	// Optimización: Escribir grandes bloques de ceros
	blockSize := 1024 * 1024             // Bloques de 1MB
	zeroBlock := make([]byte, blockSize) // Crear un bloque de ceros

	remainingSize := size

	for remainingSize > 0 {
		if remainingSize < blockSize {
			// Escribe lo que queda si es menor que el tamaño del bloque
			zeroBlock = make([]byte, remainingSize)
		}
		_, err := file.Write(zeroBlock)
		if err != nil {
			fmt.Print(buffer, "Error escribiendo ceros:", err)
			return
		}
		remainingSize -= blockSize
	}

	// Crear el MBR
	var newMRB Estructura.MRB
	newMRB.MRBSize = int32(size)
	newMRB.MRBSignature = rand.Int31() // Número aleatorio rand.Int31() genera solo números no negativos
	copy(newMRB.MRBFit[:], fit)

	// Obtener la fecha actual en formato YYYY-MM-DD
	currentTime := time.Now()
	formattedDate := currentTime.Format("2006-01-02")
	copy(newMRB.MRBCreationDate[:], formattedDate)

	// Escribir el MBR en el archivo
	if err := Utilidades.WriteObject(file, newMRB, 0); err != nil {
		return
	}

	// Leer el archivo y verificar el MBR
	var TempMBR Estructura.MRB
	if err := Utilidades.ReadObject(file, &TempMBR, 0); err != nil {
		return
	}

	// Imprimir el MBR
	Estructura.PrintMBR(buffer, TempMBR)

	// Cerrar el archivo
	defer file.Close()
	// // ================================= VALIDACIONES =================================
	// if size <= 0 {
	// 	fmt.Fprintln(buffer, "Error: El tamaño debe ser mayor que 0.")
	// 	return
	// }

	// if fit != "bf" && fit != "wf" && fit != "ff" {
	// 	fmt.Fprintln(buffer, "Error: El fit debe ser BF, WF, o FF.")
	// 	return
	// }

	// if unit != "k" && unit != "m" {
	// 	fmt.Fprintln(buffer, "Error: La unit debe ser K o M.")
	// 	return
	// }

	// if path == "" {
	// 	fmt.Fprintln(buffer, "Error: La path es obligatoria.")
	// 	return
	// }

	// err := Utilidades.CreateFile(path)

	// if err != nil {
	// 	fmt.Fprintln(buffer, "Error: ", err)
	// 	return
	// }

	// if unit == "k" {
	// 	size = size * 1024
	// } else {
	// 	size = size * 1024 * 1024
	// }

	// // ================================= ABRIR ARCHIVO =================================
	// archivo, err := Utilidades.OpenFile(path)

	// if err != nil {
	// 	fmt.Fprintln(buffer, "Error: ", err)
	// 	return
	// }

	// // ================================= inicializar el archivo con 0
	// for i := 0; i < size; i++ {
	// 	err := Utilidades.WriteObject(archivo, byte(0), int64(i))
	// 	if err != nil {
	// 		fmt.Fprintln(buffer, "Error: ", err)
	// 		return
	// 	}
	// }

	// // ================================= Inicializar el MBR
	// var nuevo_mbr Estructura.MRB
	// nuevo_mbr.MRBSize = int32(size)
	// nuevo_mbr.MRBSignature = rand.Int31()
	// currentTime := time.Now()
	// fechaFormateada := currentTime.Format("02-01-2006")
	// copy(nuevo_mbr.MRBCreationDate[:], fechaFormateada)
	// copy(nuevo_mbr.MRBFit[:], fit)

	// // ================================= Escribir el MBR en el archivo
	// if err := Utilidades.WriteObject(archivo, nuevo_mbr, 0); err != nil {
	// 	fmt.Fprintln(buffer, "Error: ", err)
	// 	return
	// }
	// defer archivo.Close()
	// fmt.Fprintln(buffer, "Disco creado con éxito en la path: ", path)
	// println("Disco creado con éxito en la path: ", path)
	AgregarRutaDisco(path)
	fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN MKDISK=-=-=-=-=-=-=")
}

// YA REVISADO
func Rmdisk(path string, buffer *bytes.Buffer) {
	fmt.Fprintln(buffer, "=-=-=-=-=-=-=INICIO RMDISK=-=-=-=-=-=-=")

	// ================================= Validar la path (path)
	if path == "" {
		fmt.Fprintln(buffer, "Error RMDISK: La path es obligatoria.")
		return
	}

	// ================================= Eliminar el archivo en la path especificada
	err := Utilidades.DeleteFile(path)
	if err != nil {
		fmt.Fprintln(buffer, "Error RMDISK:", err)
		fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN RMDISK=-=-=-=-=-=-=")
		return
	}

	// ================================= Eliminar las particiones montadas asociadas al disco
	EliminarDiscoPorRuta(path, buffer)
	EliminarRutaDisco(path)
	//fmt.Fprintln(buffer, "Disco eliminado con éxito en la path:", path)
	fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN RMDISK=-=-=-=-=-=-=")
}

// YA REVISADO
func Fdisk(size int, path string, name string, unit string, type_ string, fit string, buffer *bytes.Buffer) {
	fmt.Fprintln(buffer, "=-=-=-=-=-=-=INICIO FDISK=-=-=-=-=-=-=")
	// Validar el tamaño (size)
	if size <= 0 {
		fmt.Fprintln(buffer, "Error FDISK: EL tamaño de la partición debe ser mayor que 0.")
		fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN FDISK=-=-=-=-=-=-=")
		return
	}
	// Validar la unit (unit)
	if unit != "b" && unit != "k" && unit != "m" {
		fmt.Fprintln(buffer, "Error FDISK: La unit de tamaño debe ser Bytes, Kilobytes, Megabytes.")
		fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN FDISK=-=-=-=-=-=-=")
		return
	}
	// Validar la path (path)
	if path == "" {
		fmt.Fprintln(buffer, "Error FDISK: La path del disco es obligatoria.")
		fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN FDISK=-=-=-=-=-=-=")
		return
	}
	// Validar el type_ (type_)
	if type_ != "p" && type_ != "e" && type_ != "l" {
		fmt.Fprintln(buffer, "Error FDISK: El type de partición debe ser Primaria, Extendida, Lógica.")
		fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN FDISK=-=-=-=-=-=-=")
		return
	}
	// Validar el fit (fit)
	if fit != "bf" && fit != "ff" && fit != "wf" {
		fmt.Fprintln(buffer, "Error FDISK: El fit de la partición debe ser BF, WF o FF.")
		fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN FDISK=-=-=-=-=-=-=")
		return
	}
	// Validar el name (name)
	if name == "" {
		fmt.Fprintln(buffer, "Error FDISK: El name de la partición es obligatorio.")
		fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN FDISK=-=-=-=-=-=-=")
		return
	}

	// Convertir el tamaño a bytes
	if unit == "k" {
		size = size * 1024
	} else if unit == "m" {
		size = size * 1024 * 1024
	} else if unit == "b" {
		size = size * 1
	}

	// Abrir archivo binario
	archivo, err := Utilidades.OpenFile(path)
	if err != nil {
		return
	}

	var MBRTemporal Estructura.MRB
	if err := Utilidades.ReadObject(archivo, &MBRTemporal, 0); err != nil {
		return
	}

	for i := 0; i < 4; i++ {
		if strings.Contains(string(MBRTemporal.MRBPartitions[i].PART_Name[:]), name) {
			fmt.Fprintf(buffer, "Error FDISK: El name: %s ya está en uso en las particiones.\n", name)
			fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN FDISK=-=-=-=-=-=-=")
			return
		}
	}

	var ContadorPrimaria, ContadorExtendida, TotalParticiones int
	var EspacioUtilizado int32 = 0

	for i := 0; i < 4; i++ {
		if MBRTemporal.MRBPartitions[i].PART_Size != 0 {
			TotalParticiones++
			EspacioUtilizado += MBRTemporal.MRBPartitions[i].PART_Size

			if MBRTemporal.MRBPartitions[i].PART_Type[0] == 'p' {
				ContadorPrimaria++
			} else if MBRTemporal.MRBPartitions[i].PART_Type[0] == 'e' {
				ContadorExtendida++
			}
		}
	}

	if TotalParticiones >= 4 && type_ != "l" {
		fmt.Fprintln(buffer, "Error FDISK: No se pueden crear más de 4 particiones primarias o extendidas en total.")
		fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN FDISK=-=-=-=-=-=-=")
		return
	}
	if type_ == "e" && ContadorExtendida > 0 {
		fmt.Fprintln(buffer, "Error FDISK: Solo se permite una partición extendida por disco.")
		fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN FDISK=-=-=-=-=-=-=")
		return
	}
	if type_ == "l" && ContadorExtendida == 0 {
		fmt.Fprintln(buffer, "Error FDISK: No se puede crear una partición lógica sin una partición extendida.")
		fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN FDISK=-=-=-=-=-=-=")
		return
	}
	if EspacioUtilizado+int32(size) > MBRTemporal.MRBSize {
		fmt.Fprintln(buffer, "Error FDISK: No hay suficiente espacio en el disco para crear esta partición.")
		fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN FDISK=-=-=-=-=-=-=")
		return
	}

	var vacio int32 = int32(binary.Size(MBRTemporal))
	if TotalParticiones > 0 {
		vacio = MBRTemporal.MRBPartitions[TotalParticiones-1].PART_Start + MBRTemporal.MRBPartitions[TotalParticiones-1].PART_Size
	}

	for i := 0; i < 4; i++ {
		if MBRTemporal.MRBPartitions[i].PART_Size == 0 {
			if type_ == "p" || type_ == "e" {
				MBRTemporal.MRBPartitions[i].PART_Size = int32(size)
				MBRTemporal.MRBPartitions[i].PART_Start = vacio
				copy(MBRTemporal.MRBPartitions[i].PART_Name[:], name)
				copy(MBRTemporal.MRBPartitions[i].PART_Fit[:], fit)
				copy(MBRTemporal.MRBPartitions[i].PART_Status[:], "0")
				copy(MBRTemporal.MRBPartitions[i].PART_Type[:], type_)
				MBRTemporal.MRBPartitions[i].PART_Correlative = int32(TotalParticiones + 1)
				if type_ == "e" {
					EBRInicio := vacio
					EBRNuevo := Estructura.EBR{
						EBRFit:   [1]byte{fit[0]},
						EBRStart: EBRInicio,
						EBRSize:  0,
						EBRNext:  -1,
					}
					copy(EBRNuevo.EBRName[:], "")
					if err := Utilidades.WriteObject(archivo, EBRNuevo, int64(EBRInicio)); err != nil {
						return
					}
				}
				fmt.Fprintf(buffer, "Partición creada exitosamente en la path: %s con el name: %s.", path, name)
				break
			}
		}
	}

	if type_ == "l" {
		var ParticionExtendida *Estructura.Partition
		for i := 0; i < 4; i++ {
			if MBRTemporal.MRBPartitions[i].PART_Type[0] == 'e' {
				ParticionExtendida = &MBRTemporal.MRBPartitions[i]
				break
			}
		}
		if ParticionExtendida == nil {
			fmt.Fprintln(buffer, "Error FDISK: No se encontró una partición extendida para crear la partición lógica.")
			return
		}

		EBRPosterior := ParticionExtendida.PART_Start
		var EBRUltimo Estructura.EBR
		for {
			if err := Utilidades.ReadObject(archivo, &EBRUltimo, int64(EBRPosterior)); err != nil {
				return
			}
			if strings.Contains(string(EBRUltimo.EBRName[:]), name) {
				fmt.Fprintf(buffer, "Error FDISK: El name: %s ya está en uso en las particiones.", name)
				return
			}
			if EBRUltimo.EBRNext == -1 {
				break
			}
			EBRPosterior = EBRUltimo.EBRNext
		}

		var EBRNuevoPosterior int32
		if EBRUltimo.EBRSize == 0 {
			EBRNuevoPosterior = EBRPosterior
		} else {
			EBRNuevoPosterior = EBRUltimo.EBRStart + EBRUltimo.EBRSize
		}

		if EBRNuevoPosterior+int32(size)+int32(binary.Size(Estructura.EBR{})) > ParticionExtendida.PART_Start+ParticionExtendida.PART_Size {
			fmt.Fprintln(buffer, "Error FDISK: No hay suficiente espacio en la partición extendida para esta partición lógica.")
			return
		}

		if EBRUltimo.EBRSize != 0 {
			EBRUltimo.EBRNext = EBRNuevoPosterior
			if err := Utilidades.WriteObject(archivo, EBRUltimo, int64(EBRPosterior)); err != nil {
				return
			}
		}

		newEBR := Estructura.EBR{
			EBRFit:   [1]byte{fit[0]},
			EBRStart: EBRNuevoPosterior + int32(binary.Size(Estructura.EBR{})),
			EBRSize:  int32(size),
			EBRNext:  -1,
		}
		copy(newEBR.EBRName[:], name)
		if err := Utilidades.WriteObject(archivo, newEBR, int64(EBRNuevoPosterior)); err != nil {
			return
		}
		fmt.Fprintf(buffer, "Partición lógica creada exitosamente en la path: %s con el name: %s.", path, name)
		fmt.Fprintf(buffer, "---------------------------------------------")
		EBRActual := ParticionExtendida.PART_Start
		for {
			var EBRTemp Estructura.EBR
			if err := Utilidades.ReadObject(archivo, &EBRTemp, int64(EBRActual)); err != nil {
				fmt.Fprintf(buffer, "Error leyendo EBR: %v", err)
				return
			}
			Estructura.PrintEBR(buffer, EBRTemp)
			if EBRTemp.EBRNext == -1 {
				break
			}
			EBRActual = EBRTemp.EBRNext
		}
		fmt.Fprintf(buffer, "---------------------------------------------")
	}
	if err := Utilidades.WriteObject(archivo, MBRTemporal, 0); err != nil {
		return
	}
	var TempMRB Estructura.MRB
	if err := Utilidades.ReadObject(archivo, &TempMRB, 0); err != nil {
		return
	}
	fmt.Fprintf(buffer, "---------------------------------------------")
	Estructura.PrintMBRnormal(TempMRB)
	Estructura.PrintMBR(buffer, TempMRB)
	fmt.Fprintf(buffer, "---------------------------------------------")
	defer archivo.Close()

	fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN FDISK=-=-=-=-=-=-=")
	fmt.Fprintf(buffer, "")
}

// YA REVISADO
func Mount(path string, name string, buffer *bytes.Buffer) {
	fmt.Fprintln(buffer, "=-=-=-=-=-=-=INICIO MOUNT=-=-=-=-=-=-=")
	file, err := Utilidades.OpenFile(path)
	if err != nil {
		fmt.Fprintln(buffer, "Error: No se pudo abrir el archivo en la path:", path)
		fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN MOUNT=-=-=-=-=-=-=")
		return
	}
	defer file.Close()

	var TempMBR Estructura.MRB
	if err := Utilidades.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Fprint(buffer, "Error: No se pudo leer el MBR desde el archivo")
		fmt.Fprintln(buffer, "")
		fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN MOUNT=-=-=-=-=-=-=")
		return
	}

	fmt.Fprintf(buffer, "Buscando partición con name: '%s'", name)
	fmt.Fprintln(buffer, "")

	partitionFound := false
	var partition Estructura.Partition
	var partitionIndex int

	// Convertir el name a comparar a un arreglo de bytes de longitud fija
	nameBytes := [16]byte{}
	copy(nameBytes[:], []byte(name))

	for i := 0; i < 4; i++ {
		if TempMBR.MRBPartitions[i].PART_Type[0] == 'p' && bytes.Equal(TempMBR.MRBPartitions[i].PART_Name[:], nameBytes[:]) {
			partition = TempMBR.MRBPartitions[i]
			partitionIndex = i
			partitionFound = true
			break
		}
	}

	if !partitionFound {
		fmt.Fprintln(buffer, "Error: Partición no encontrada o no es una partición primaria")
		fmt.Fprintln(buffer, "")
		fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN MOUNT=-=-=-=-=-=-=")
		return
	}

	// Verificar si la partición ya está montada
	if partition.PART_Status[0] == '1' {
		fmt.Fprintf(buffer, "Error: La partición ya está montada")
		fmt.Fprintln(buffer, "")
		fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN MOUNT=-=-=-=-=-=-=")
		return
	}

	//fmt.Fprint("Partición encontrada: '%s' en posición %d\n", string(partition.Name[:]), partitionIndex+1)

	// Generar el ID de la partición
	diskID := generateDiskID(path)

	// Verificar si ya se ha montado alguna partición de este disco
	mountedPartitionsInDisk := mountedPartitions[diskID]
	var letter byte

	if len(mountedPartitionsInDisk) == 0 {
		// Es un nuevo disco, asignar la siguiente letra disponible
		if len(mountedPartitions) == 0 {
			letter = 'a'
		} else {
			lastDiskID := getLastDiskID()
			lastLetter := mountedPartitions[lastDiskID][0].ID[len(mountedPartitions[lastDiskID][0].ID)-1]
			letter = lastLetter + 1
		}
	} else {
		// Utilizar la misma letra que las otras particiones montadas en el mismo disco
		letter = mountedPartitionsInDisk[0].ID[len(mountedPartitionsInDisk[0].ID)-1]
	}

	// Incrementar el número para esta partición
	carnet := "202201947" // Cambiar su carnet aquí
	lastTwoDigits := carnet[len(carnet)-2:]
	indice := len(mountedPartitionsInDisk)
	partitionID := fmt.Sprintf("%s%d%c", lastTwoDigits, indice+1, letter)

	// Actualizar el estado de la partición a montada y asignar el ID
	partition.PART_Status[0] = '1'
	copy(partition.PART_Id[:], partitionID)
	TempMBR.MRBPartitions[partitionIndex] = partition
	mountedPartitions[diskID] = append(mountedPartitions[diskID], PartitionMounted{
		Path:   path,
		Name:   name,
		ID:     partitionID,
		Status: '1',
	})

	// Escribir el MBR actualizado al archivo
	if err := Utilidades.WriteObject(file, TempMBR, 0); err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo sobrescribir el MBR en el archivo")
		fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN MOUNT=-=-=-=-=-=-=")
		return
	}

	fmt.Fprintf(buffer, "Partición montada con ID: %s", partitionID)

	fmt.Fprintf(buffer, "")
	// Imprimir el MBR actualizado
	fmt.Fprintf(buffer, "MBR actualizado:")
	Estructura.PrintMBRnormal(TempMBR)
	Estructura.PrintMBR(buffer, TempMBR)
	fmt.Fprintf(buffer, "")

	// Imprimir las particiones montadas (solo estan mientras dure la sesion de la consola)

	fmt.Fprintf(buffer, "REVISION DE PARTICIONES MONTADAS")
	fmt.Fprintf(buffer, "")
	fmt.Fprintf(buffer, "")
	//PrintMountedPartitions()
	fmt.Fprintf(buffer, "")
	fmt.Fprintf(buffer, "")
	fmt.Fprintf(buffer, "FIN DE REVISION DE PARTICIONES MONTADAS")

	fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN MOUNT=-=-=-=-=-=-=")
}

func Ldisk(buffer *bytes.Buffer) {
	fmt.Fprintln(buffer, "=-=-=-=-=-=-=INICIO LDISK=-=-=-=-=-=-=")
	PrintMountedPartitions(buffer)
	fmt.Fprintln(buffer, "=-=-=-=-=-=-=FIN LDISK=-=-=-=-=-=-=")
}
