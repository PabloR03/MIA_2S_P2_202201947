package ManejoRoot

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"proyecto1/Estructura"
	systemadministation "proyecto1/ManejadorArchivo"
	"proyecto1/ManejadorDisco"
	"proyecto1/Usuario"
	"proyecto1/Utilidades"
	"strconv"
	"strings"
)

func Mkusr(user string, pass string, grp string, buffer *bytes.Buffer) {
	if Usuario.Dato.GetIDUsuario() == "" && Usuario.Dato.GetIDParticion() == "" {
		fmt.Fprintf(buffer, "Error: No hay un usuario logueado\n ")
		return
	}
	if Usuario.Dato.GetIDUsuario() != "root" {
		fmt.Fprintf(buffer, "Error: El usuario no tiene permiso de escritura\n ")
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
		if filepath != "" {
			break
		}
	}

	file, err := Utilidades.OpenFile(filepath)
	if err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo abrir el archivo:%v", err)
		return
	}
	defer file.Close()

	var TempMBR Estructura.MRB
	if err := Utilidades.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo leer el MBR:%v", err)
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
		fmt.Fprintf(buffer, "Error: No se encontró la partición\n ")
		return
	}

	var tempSuperblock Estructura.SuperBlock
	if err := Utilidades.ReadObject(file, &tempSuperblock, int64(TempMBR.MRBPartitions[index].PART_Start)); err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo leer el superblock:%v", err)
		return
	}

	indexInode := systemadministation.BuscarStart("/users.txt", file, tempSuperblock, buffer)
	if indexInode == -1 {
		fmt.Fprintf(buffer, "Error: No se encontró el archivo /users.txt\n ")
		return
	}

	var crrInode Estructura.Inode
	if err := Utilidades.ReadObject(file, &crrInode, int64(tempSuperblock.S_Inode_Start+indexInode*int32(binary.Size(Estructura.Inode{})))); err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo leer el inodo del archivo /users.txt\n ")
		return
	}

	data := readAllFileBlocks(&crrInode, file, tempSuperblock, buffer)
	cleanedData := LimpiarNull(data)

	// Mejorar la validación del usuario existente
	if userExists(cleanedData, user) {
		fmt.Fprintf(buffer, "Error: El usuario ya existe\n ")
		return
	}

	if !grupExiste(cleanedData, grp, buffer) {
		fmt.Fprintf(buffer, "Error: El grupo no existe\n ")
		return
	}

	lastGroupID := getLastGroupID(cleanedData, buffer) + 1
	newUserData := fmt.Sprintf("%d,U,%s,%s,%s\n", lastGroupID, grp, user, pass)

	if err := writeNewUserData(&crrInode, cleanedData, newUserData, file, tempSuperblock); err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo actualizar el archivo /users.txt:%v", err)
		return
	}

	fmt.Fprintf(buffer, "Usuario creado con éxito:%v \n", user)
}

// Nueva función para verificar si el usuario existe
func userExists(data string, user string) bool {
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) >= 4 && fields[1] == "U" && fields[3] == user {
			return true
		}
	}
	return false
}

func grupExiste(data string, grupo string, buffer *bytes.Buffer) bool {
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		fields := strings.Split(line, ",")
		fmt.Fprintf(buffer, "fields: %v", fields)
		if len(fields) >= 2 && fields[1] == "G" && fields[2] == grupo {
			return true
		}
	}
	return false
}

func LimpiarNull(data string) string {
	return strings.TrimRight(data, "\x00")
}

func readAllFileBlocks(inode *Estructura.Inode, file *os.File, superblock Estructura.SuperBlock, buffer *bytes.Buffer) string {
	var data string
	for _, block := range inode.I_Block {
		if block != -1 {
			var fileBlock Estructura.FileBlock
			if err := Utilidades.ReadObject(file, &fileBlock, int64(superblock.S_Block_Start+block*int32(binary.Size(Estructura.FileBlock{})))); err != nil {
				fmt.Fprintf(buffer, "Error: No se pudo leer el FileBlock:%v", err)
				continue
			}
			data += string(fileBlock.B_Content[:])
		}
	}
	return data
}

func getLastGroupID(data string, buffer *bytes.Buffer) int {
	lines := strings.Split(data, "\n")
	valor := 0
	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) >= 4 && fields[1] == "U" {
			v, err := strconv.Atoi(fields[0])
			if err == nil {
				valor = v
			} else {
				fmt.Fprintf(buffer, "Error al convertir a entero:%v", err)
			}
		}
	}
	fmt.Fprintf(buffer, "Ultimo valor: %v", valor)
	return valor
}

