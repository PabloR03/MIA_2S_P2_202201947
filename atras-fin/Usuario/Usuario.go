package Usuario

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"proyecto1/Estructura"
	"proyecto1/ManejadorDisco"
	"proyecto1/Utilidades"
	"strings"
)

type ParticionUsuario struct {
	IDParticion string
	IDUsuario   string
}

func (Dato *ParticionUsuario) GetIDParticion() string {
	return Dato.IDParticion
}

func (Dato *ParticionUsuario) GetIDUsuario() string {
	return Dato.IDUsuario
}

func (Dato *ParticionUsuario) SetIDParticion(idParticion string) {
	Dato.IDParticion = idParticion
}

func (Dato *ParticionUsuario) SetIDUsuario(idUsuario string) {
	Dato.IDUsuario = idUsuario
}

var Dato ParticionUsuario

func Login(user string, pass string, id string, buffer *bytes.Buffer) {
	fmt.Fprint(buffer, "=-=-=-=-=-=-=INICIO LOGIN=-=-=-=-=-=-=")
	ParticionesMontadas := ManejadorDisco.GetMountedPartitions()
	var RutaArchivo string
	var ParticionEncontrada bool
	var Login bool = false

	print("Hola ParticionesMontadas: ", ParticionesMontadas)
	for _, Particiones := range ParticionesMontadas {
		for _, Particion := range Particiones {
			if Particion.ID == id && Particion.LoggedIn {
				fmt.Fprintf(buffer, "Error LOGIN: Ya existe un usuario logueado en la partición:%s\n", id)
				return
			}
			if Particion.ID == id {
				RutaArchivo = Particion.Path
				ParticionEncontrada = true
				break
			}
		}
		if ParticionEncontrada {
			break
		}
	}
	print("Particion no enconrada")
	if !ParticionEncontrada {
		fmt.Fprintf(buffer, "Error LOGIN: No se encontró ninguna partición montada con el ID: %s\n", id)
		return
	}

	file, err := Utilidades.OpenFile(RutaArchivo)
	if err != nil {
		return
	}
	defer file.Close()

	var MBRTemporal Estructura.MRB

	if err := Utilidades.ReadObject(file, &MBRTemporal, 0); err != nil {
		return
	}

	var index int = -1
	for i := 0; i < 4; i++ {
		if MBRTemporal.MRBPartitions[i].PART_Size != 0 {
			if strings.Contains(string(MBRTemporal.MRBPartitions[i].PART_Id[:]), id) {
				if MBRTemporal.MRBPartitions[i].PART_Status[0] == '1' {
					index = i
				} else {
					return
				}
				break
			}
		}
	}
	println("hoay particion en id ")
	if index == -1 {
		fmt.Fprintf(buffer, "Error LOGIN: No se encontró ninguna partición con el ID: %s\n", id)
		return
	}

	var SuperBloqueTemporal Estructura.SuperBlock
	if err := Utilidades.ReadObject(file, &SuperBloqueTemporal, int64(MBRTemporal.MRBPartitions[index].PART_Start)); err != nil {
		return
	}

	IndexInode := InitSearch("/users.txt", file, SuperBloqueTemporal, buffer)

	var crrInode Estructura.Inode
	// Leer el Inodo desde el archivo binario
	if err := Utilidades.ReadObject(file, &crrInode, int64(SuperBloqueTemporal.S_Inode_Start+IndexInode*int32(binary.Size(Estructura.Inode{})))); err != nil {
		return
	}
	data := GetInodeFileData(crrInode, file, SuperBloqueTemporal, buffer)

	lines := strings.Split(data, "\n")
	for _, line := range lines {
		words := strings.Split(line, ",")
		println("words: %s\n", len(words))
		if len(words) == 5 {
			if (strings.Contains(words[3], user)) && (strings.Contains(words[4], pass)) {
				Login = true
				break
			}
		}
	}
	println("login yes")
	if Login {
		fmt.Fprintf(buffer, "=-=-=-=-=-=-=FIN LOGIN=-=-=-=-=-=-=")
		fmt.Fprintf(buffer, "Usuario logueado con éxito en la partición:%s\n", id)
		ManejadorDisco.MarkPartitionAsLoggedIn(id)
	}
	Dato.SetIDParticion(id)
	Dato.SetIDUsuario(user)
	println(Login)
}

