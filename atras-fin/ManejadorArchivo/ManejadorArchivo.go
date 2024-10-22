package ManejadorArchivo

import (
	"bytes"
	"encoding/binary"
	"fmt"

	//"io"
	"os"
	"proyecto1/Estructura"
	"proyecto1/ManejadorDisco"
	"proyecto1/Usuario"
	"proyecto1/Utilidades"
	"strings"
	"time"
)

// YA REVISADO
func Mkfs(id string, type_ string, buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "MKFS---------------------------------------------------------------------\n")

	var ParticionesMontadas ManejadorDisco.PartitionMounted
	var ParticionEncontrada bool

	for _, Particiones := range ManejadorDisco.GetMountedPartitions() {
		for _, Particion := range Particiones {
			if Particion.ID == id {
				ParticionesMontadas = Particion
				ParticionEncontrada = true
				break
			}
		}
		if ParticionEncontrada {
			break
		}
	}

	if !ParticionEncontrada {
		fmt.Fprintf(buffer, "Error MFKS: La partición: %s no existe.\n", id)
		return
	}

	if ParticionesMontadas.Status != '1' {
		fmt.Fprintf(buffer, "Error MFKS: La partición %s aún no está montada.\n", id)
		return
	}

	archivo, err := Utilidades.OpenFile(ParticionesMontadas.Path)
	if err != nil {
		return
	}

	var MBRTemporal Estructura.MRB
	if err := Utilidades.ReadObject(archivo, &MBRTemporal, 0); err != nil {
		return
	}

	var IndiceParticion int = -1
	for i := 0; i < 4; i++ {
		if MBRTemporal.MRBPartitions[i].PART_Size != 0 {
			if strings.Contains(string(MBRTemporal.MRBPartitions[i].PART_Id[:]), id) {
				IndiceParticion = i
				break
			}
		}
	}

	if IndiceParticion == -1 {
		fmt.Fprintf(buffer, "Error MFKS: La partición: %s no existe.\n", id)
		return
	}

	numerador := int32(MBRTemporal.MRBPartitions[IndiceParticion].PART_Size - int32(binary.Size(Estructura.SuperBlock{})))
	denrominador_base := int32(4 + int32(binary.Size(Estructura.Inode{})) + 3*int32(binary.Size(Estructura.FileBlock{})))
	denrominador := denrominador_base
	n := int32(numerador / denrominador)

	// Crear el Superbloque
	var NuevoSuperBloque Estructura.SuperBlock
	NuevoSuperBloque.S_Filesystem_Type = 2
	NuevoSuperBloque.S_Inodes_Count = n
	NuevoSuperBloque.S_Blocks_Count = 3 * n
	NuevoSuperBloque.S_Free_Blocks_Count = 3*n - 2
	NuevoSuperBloque.S_Free_Inodes_Count = n - 2
	FechaActual := time.Now()
	FechaString := FechaActual.Format("02-01-2006 15:04:05")
	FechaBytes := []byte(FechaString)
	copy(NuevoSuperBloque.S_Mtime[:], FechaBytes)
	copy(NuevoSuperBloque.S_Umtime[:], FechaBytes)
	NuevoSuperBloque.S_Mnt_Count = 1
	NuevoSuperBloque.S_Magic = 0xEF53
	NuevoSuperBloque.S_Inode_Size = int32(binary.Size(Estructura.Inode{}))
	NuevoSuperBloque.S_Block_Size = int32(binary.Size(Estructura.FileBlock{}))
	// Calcular las posiciones de los bloques
	NuevoSuperBloque.S_BM_Inode_Start = MBRTemporal.MRBPartitions[IndiceParticion].PART_Start + int32(binary.Size(Estructura.SuperBlock{}))
	NuevoSuperBloque.S_BM_Block_Start = NuevoSuperBloque.S_BM_Inode_Start + n
	NuevoSuperBloque.S_Inode_Start = NuevoSuperBloque.S_BM_Block_Start + 3*n
	NuevoSuperBloque.S_Block_Start = NuevoSuperBloque.S_Inode_Start + n*int32(binary.Size(Estructura.Inode{}))
	// Escribir el superbloque en el archivo
	SistemaEXT2(n, MBRTemporal.MRBPartitions[IndiceParticion], NuevoSuperBloque, FechaString, archivo, buffer)
	defer archivo.Close()
}