func writeNewUserData(inode *Estructura.Inode, existingData, newUserData string, file *os.File, superblock Estructura.SuperBlock) error {
	fullData := existingData + newUserData
	var currentBlock int32 = 0
	var currentOffset int = 0

	for currentOffset < len(fullData) {
		if currentBlock >= int32(len(inode.I_Block)) {
			return fmt.Errorf("no hay suficientes bloques disponibles\n ")
		}

		if inode.I_Block[currentBlock] == -1 {
			newBlockIndex, err := createNewFileBlock(inode, &superblock, file)
			if err != nil {
				return err
			}
			inode.I_Block[currentBlock] = newBlockIndex
		}

		var fileBlock Estructura.FileBlock
		start := currentOffset
		end := currentOffset + 64
		if end > len(fullData) {
			end = len(fullData)
		}
		copy(fileBlock.B_Content[:], fullData[start:end])

		if err := Utilidades.WriteObject(file, fileBlock, int64(superblock.S_Block_Start+inode.I_Block[currentBlock]*int32(binary.Size(Estructura.FileBlock{})))); err != nil {
			return fmt.Errorf("error al escribir el bloque actualizado: %v", err)
		}

		currentOffset = end
		currentBlock++
	}

	inode.I_Size = int32(len(fullData))
	if err := Utilidades.WriteObject(file, *inode, int64(superblock.S_Inode_Start+inode.I_Block[0]*int32(binary.Size(Estructura.Inode{})))); err != nil {
		return fmt.Errorf("error al actualizar el inodo: %v", err)
	}

	return nil
}

func Mkgrp(grupos string, buffer *bytes.Buffer) {
	if Usuario.Dato.GetIDUsuario() == "" && Usuario.Dato.GetIDParticion() == "" {
		fmt.Fprintf(buffer, "Error: No hay un usuario logueado\n ")
		return
	}
	if Usuario.Dato.GetIDUsuario() != "root" {
		fmt.Fprintf(buffer, "Error: El usuario no tiene permiso de escritura\n ")
		return
	}

	ParticionesMount := ManejadorDisco.GetMountedPartitions()
	var filepath string
	var id string

	for _, particiones := range ParticionesMount {
		for _, particion := range particiones {
			if particion.LoggedIn {
				filepath = particion.Path
				id = particion.ID
				break
			}
		}
		if filepath != "" {
			break
		}
	}

	file, err := Utilidades.OpenFile(filepath)
	if err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo abrir el archivo:%v", err)
		return
	}
	defer file.Close()

	var TempMBR Estructura.MRB
	if err := Utilidades.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo leer el MBR:%v", err)
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
		fmt.Fprintf(buffer, "Error: No se encontró la partición\n ")
		return
	}

	var tempSuperblock Estructura.SuperBlock
	if err := Utilidades.ReadObject(file, &tempSuperblock, int64(TempMBR.MRBPartitions[index].PART_Start)); err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo leer el SuperBlock: %v", err)
		return
	}

	indexInode := systemadministation.BuscarStart("/user.txt", file, tempSuperblock, buffer)
	if indexInode == -1 {
		fmt.Fprintf(buffer, "Error: No se encontró el archivo usuarios.txt\n ")
		return
	}

	var crrInode Estructura.Inode
	if err := Utilidades.ReadObject(file, &crrInode, int64(tempSuperblock.S_Inode_Start+indexInode*int32(binary.Size(Estructura.Inode{})))); err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo leer el Inode del archivo usuarios.txt\n ")
		return
	}

	newGroupID, err := getNewGroupID(file, &tempSuperblock, &crrInode, grupos, buffer)
	if err != nil {
		fmt.Fprintf(buffer, "Error:%v", err)
		return
	}

	newGroupEntry := fmt.Sprintf("%d,G,%s\n", newGroupID, grupos)

	if err := writeNewGroupEntry(file, &tempSuperblock, &crrInode, newGroupEntry); err != nil {
		fmt.Fprintf(buffer, "Error:%v", err)
		return
	}

	// Guardar cambios del inodo
	if err := Utilidades.WriteObject(file, crrInode, int64(tempSuperblock.S_Inode_Start+indexInode*int32(binary.Size(Estructura.Inode{})))); err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo escribir el inodo actualizado\n ")
		return
	}

	fmt.Fprintf(buffer, "Grupo creado exitosamente\n "+grupos)
}