func Logout(buffer *bytes.Buffer) {
	fmt.Fprint(buffer, "=-=-=-=-=-=-=INCIO LOGOUT=-=-=-=-=-=-=")
	ParticionesMontadas := ManejadorDisco.GetMountedPartitions()
	var SesionActiva bool
	print("Hola ParticionesMontadas: ", ParticionesMontadas)
	if len(ParticionesMontadas) == 0 {
		fmt.Fprintf(buffer, "Error LOGOUT: No hay ninguna partición montada.\n")
		return
	}

	for _, Particiones := range ParticionesMontadas {
		for _, Particion := range Particiones {
			if Particion.LoggedIn {
				SesionActiva = true
				break
			}
		}
		if SesionActiva {
			break
		}
	}
	if !SesionActiva {
		fmt.Fprintf(buffer, "Error LOGOUT: No hay ninguna sesión activa.\n")
		return
	} else {
		ManejadorDisco.MarkPartitionAsLoggedOut(Dato.GetIDParticion())
		fmt.Fprintf(buffer, "=-=-=-=-=-=-=FIN LOGOUT=-=-=-=-=-=-=")
		fmt.Fprintf(buffer, "Sesión cerrada con éxito de la partición:%s\n", Dato.GetIDParticion())
	}
	Dato.SetIDParticion("")
	Dato.SetIDUsuario("")
}

func InitSearch(path string, file *os.File, SuperBloqueTemporal Estructura.SuperBlock, buffer *bytes.Buffer) int32 {
	TempStepsPath := strings.Split(path, "/")
	StepsPath := TempStepsPath[1:]
	var Inode0 Estructura.Inode
	if err := Utilidades.ReadObject(file, &Inode0, int64(SuperBloqueTemporal.S_Inode_Start)); err != nil {
		return -1
	}
	return SarchInodeByPath(StepsPath, Inode0, file, SuperBloqueTemporal, buffer)
}

func SarchInodeByPath(StepsPath []string, Inode Estructura.Inode, file *os.File, SuperBloqueTemporal Estructura.SuperBlock, buffer *bytes.Buffer) int32 {
	index := int32(0)
	SearchedName := strings.Replace(pop(&StepsPath), " ", "", -1)
	for _, block := range Inode.I_Block {
		if block != -1 {
			if index < 13 {
				var crrFolderBlock Estructura.FolderBlock
				if err := Utilidades.ReadObject(file, &crrFolderBlock, int64(SuperBloqueTemporal.S_Block_Start+block*int32(binary.Size(Estructura.FolderBlock{})))); err != nil {
					return -1
				}
				for _, folder := range crrFolderBlock.B_Content {
					if strings.Contains(string(folder.B_Name[:]), SearchedName) {
						if len(StepsPath) == 0 {
							return folder.B_Inodo
						} else {
							var NextInode Estructura.Inode
							if err := Utilidades.ReadObject(file, &NextInode, int64(SuperBloqueTemporal.S_Inode_Start+folder.B_Inodo*int32(binary.Size(Estructura.Inode{})))); err != nil {
								return -1
							}
							return SarchInodeByPath(StepsPath, NextInode, file, SuperBloqueTemporal, buffer)
						}
					}
				}
			}
		}
		index++
	}
	return 0
}

func GetInodeFileData(Inode Estructura.Inode, file *os.File, SuperBloqueTemporal Estructura.SuperBlock, buffer *bytes.Buffer) string {
	index := int32(0)
	var content string
	for _, block := range Inode.I_Block {
		if block != -1 {
			if index < 13 {
				var crrFileBlock Estructura.FileBlock
				if err := Utilidades.ReadObject(file, &crrFileBlock, int64(SuperBloqueTemporal.S_Block_Start+block*int32(binary.Size(Estructura.FileBlock{})))); err != nil {
					return ""
				}
				content += string(crrFileBlock.B_Content[:])
			}
		}
		index++
	}
	return content
}

func pop(s *[]string) string {
	lastIndex := len(*s) - 1
	last := (*s)[lastIndex]
	*s = (*s)[:lastIndex]
	return last
}

func IsUserLoggedIn() bool {
	return Dato.GetIDUsuario() != ""
}