func SistemaEXT2(n int32, Particion Estructura.Partition, NuevoSuperBloque Estructura.SuperBlock, Fecha string, archivo *os.File, buffer *bytes.Buffer) {
	for i := int32(0); i < n; i++ {
		err := Utilidades.WriteObject(archivo, byte(0), int64(NuevoSuperBloque.S_BM_Inode_Start+i))
		if err != nil {
			return
		}
	}
	for i := int32(0); i < 3*n; i++ {
		err := Utilidades.WriteObject(archivo, byte(0), int64(NuevoSuperBloque.S_BM_Block_Start+i))
		if err != nil {
			return
		}
	}
	// Inicializa inodos y bloques con valores predeterminados
	if err := initInodesAndBlocks(n, NuevoSuperBloque, archivo); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	// Crea la carpeta raíz y el archivo users.txt
	if err := createRootAndUsersFile(NuevoSuperBloque, Fecha, archivo); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	// Escribe el superbloque actualizado al archivo
	if err := Utilidades.WriteObject(archivo, NuevoSuperBloque, int64(Particion.PART_Start)); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	// Marca los primeros inodos y bloques como usados
	if err := markUsedInodesAndBlocks(NuevoSuperBloque, archivo); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	// Imprimir el SuperBlock final
	Estructura.PrintSuperBlock(buffer, NuevoSuperBloque)
	fmt.Fprintf(buffer, "Partición: %s formateada exitosamente.\n", string(Particion.PART_Name[:]))

}

// Función auxiliar para inicializar inodos y bloques
func initInodesAndBlocks(n int32, newSuperblock Estructura.SuperBlock, file *os.File) error {
	var newInode Estructura.Inode
	for i := int32(0); i < 15; i++ {
		newInode.I_Block[i] = -1
	}

	for i := int32(0); i < n; i++ {
		if err := Utilidades.WriteObject(file, newInode, int64(newSuperblock.S_Inode_Start+i*int32(binary.Size(Estructura.Inode{})))); err != nil {
			return err
		}
	}

	var newFileblock Estructura.FileBlock
	for i := int32(0); i < 3*n; i++ {
		if err := Utilidades.WriteObject(file, newFileblock, int64(newSuperblock.S_Block_Start+i*int32(binary.Size(Estructura.FileBlock{})))); err != nil {
			return err
		}
	}

	return nil
}

// Función auxiliar para crear la carpeta raíz y el archivo users.txt
func createRootAndUsersFile(newSuperblock Estructura.SuperBlock, date string, file *os.File) error {
	var Inode0, Inode1 Estructura.Inode
	initInode(&Inode0, date)
	initInode(&Inode1, date)

	Inode0.I_Block[0] = 0
	Inode1.I_Block[0] = 1

	// Asignar el tamaño real del contenido
	data := "1,G,root\n1,U,root,root,123\n"
	actualSize := int32(len(data))
	Inode1.I_Size = actualSize // Esto ahora refleja el tamaño real del contenido

	var Fileblock1 Estructura.FileBlock
	copy(Fileblock1.B_Content[:], data) // Copia segura de datos a FileBlock

	var Folderblock0 Estructura.FolderBlock
	Folderblock0.B_Content[0].B_Inodo = 0
	copy(Folderblock0.B_Content[0].B_Name[:], ".")
	Folderblock0.B_Content[1].B_Inodo = 0
	copy(Folderblock0.B_Content[1].B_Name[:], "..")
	Folderblock0.B_Content[2].B_Inodo = 1
	copy(Folderblock0.B_Content[2].B_Name[:], "users.txt")

	// Escribir los inodos y bloques en las posiciones correctas
	if err := Utilidades.WriteObject(file, Inode0, int64(newSuperblock.S_Inode_Start)); err != nil {
		return err
	}
	if err := Utilidades.WriteObject(file, Inode1, int64(newSuperblock.S_Inode_Start+int32(binary.Size(Estructura.Inode{})))); err != nil {
		return err
	}
	if err := Utilidades.WriteObject(file, Folderblock0, int64(newSuperblock.S_Block_Start)); err != nil {
		return err
	}
	if err := Utilidades.WriteObject(file, Fileblock1, int64(newSuperblock.S_Block_Start+int32(binary.Size(Estructura.FolderBlock{})))); err != nil {
		return err
	}

	return nil
}