func getNewGroupID(file *os.File, superblock *Estructura.SuperBlock, inode *Estructura.Inode, grupo string, buffer *bytes.Buffer) (int, error) {
	lastID := 0
	for _, block := range inode.I_Block {
		if block != -1 {
			var fileBlock Estructura.FileBlock
			if err := Utilidades.ReadObject(file, &fileBlock, int64(superblock.S_Block_Start+block*int32(binary.Size(Estructura.FileBlock{})))); err != nil {
				fmt.Fprintf(buffer, "Error: No se pudo leer el FileBlock\n ")
				continue
			}
			content := strings.TrimRight(string(fileBlock.B_Content[:]), "\x00")
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.Contains(line, "G") {
					parts := strings.Split(line, ",")
					if len(parts) > 1 {
						existingGroup := strings.TrimSpace(parts[2])
						if existingGroup == grupo {
							return 0, fmt.Errorf("el grupo '%s' ya existe", grupo)
						}
					}
					if len(parts) > 0 {
						id, err := strconv.Atoi(strings.TrimSpace(parts[0]))
						if err == nil && id > lastID {
							lastID = id
						}
					}
				}
			}
		}
	}
	return lastID + 1, nil
}

func writeNewGroupEntry(file *os.File, superblock *Estructura.SuperBlock, inode *Estructura.Inode, newEntry string) error {
	for i, block := range inode.I_Block {
		if block != -1 {
			var fileBlock Estructura.FileBlock
			if err := Utilidades.ReadObject(file, &fileBlock, int64(superblock.S_Block_Start+block*int32(binary.Size(Estructura.FileBlock{})))); err != nil {
				return fmt.Errorf("no se pudo leer el FileBlock\n ")
			}

			content := strings.TrimRight(string(fileBlock.B_Content[:]), "\x00")
			remainingSpace := 64 - len(content)

			if len(newEntry) <= remainingSpace {
				// El nuevo grupo cabe en este bloque
				copy(fileBlock.B_Content[len(content):], []byte(newEntry))
				return Utilidades.WriteObject(file, fileBlock, int64(superblock.S_Block_Start+block*int32(binary.Size(Estructura.FileBlock{}))))
			}
		} else {
			// Encontramos un bloque vacío, vamos a crear uno nuevo
			newBlockIndex, err := createNewFileBlock(inode, superblock, file)
			if err != nil {
				return err
			}
			inode.I_Block[i] = newBlockIndex

			var newFileBlock Estructura.FileBlock
			copy(newFileBlock.B_Content[:], []byte(newEntry))
			return Utilidades.WriteObject(file, newFileBlock, int64(superblock.S_Block_Start+newBlockIndex*int32(binary.Size(Estructura.FileBlock{}))))
		}
	}

	return fmt.Errorf("no hay espacio disponible para crear el nuevo grupo\n ")
}

func createNewFileBlock(_ *Estructura.Inode, superblock *Estructura.SuperBlock, file *os.File) (int32, error) {
	var newBlockIndex int32 = -1
	for i := 0; i < int(superblock.S_Blocks_Count); i++ {
		var blockStatus byte
		if err := Utilidades.ReadObject(file, &blockStatus, int64(superblock.S_BM_Block_Start+int32(i))); err != nil {
			return -1, err
		}
		if blockStatus == 0 {
			newBlockIndex = int32(i)
			break
		}
	}

	if newBlockIndex == -1 {
		return -1, fmt.Errorf("no hay bloques disponibles\n ")
	}

	if err := Utilidades.WriteObject(file, byte(1), int64(superblock.S_BM_Block_Start+newBlockIndex)); err != nil {
		return -1, err
	}

	return newBlockIndex, nil
}

//eliminar grupo

func Rmgrp(grupo string, buffer *bytes.Buffer) {
	fmt.Fprintf(buffer, "======Start Rmgrp======\n ")
	if Usuario.Dato.GetIDUsuario() == "" && Usuario.Dato.GetIDParticion() == "" {
		fmt.Fprintf(buffer, "Error: No hay un usuario logueado\n ")
		return
	}
	if Usuario.Dato.GetIDUsuario() != "root" {
		fmt.Fprintf(buffer, "Error: El usuario no tiene permiso de escritura\n ")
		return
	}

	ParticionesMount := ManejadorDisco.GetMountedPartitions()
	var filepath string
	var id string

	for _, particiones := range ParticionesMount {
		for _, particion := range particiones {
			if particion.LoggedIn {
				filepath = particion.Path
				id = particion.ID
				break
			}
		}
		if filepath != "" {
			break
		}
	}

	file, err := Utilidades.OpenFile(filepath)
	if err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo abrir el archivo:%v", err)
		return
	}
	defer file.Close()

	var TempMBR Estructura.MRB
	if err := Utilidades.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo leer el MBR:%v", err)
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
		fmt.Fprintf(buffer, "Error: No se encontró la partición\n ")
		return
	}

	var tempSuperblock Estructura.SuperBlock
	if err := Utilidades.ReadObject(file, &tempSuperblock, int64(TempMBR.MRBPartitions[index].PART_Start)); err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo leer el SuperBlock:%v", err)
		return
	}

	indexInode := systemadministation.BuscarStart("/users.txt", file, tempSuperblock, buffer)
	if indexInode == -1 {
		fmt.Fprintf(buffer, "Error: No se encontró el archivo users.txt\n ")
		return
	}

	var crrInode Estructura.Inode
	if err := Utilidades.ReadObject(file, &crrInode, int64(tempSuperblock.S_Inode_Start+indexInode*int32(binary.Size(Estructura.Inode{})))); err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo leer el Inode del archivo users.txt\n ")
		return
	}

	if err := removeGroup(file, &tempSuperblock, &crrInode, grupo); err != nil {
		fmt.Fprintf(buffer, "Error:%v", err)
		return
	}

	// Guardar cambios del inodo
	if err := Utilidades.WriteObject(file, crrInode, int64(tempSuperblock.S_Inode_Start+indexInode*int32(binary.Size(Estructura.Inode{})))); err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo escribir el inodo actualizado\n ")
		return
	}

	fmt.Fprintf(buffer, "Grupo eliminado exitosamente"+grupo)
	fmt.Fprintf(buffer, "======End Rmgrp======\n ")
}

func removeGroup(file *os.File, superblock *Estructura.SuperBlock, inode *Estructura.Inode, grupo string) error {
	var newContent strings.Builder
	groupFound := false

	for _, block := range inode.I_Block {
		if block != -1 {
			var fileBlock Estructura.FileBlock
			if err := Utilidades.ReadObject(file, &fileBlock, int64(superblock.S_Block_Start+block*int32(binary.Size(Estructura.FileBlock{})))); err != nil {
				return fmt.Errorf("no se pudo leer el FileBlock\n ")
			}
			content := strings.TrimRight(string(fileBlock.B_Content[:]), "\x00")
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.Contains(line, "G") {
					parts := strings.Split(line, ",")
					if len(parts) > 2 && strings.TrimSpace(parts[2]) == grupo {
						groupFound = true
						continue // Skip this line to remove the group
					}
				}
				newContent.WriteString(line + "\n")
			}
		}
	}

	if !groupFound {
		return fmt.Errorf("el grupo '%s' no existe", grupo)
	}

	// Write the updated content back to the file blocks
	return writeUpdatedContent(file, superblock, inode, newContent.String())
}

func Rmusr(usuario string, buffer *bytes.Buffer) {
	if Usuario.Dato.GetIDUsuario() == "" && Usuario.Dato.GetIDParticion() == "" {
		fmt.Fprintf(buffer, "Error: No hay un usuario logueado\n ")
		return
	}
	if Usuario.Dato.GetIDUsuario() != "root" {
		fmt.Fprintf(buffer, "Error: El usuario no tiene permiso de escritura\n ")
		return
	}

	ParticionesMount := ManejadorDisco.GetMountedPartitions()
	var filepath string
	var id string

	for _, particiones := range ParticionesMount {
		for _, particion := range particiones {
			if particion.LoggedIn {
				filepath = particion.Path
				id = particion.ID
				break
			}
		}
		if filepath != "" {
			break
		}
	}

	file, err := Utilidades.OpenFile(filepath)
	if err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo abrir el archivo:%v", err)
		return
	}
	defer file.Close()

	var TempMBR Estructura.MRB
	if err := Utilidades.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo leer el MBR:%v", err)
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
		fmt.Fprintf(buffer, "Error: No se encontró la partición\n ")
		return
	}

	var tempSuperblock Estructura.SuperBlock
	if err := Utilidades.ReadObject(file, &tempSuperblock, int64(TempMBR.MRBPartitions[index].PART_Start)); err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo leer el SuperBlock:%v", err)
		return
	}

	indexInode := systemadministation.BuscarStart("/users.txt", file, tempSuperblock, buffer)
	if indexInode == -1 {
		fmt.Fprintf(buffer, "Error: No se encontró el archivo users.txt\n ")
		return
	}

	var crrInode Estructura.Inode
	if err := Utilidades.ReadObject(file, &crrInode, int64(tempSuperblock.S_Inode_Start+indexInode*int32(binary.Size(Estructura.Inode{})))); err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo leer el Inode del archivo users.txt\n ")
		return
	}

	if err := removeUser(file, &tempSuperblock, &crrInode, usuario); err != nil {
		fmt.Fprintf(buffer, "Error:%v", err)
		return
	}

	// Guardar cambios del inodo
	if err := Utilidades.WriteObject(file, crrInode, int64(tempSuperblock.S_Inode_Start+indexInode*int32(binary.Size(Estructura.Inode{})))); err != nil {
		fmt.Fprintf(buffer, "Error: No se pudo escribir el inodo actualizado\n ")
		return
	}

	fmt.Fprintf(buffer, "Usuario eliminado exitosamente\n ")
}