// Función auxiliar para inicializar un inodo
func initInode(inode *Estructura.Inode, date string) {
	inode.I_Uid = 1
	inode.I_Gid = 1
	inode.I_Size = 0
	copy(inode.I_Atime[:], date)
	copy(inode.I_Ctime[:], date)
	copy(inode.I_Mtime[:], date)
	copy(inode.I_Perm[:], "664")

	for i := int32(0); i < 15; i++ {
		inode.I_Block[i] = -1
	}
}

// Función auxiliar para marcar los inodos y bloques usados
func markUsedInodesAndBlocks(newSuperblock Estructura.SuperBlock, file *os.File) error {
	if err := Utilidades.WriteObject(file, byte(1), int64(newSuperblock.S_BM_Inode_Start)); err != nil {
		return err
	}
	if err := Utilidades.WriteObject(file, byte(1), int64(newSuperblock.S_BM_Inode_Start+1)); err != nil {
		return err
	}
	if err := Utilidades.WriteObject(file, byte(1), int64(newSuperblock.S_BM_Block_Start)); err != nil {
		return err
	}
	if err := Utilidades.WriteObject(file, byte(1), int64(newSuperblock.S_BM_Block_Start+1)); err != nil {
		return err
	}
	return nil
}

func Cat(files []string, buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "CAT---------------------------------------------------------------------\n")
	if Usuario.Dato.GetIDParticion() == "" && Usuario.Dato.GetIDUsuario() == "" {
		fmt.Fprintf(buffer, "Error CAT: No hay un usuario logueado.\n")
		return
	}

	ParticionesMount := ManejadorDisco.GetMountedPartitions()
	var filepath string
	var id string

	for _, partitions := range ParticionesMount {
		for _, partition := range partitions {
			if partition.LoggedIn {
				filepath = partition.Path
				id = partition.ID
				break
			}
		}
	}

	file, err := Utilidades.OpenFile(filepath)
	if err != nil {
		return
	}
	defer file.Close()

	// Read the MBR
	var TempMBR Estructura.MRB
	if err := Utilidades.ReadObject(file, &TempMBR, 0); err != nil {
		return
	}

	var index int = -1
	for i := 0; i < 4; i++ {
		if TempMBR.MRBPartitions[i].PART_Size != 0 && strings.Contains(string(TempMBR.MRBPartitions[i].PART_Id[:]), id) {
			if TempMBR.MRBPartitions[i].PART_Status[0] == '1' {
				index = i
				break
			}
		}
	}

	if index == -1 {
		fmt.Fprintf(buffer, "Error CAT: No se encontró la partición.\n")
		return
	}

	var tempSuperblock Estructura.SuperBlock
	if err := Utilidades.ReadObject(file, &tempSuperblock, int64(TempMBR.MRBPartitions[index].PART_Start)); err != nil {
		return
	}

	for _, filePath := range files {

		indexInode := BuscarStart(filePath, file, tempSuperblock, buffer)
		if indexInode == -1 {
			fmt.Fprintf(buffer, "Error: No se pudo encontrar el archivo %s\n", filePath)
			continue
		}

		var crrInode Estructura.Inode
		if err := Utilidades.ReadObject(file, &crrInode, int64(tempSuperblock.S_Inode_Start+indexInode*int32(binary.Size(Estructura.Inode{})))); err != nil {
			continue
		}
		for _, block := range crrInode.I_Block {
			if block != -1 {
				var fileblock Estructura.FileBlock
				if err := Utilidades.ReadObject(file, &fileblock, int64(tempSuperblock.S_Block_Start+block*int32(binary.Size(Estructura.FileBlock{})))); err != nil {
					continue
				}
				Estructura.PrintFileBlock(buffer, fileblock)
			}
		}
		fmt.Fprintf(buffer, "CAT: Archivo %s Impreso Exitosamente.\n", filePath)
	}
}