func removeUser(file *os.File, superblock *Estructura.SuperBlock, inode *Estructura.Inode, usuario string) error {
	var newContent strings.Builder
	userFound := false

	for _, block := range inode.I_Block {
		if block != -1 {
			var fileBlock Estructura.FileBlock
			if err := Utilidades.ReadObject(file, &fileBlock, int64(superblock.S_Block_Start+block*int32(binary.Size(Estructura.FileBlock{})))); err != nil {
				return fmt.Errorf("no se pudo leer el FileBlock\n ")
			}
			content := strings.TrimRight(string(fileBlock.B_Content[:]), "\x00")
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.Contains(line, "U") {
					parts := strings.Split(line, ",")
					if len(parts) >= 4 && strings.TrimSpace(parts[3]) == usuario {
						userFound = true
						continue // Skip this line to remove the user
					}
				}
				newContent.WriteString(line + "\n")
			}
		}
	}

	if !userFound {
		return fmt.Errorf("el usuario '%s' no existe", usuario)
	}

	// Write the updated content back to the file blocks
	return writeUpdatedContent(file, superblock, inode, newContent.String())
}

func writeUpdatedContent(file *os.File, superblock *Estructura.SuperBlock, inode *Estructura.Inode, content string) error {
	contentBytes := []byte(content)
	var currentBlock int32 = 0
	var currentOffset int = 0

	for currentOffset < len(contentBytes) {
		if currentBlock >= int32(len(inode.I_Block)) {
			return fmt.Errorf("no hay suficientes bloques disponibles\n ")
		}

		if inode.I_Block[currentBlock] == -1 {
			newBlockIndex, err := createNewFileBlock(inode, superblock, file)
			if err != nil {
				return err
			}
			inode.I_Block[currentBlock] = newBlockIndex
		}

		var fileBlock Estructura.FileBlock
		start := currentOffset
		end := currentOffset + 64
		if end > len(contentBytes) {
			end = len(contentBytes)
		}
		copy(fileBlock.B_Content[:], contentBytes[start:end])

		if err := Utilidades.WriteObject(file, fileBlock, int64(superblock.S_Block_Start+inode.I_Block[currentBlock]*int32(binary.Size(Estructura.FileBlock{})))); err != nil {
			return fmt.Errorf("error al escribir el bloque actualizado: %v", err)
		}

		currentOffset = end
		currentBlock++
	}

	inode.I_Size = int32(len(contentBytes))
	return nil
}

func Chgrp(user string, newGroup string, buffer *bytes.Buffer) {
	if Usuario.Dato.GetIDUsuario() == "" && Usuario.Dato.GetIDParticion() == "" {
		fmt.Fprintf(buffer, "Error: No hay un usuario logueado\n ")
		return
	}
	if Usuario.Dato.GetIDUsuario() != "root" {
		fmt.Fprintf(buffer, "Error: El usuario no tiene permiso de escritura\n ")
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
		if filepath != "" {
			break
		}
	}

	file, err := Utilidades.OpenFile(filepath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo:", err)
		return
	}
	defer file.Close()

	var TempMBR Estructura.MRB
	if err := Utilidades.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
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
		fmt.Println("Error: No se encontró la partición")
		return
	}

	var tempSuperblock Estructura.SuperBlock
	if err := Utilidades.ReadObject(file, &tempSuperblock, int64(TempMBR.MRBPartitions[index].PART_Start)); err != nil {
		fmt.Println("Error: No se pudo leer el superblock:", err)
		return
	}

	indexInode := systemadministation.BuscarStart("/users.txt", file, tempSuperblock, buffer)
	if indexInode == -1 {
		fmt.Println("Error: No se encontró el archivo /users.txt")
		return
	}

	var crrInode Estructura.Inode
	if err := Utilidades.ReadObject(file, &crrInode, int64(tempSuperblock.S_Inode_Start+indexInode*int32(binary.Size(Estructura.Inode{})))); err != nil {
		fmt.Println("Error: No se pudo leer el inodo del archivo /users.txt")
		return
	}

	data := readAllFileBlocks(&crrInode, file, tempSuperblock, buffer)
	cleanedData := LimpiarNull(data)

	if !userExists(cleanedData, user) {
		fmt.Println("Error: El usuario no existe")
		fmt.Fprintln(buffer, "Error: El usuario: "+user+" no existe")
		return
	}

	if !grupExiste(cleanedData, newGroup, buffer) {
		fmt.Println("Error: El grupo nuevo no existe")
		fmt.Fprintln(buffer, "Error: El grupo: "+newGroup+" no existe")
		return
	}

	updatedData, changed := updateUserGroup(cleanedData, user, newGroup)
	if !changed {
		fmt.Println("Error: No se pudo cambiar el grupo del usuario")
		return
	}

	if err := writeUpdatedUserData(&crrInode, updatedData, file, tempSuperblock); err != nil {
		fmt.Println("Error: No se pudo actualizar el archivo /users.txt:", err)
		return
	}

	fmt.Println("Grupo del usuario cambiado con éxito:", user, "al grupo", newGroup)
	fmt.Fprintln(buffer, "Grupo del usuario cambiado con éxito: "+user+" al grupo "+newGroup)
}

func updateUserGroup(data string, user string, newGroup string) (string, bool) {
	lines := strings.Split(data, "\n")
	changed := false
	for i, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) >= 4 && fields[1] == "U" && fields[3] == user {
			fields[2] = newGroup
			lines[i] = strings.Join(fields, ",")
			changed = true
			break
		}
	}
	return strings.Join(lines, "\n"), changed
}

func writeUpdatedUserData(inode *Estructura.Inode, updatedData string, file *os.File, superblock Estructura.SuperBlock) error {
	// Calcular el tamaño de un bloque de archivo
	blockSize := int32(binary.Size(Estructura.FileBlock{}))

	// Dividir los datos actualizados en bloques
	dataBlocks := []string{}
	for i := 0; i < len(updatedData); i += int(blockSize) {
		end := i + int(blockSize)
		if end > len(updatedData) {
			end = len(updatedData)
		}
		dataBlocks = append(dataBlocks, updatedData[i:end])
	}

	// Actualizar o crear bloques según sea necesario
	for i, block := range dataBlocks {
		if int32(i) >= int32(len(inode.I_Block)) {
			return fmt.Errorf("no hay suficientes bloques disponibles en el inodo")
		}

		// Si el bloque no existe, crear uno nuevo
		if inode.I_Block[i] == -1 {
			newBlockIndex, err := createNewFileBlock(inode, &superblock, file)
			if err != nil {
				return fmt.Errorf("error al crear un nuevo bloque: %v", err)
			}
			inode.I_Block[i] = newBlockIndex
		}

		// Escribir los datos en el bloque
		var fileBlock Estructura.FileBlock
		copy(fileBlock.B_Content[:], block)

		if err := Utilidades.WriteObject(file, fileBlock, int64(superblock.S_Block_Start+inode.I_Block[i]*blockSize)); err != nil {
			return fmt.Errorf("error al escribir el bloque %d: %v", i, err)
		}
	}

	// Actualizar el tamaño del inodo
	inode.I_Size = int32(len(updatedData))

	// Escribir el inodo actualizado
	if err := Utilidades.WriteObject(file, *inode, int64(superblock.S_Inode_Start+inode.I_Block[0]*int32(binary.Size(Estructura.Inode{})))); err != nil {
		return fmt.Errorf("error al actualizar el inodo: %v", err)
	}

	// Liberar bloques no utilizados si los datos actualizados son más pequeños
	for i := len(dataBlocks); i < len(inode.I_Block); i++ {
		if inode.I_Block[i] != -1 {
			// Aquí podrías implementar la lógica para marcar el bloque como libre en el bitmap de bloques
			inode.I_Block[i] = -1
		}
	}

	return nil
}