func BuscarStart(path string, file *os.File, tempSuperblock Estructura.SuperBlock, buffer *bytes.Buffer) int32 {
	TempStepsPath := strings.Split(path, "/")
	RutaPasada := TempStepsPath[1:]
	var Inode0 Estructura.Inode
	if err := Utilidades.ReadObject(file, &Inode0, int64(tempSuperblock.S_Inode_Start)); err != nil {
		return -1
	}
	return BuscarInodoRuta(RutaPasada, Inode0, file, tempSuperblock, buffer)
}

func BuscarInodoRuta(RutaPasada []string, Inode Estructura.Inode, file *os.File, tempSuperblock Estructura.SuperBlock, buffer *bytes.Buffer) int32 {
	SearchedName := strings.Replace(pop(&RutaPasada), " ", "", -1)
	for _, block := range Inode.I_Block {
		if block != -1 {
			if len(RutaPasada) == 0 {
				var fileblock Estructura.FileBlock
				if err := Utilidades.ReadObject(file, &fileblock, int64(tempSuperblock.S_Block_Start+block*int32(binary.Size(Estructura.FileBlock{})))); err != nil {
					return -1
				}
				return 1
			} else {
				var crrFolderBlock Estructura.FolderBlock
				if err := Utilidades.ReadObject(file, &crrFolderBlock, int64(tempSuperblock.S_Block_Start+block*int32(binary.Size(Estructura.FolderBlock{})))); err != nil {
					return -1
				}
				for _, folder := range crrFolderBlock.B_Content {
					if strings.Contains(string(folder.B_Name[:]), SearchedName) {
						var NextInode Estructura.Inode
						if err := Utilidades.ReadObject(file, &NextInode, int64(tempSuperblock.S_Inode_Start+folder.B_Inodo*int32(binary.Size(Estructura.Inode{})))); err != nil {
							return -1
						}
						return BuscarInodoRuta(RutaPasada, NextInode, file, tempSuperblock, buffer)
					}
				}
			}
		}
	}
	return -1
}

func pop(s *[]string) string {
	lastIndex := len(*s) - 1
	last := (*s)[lastIndex]
	*s = (*s)[:lastIndex]
	return last
}

func tienePermiso(buffer *bytes.Buffer) bool {
	ParticionesMount := ManejadorDisco.GetMountedPartitions()
	var filepath string
	var id string

	for _, partitions := range ParticionesMount {
		for _, partition := range partitions {
			// Verifica si alguna partición tiene un usuario logueado
			if partition.LoggedIn {
				filepath = partition.Path
				id = partition.ID
				break
			}
		}
	}

	file, err := Utilidades.OpenFile(filepath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo:", err)
		return false
	}
	defer file.Close()

	var TempMBR Estructura.MRB

	if err := Utilidades.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		return false
	}

	var index int = -1

	for i := 0; i < 4; i++ {
		if TempMBR.MRBPartitions[i].PART_Size != 0 {
			if strings.Contains(string(TempMBR.MRBPartitions[i].PART_Id[:]), id) {
				if TempMBR.MRBPartitions[i].PART_Status[0] == '1' {
					index = i
				} else {
					return false
				}
				break
			}
		}
	}

	if index == -1 {
		return false
	}

	var tempSuperblock Estructura.SuperBlock
	if err := Utilidades.ReadObject(file, &tempSuperblock, int64(TempMBR.MRBPartitions[index].PART_Start)); err != nil {
		return false
	}

	indexInode := buscarStart("/users.txt", file, tempSuperblock, buffer)

	var crrInode Estructura.Inode

	if err := Utilidades.ReadObject(file, &crrInode, int64(tempSuperblock.S_Inode_Start+indexInode*int32(binary.Size(Estructura.Inode{})))); err != nil {
		return false
	}

	perm := string(crrInode.I_Perm[:])
	return strings.Contains(perm, "664")
}

// Función modificada para buscar y leer Fileblocks en lugar de Folderblocks
func buscarStart(path string, file *os.File, tempSuperblock Estructura.SuperBlock, buffer *bytes.Buffer) int32 {
	TempStepsPath := strings.Split(path, "/")
	RutaPasada := TempStepsPath[1:]

	var Inode0 Estructura.Inode
	if err := Utilidades.ReadObject(file, &Inode0, int64(tempSuperblock.S_Inode_Start)); err != nil {
		return -1
	}

	return BuscarInodoRuta(RutaPasada, Inode0, file, tempSuperblock, buffer)
}
