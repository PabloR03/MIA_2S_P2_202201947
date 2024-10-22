package Reportes

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"proyecto1/Estructura"
	"proyecto1/ManejadorDisco"
	"proyecto1/Utilidades"
	"strings"
)

func Rep(name string, path string, id string, path_file_ls string, buffer *bytes.Buffer) {
	if name == "" {
		fmt.Fprintf(buffer, "Error REP: El tipo de reporte es obligatorio.\n")
		return
	}
	if path == "" {
		fmt.Fprintf(buffer, "Error REP: La ruta del reporte es obligatoria.\n")
		return
	}
	if id == "" {
		fmt.Fprintf(buffer, "Error REP: El ID de la partición es obligatoria.\n")
		return
	}
	if name == "mbr" {
		ReporteMBR(id, path, buffer)
	} else if name == "disk" {
		ReporteDisk(id, path, buffer)
	} else if name == "sb" {
		ReporteSB(id, path, buffer)
	} else if name == "bm_inode" {
		ReporteBMInode(id, path, buffer)
	} else if name == "bm_block" {
		ReporteBMBlock(id, path, buffer)
	} else {
		fmt.Fprintf(buffer, "Error REP: Tipo de reporte no válido.\n")
	}

}

func ReporteMBR(id string, path string, buffer *bytes.Buffer) {
	var ParticionesMontadas ManejadorDisco.PartitionMounted
	var ParticionEncontrada bool

	for _, particiones := range ManejadorDisco.GetMountedPartitions() {
		for _, particion := range particiones {
			if particion.ID == id {
				ParticionesMontadas = particion
				ParticionEncontrada = true
				break
			}
		}
		if ParticionEncontrada {
			break
		}
	}

	if !ParticionEncontrada {
		fmt.Fprintf(buffer, "Error REP MBR: No se encontró la partición con el ID: %s.\n", id)
		return
	}

	archivo, err := Utilidades.OpenFile(ParticionesMontadas.Path)
	if err != nil {
		return
	}
	defer archivo.Close()

	var MBRTemporal Estructura.MRB
	if err := Utilidades.ReadObject(archivo, &MBRTemporal, 0); err != nil {
		return
	}

	dot := "digraph G {\n"
	dot += "node [shape=plaintext];\n"
	dot += "tabla [label=<\n"
	dot += "<table border='1' cellborder='1' cellspacing='0'>\n"

	// Encabezado de la tabla
	dot += "<tr><td colspan='8' bgcolor='lightblue'><b>Reporte MBR</b></td></tr>\n"
	dot += "<tr><td bgcolor='lightgray'><b>Tipo</b></td><td bgcolor='lightgray'><b>Estado</b></td><td bgcolor='lightgray'><b>Tipo</b></td><td bgcolor='lightgray'><b>Ajuste</b></td><td bgcolor='lightgray'><b>Inicio</b></td><td bgcolor='lightgray'><b>Tamaño</b></td><td bgcolor='lightgray'><b>Nombre</b></td><td bgcolor='lightgray'><b>Correlativo</b></td></tr>\n"

	// Información del MBR
	dot += fmt.Sprintf("<tr><td bgcolor='blue'>MBR</td><td colspan='7'>Tamaño: %d, Fecha de Creación: %s, Ajuste: %s, Signature: %d</td></tr>\n",
		MBRTemporal.MRBSize, string(MBRTemporal.MRBCreationDate[:]), string(MBRTemporal.MRBFit[:]), MBRTemporal.MRBSignature)

	for _, particion := range MBRTemporal.MRBPartitions {
		if particion.PART_Size != 0 {
			tipoParticion := "Primaria"
			colorFondo := "lightpink"
			if particion.PART_Type[0] == 'e' {
				tipoParticion = "Extendida"
				colorFondo = "lightyellow"
			}

			dot += fmt.Sprintf("<tr><td bgcolor='%s'>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%d</td><td>%d</td><td>%s</td><td>%d</td></tr>\n",
				colorFondo, tipoParticion, string(particion.PART_Status[:]), string(particion.PART_Type[:]),
				string(particion.PART_Fit[:]), particion.PART_Start, particion.PART_Size,
				strings.Trim(string(particion.PART_Name[:]), "\x00"), particion.PART_Correlative)

			if particion.PART_Type[0] == 'e' {
				var EBR Estructura.EBR
				if err := Utilidades.ReadObject(archivo, &EBR, int64(particion.PART_Start)); err != nil {
					return
				}
				if EBR.EBRSize != 0 {
					var ContadorLogicas int = 0
					for {
						// Información del EBR
						dot += fmt.Sprintf("<tr><td bgcolor='lightgreen'>EBR</td><td colspan='7'>Nombre: %s, Ajuste: %s, Inicio: %d, Tamaño: %d, Siguiente: %d</td></tr>\n",
							strings.Trim(string(EBR.EBRName[:]), "\x00"), string(EBR.EBRFit[:]), EBR.EBRStart, EBR.EBRSize, EBR.EBRNext)

						// Información de la partición lógica
						dot += fmt.Sprintf("<tr><td bgcolor='lavender'>Lógica</td><td>0</td><td>l</td><td>%s</td><td>%d</td><td>%d</td><td>%s</td><td>%d</td></tr>\n",
							string(EBR.EBRFit[:]), EBR.EBRStart, EBR.EBRSize, strings.Trim(string(EBR.EBRName[:]), "\x00"), ContadorLogicas+1)

						if EBR.EBRNext == -1 {
							break
						}
						if err := Utilidades.ReadObject(archivo, &EBR, int64(EBR.EBRNext)); err != nil {
							fmt.Fprintf(buffer, "Error al leer siguiente EBR: %v\n", err)
							return
						}
						ContadorLogicas++
					}
				}
			}
		}
	}

	dot += "</table>\n"
	dot += ">];\n"
	dot += "}\n"

	dotFilePath := "REPORTEMBR.dot"
	err = os.WriteFile(dotFilePath, []byte(dot), 0644)
	if err != nil {
		fmt.Fprintf(buffer, "Error REP MBR: Error al escribir el archivo DOT.\n")
		return
	}

	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Fprintf(buffer, "Error REP MBR: Error al crear el directorio.\n")
			return
		}
	}

	cmd := exec.Command("dot", "-Tjpg", dotFilePath, "-o", path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(buffer, "Error REP MBR: Error al ejecutar Graphviz.\n")
		fmt.Println("Error al ejecutar Graphviz:", err)
		return
	}
	fmt.Fprintf(buffer, "Reporte de MBR generado exitosamente en la ruta: %s\n", path)
}

func ReporteDisk(id string, path string, buffer *bytes.Buffer) {
	var ParticionesMontadas ManejadorDisco.PartitionMounted
	var ParticionEncontrada bool

	for _, particiones := range ManejadorDisco.GetMountedPartitions() {
		for _, particion := range particiones {
			if particion.ID == id {
				ParticionesMontadas = particion
				ParticionEncontrada = true
				break
			}
		}
		if ParticionEncontrada {
			break
		}
	}

	if !ParticionEncontrada {
		fmt.Fprintf(buffer, "Error REP DISK: No se encontró la partición con el ID: %s.\n", id)
		return
	}

	archivo, err := Utilidades.OpenFile(ParticionesMontadas.Path)
	if err != nil {
		return
	}
	defer archivo.Close()

	var MBRTemporal Estructura.MRB
	if err := Utilidades.ReadObject(archivo, &MBRTemporal, 0); err != nil {
		return
	}

	// Variables para calcular el porcentaje
	totalSize := float64(MBRTemporal.MRBSize)
	usedSize := 0.0

	textoDot := "digraph G {\n"
	textoDot += "node [shape=record, height=.1, fontsize=10];\n"
	textoDot += "rankdir=LR;\n"
	textoDot += "ranksep=0.1;\n"
	textoDot += "nodesep=0.1;\n"
	textoDot += "subgraph cluster0 {\n"
	textoDot += "label=\"Disco\";\n"
	textoDot += "style=\"rounded,filled\";\n"
	textoDot += "color=lightgrey;\n"
	textoDot += "node [style=filled, color=white];\n"

	// MBR siempre al inicio
	textoDot += "MBR [label=\"MBR\"];\n"
	lastNode := "MBR"

	// Analizando particiones
	for i, partition := range MBRTemporal.MRBPartitions {
		if partition.PART_Status[0] != 0 {
			partSize := float64(partition.PART_Size)
			usedSize += partSize

			nodeName := fmt.Sprintf("P%d", i+1)
			if partition.PART_Type[0] == 'e' || partition.PART_Type[0] == 'E' {
				// Partición extendida
				textoDot += fmt.Sprintf("%s [label=\"{Extendida|%.2f%%}|{", nodeName, (partSize/totalSize)*100)

				// Particiones lógicas dentro de la extendida
				extendedFreeSpace := partSize
				finEbr := partition.PART_Start
				logicalCount := 0
				for {
					var ebr Estructura.EBR
					if err := Utilidades.ReadObject(archivo, &ebr, int64(finEbr)); err != nil {
						break
					}

					ebrSize := float64(ebr.EBRSize)
					usedSize += ebrSize
					extendedFreeSpace -= ebrSize
					logicalCount++

					textoDot += fmt.Sprintf("{EBR|{Lógica %d|%.2f%%}}|", logicalCount, (ebrSize/totalSize)*100)

					if ebr.EBRNext <= 0 {
						break
					}
					finEbr = ebr.EBRNext
				}

				// Espacio libre en la extendida
				if extendedFreeSpace > 0 {
					textoDot += fmt.Sprintf("{Libre en Ext|%.2f%%}", (extendedFreeSpace/totalSize)*100)
				} else {
					textoDot = textoDot[:len(textoDot)-1] // Remover el último "|" si no hay espacio libre
				}

				textoDot += "}\"];\n"
			} else {
				// Partición primaria
				textoDot += fmt.Sprintf("%s [label=\"{Partición %d|%.2f%%}\"];\n", nodeName, i+1, (partSize/totalSize)*100)
			}
			textoDot += fmt.Sprintf("%s -> %s [style=invis];\n", lastNode, nodeName)
			lastNode = nodeName
		}
	}

	// Espacio libre general restante en el disco
	freeSize := totalSize - usedSize
	if freeSize > 0 {
		freeNodeName := "FreeSpace"
		textoDot += fmt.Sprintf("%s [label=\"{Libre|%.2f%%}\"];\n", freeNodeName, (freeSize/totalSize)*100)
		textoDot += fmt.Sprintf("%s -> %s [style=invis];\n", lastNode, freeNodeName)
	}

	textoDot += "}\n"
	textoDot += "}\n"

	// Guardar el archivo .dot y generar la imagen
	rutaDot := "ReporteDISK.dot"
	err = os.WriteFile(rutaDot, []byte(textoDot), 0644)
	if err != nil {
		fmt.Fprintf(buffer, "Error al escribir el archivo DOT")
		fmt.Println("Error al escribir el archivo DOT:", err)
		return
	}

	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Fprintf(buffer, "Error al crear el directorio")
			fmt.Println("Error al crear el directorio:", err)
			return
		}
	}

	cmd := exec.Command("dot", "-Tjpg", rutaDot, "-o", path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(buffer, "Error al ejecutar Graphviz")
		fmt.Println("Error al ejecutar Graphviz:", err)
		fmt.Println("Detalles del error:", stderr.String())
		return
	}

	fmt.Fprintf(buffer, "Reporte de uso de disco generado exitosamente en "+path)
	fmt.Println("Reporte de uso de disk generado exitosamente")
	fmt.Println("====== FIN REP USO DISK ======")
}

func ReporteSB(id string, path string, buffer *bytes.Buffer) {
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
		fmt.Fprintf(buffer, "Error REP SB: No se encontró la partición con el ID: %s.\n", id)
		return
	}

	archivo, err := Utilidades.OpenFile(ParticionesMontadas.Path)
	if err != nil {
		return
	}
	defer archivo.Close()

	var MBRTemporal Estructura.MRB
	if err := Utilidades.ReadObject(archivo, &MBRTemporal, 0); err != nil {
		return
	}

	index := 0
	for i := 0; i < 4; i++ {
		if strings.Contains(string(MBRTemporal.MRBPartitions[i].PART_Id[:]), id) {
			index = i
			break
		}
	}

	var TemporalSuperBloque = Estructura.SuperBlock{}
	if err := Utilidades.ReadObject(archivo, &TemporalSuperBloque, int64(MBRTemporal.MRBPartitions[index].PART_Start)); err != nil {
		return
	}

	dot := "digraph G {\n"
	dot += "node [shape=plaintext];\n"
	dot += "fontname=\"Helvetica,Arial,sans-serif\";\n"
	dot += "bgcolor=\"#F0F8FF\";\n" // Light blue background
	dot += "title [label=\"REPORTE SB\", fontsize=20, fontcolor=\"#4B0082\"];\n"
	dot += "SBTable [label=<\n"
	dot += "<table border='0' cellborder='1' cellspacing='0' cellpadding='8'>\n"
	dot += "<tr><td bgcolor=\"#4169E1\" colspan='2'><font color=\"white\"><b>Super Bloque</b></font></td></tr>\n"

	// Function to alternate row colors
	rowColor := func(i int) string {
		if i%2 == 0 {
			return "#E6F3FF"
		}
		return "#FFFFFF"
	}

	// Add rows with alternating colors
	addRow := func(i int, label string, value interface{}) string {
		return fmt.Sprintf("<tr><td bgcolor=\"%s\"><font color=\"#333333\">%s</font></td><td bgcolor=\"%s\"><font color=\"#0000CD\">%v</font></td></tr>\n", rowColor(i), label, rowColor(i), value)
	}

	dot += addRow(0, "SB FileSystem Type", int(TemporalSuperBloque.S_Filesystem_Type))
	dot += addRow(1, "SB Inodes Count", int(TemporalSuperBloque.S_Inodes_Count))
	dot += addRow(2, "SB Blocks Count", int(TemporalSuperBloque.S_Blocks_Count))
	dot += addRow(3, "SB Free Blocks Count", int(TemporalSuperBloque.S_Free_Blocks_Count))
	dot += addRow(4, "SB Free Inodes Count", int(TemporalSuperBloque.S_Free_Inodes_Count))
	dot += addRow(5, "SB Mtime", string(TemporalSuperBloque.S_Mtime[:]))
	dot += addRow(6, "SB Umtime", string(TemporalSuperBloque.S_Umtime[:]))
	dot += addRow(7, "SB Mnt Count", int(TemporalSuperBloque.S_Mnt_Count))
	dot += addRow(8, "SB Magic", int(TemporalSuperBloque.S_Magic))
	dot += addRow(9, "SB Inode Size", int(TemporalSuperBloque.S_Inode_Size))
	dot += addRow(10, "SB Block Size", int(TemporalSuperBloque.S_Block_Size))
	dot += addRow(11, "SB First Inode", int(TemporalSuperBloque.S_Fist_Ino))
	dot += addRow(12, "SB First Block", int(TemporalSuperBloque.S_First_Blo))
	dot += addRow(13, "SB Bm Inode Start", int(TemporalSuperBloque.S_BM_Inode_Start))
	dot += addRow(14, "SB Bm Block Start", int(TemporalSuperBloque.S_BM_Block_Start))
	dot += addRow(15, "SB Inode Start", int(TemporalSuperBloque.S_Inode_Start))
	dot += addRow(16, "SB Block Start", int(TemporalSuperBloque.S_Block_Start))

	dot += "</table>\n"
	dot += ">];\n"
	dot += "}\n"

	RutaReporte := "REPORTESB.dot"
	err = os.WriteFile(RutaReporte, []byte(dot), 0644)
	if err != nil {
		fmt.Fprintf(buffer, "Error REP SB: Error al escribir el archivo DOT.\n")
		fmt.Println("Error REP DISK:", err)
		return
	}
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Fprintf(buffer, "Error REP SB: Error al crear el directorio.\n")
			fmt.Println("Error REP DISK:", err)
			return
		}
	}
	cmd := exec.Command("dot", "-Tjpg", RutaReporte, "-o", path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(buffer, "Error REP SB: Error al ejecutar Graphviz.")
		fmt.Println("Error REP DISK:", err)
		return
	}
	fmt.Fprintf(buffer, "Reporte de SB de la partición:%s generado con éxito en la ruta: %s\n", id, path)
}

func ReporteBMInode(id string, path string, buffer *bytes.Buffer) {

	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Println("Error al crear el directorio:", err)
			return
		}
	}

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
		fmt.Fprintf(buffer, "Error REP SB: No se encontró la partición con el ID: %s.\n", id)
		return
	}

	archivo, err := Utilidades.OpenFile(ParticionesMontadas.Path)
	if err != nil {
		return
	}
	defer archivo.Close()

	var MBRTemporal Estructura.MRB
	if err := Utilidades.ReadObject(archivo, &MBRTemporal, 0); err != nil {
		return
	}

	var index int = -1
	for i := 0; i < 4; i++ {
		if MBRTemporal.MRBPartitions[i].PART_Size != 0 {
			if strings.Contains(string(MBRTemporal.MRBPartitions[i].PART_Id[:]), id) {
				if MBRTemporal.MRBPartitions[i].PART_Status[0] == '1' {
					index = i
				} else {
					fmt.Fprintf(buffer, "Error REP SB: La partición con el ID:%s no está montada.\n", id)
					return
				}
				break
			}
		}
	}

	if index == -1 {
		fmt.Fprintf(buffer, "Error REP SB: No se encontró la partición con el ID: %s.\n", id)
		return
	}

	var TemporalSuperBloque = Estructura.SuperBlock{}
	if err := Utilidades.ReadObject(archivo, &TemporalSuperBloque, int64(MBRTemporal.MRBPartitions[index].PART_Start)); err != nil {
		fmt.Fprintf(buffer, "Error REP SB: Error al leer el SuperBloque.\n")
		return
	}

	bitmapInode := make([]byte, TemporalSuperBloque.S_Inodes_Count)
	if _, err := archivo.ReadAt(bitmapInode, int64(TemporalSuperBloque.S_BM_Inode_Start)); err != nil {
		fmt.Println("Error: No se pudo leer el bitmap de inodos:", err)
		return
	}

	outputFile, err := os.Create(path)
	if err != nil {
		fmt.Println("Error al crear el archivo de reporte:", err)
		return
	}
	defer outputFile.Close()

	// Escribir el reporte en el archivo de texto
	fmt.Fprintln(outputFile, "REPORTE BITMAP INODE")
	fmt.Fprintln(outputFile, "")

	// Mostrar 20 bits por línea
	for i, bit := range bitmapInode {
		if i > 0 && i%20 == 0 {
			// Nueva línea cada 20 bits
			fmt.Fprintln(outputFile)
		}
		fmt.Fprintf(outputFile, "%d ", bit)
	}

	fmt.Fprintf(buffer, "Reporte de BITMAP INODE de la partición:%s generado con éxito en la ruta: %s\n", id, path)
}

func ReporteBMBlock(id string, path string, buffer *bytes.Buffer) {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Println("Error al crear el directorio:", err)
			return
		}
	}

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
		fmt.Fprintf(buffer, "Error REP SB: No se encontró la partición con el ID: %s.\n", id)
		return
	}

	archivo, err := Utilidades.OpenFile(ParticionesMontadas.Path)
	if err != nil {
		return
	}
	defer archivo.Close()

	var MBRTemporal Estructura.MRB
	if err := Utilidades.ReadObject(archivo, &MBRTemporal, 0); err != nil {
		return
	}

	var index int = -1
	for i := 0; i < 4; i++ {
		if MBRTemporal.MRBPartitions[i].PART_Size != 0 {
			if strings.Contains(string(MBRTemporal.MRBPartitions[i].PART_Id[:]), id) {
				if MBRTemporal.MRBPartitions[i].PART_Status[0] == '1' {
					index = i
				} else {
					fmt.Fprintf(buffer, "Error REP SB: La partición con el ID:%s no está montada.\n", id)
					return
				}
				break
			}
		}
	}

	if index == -1 {
		fmt.Fprintf(buffer, "Error REP SB: No se encontró la partición con el ID: %s.\n", id)
		return
	}

	var TemporalSuperBloque = Estructura.SuperBlock{}
	if err := Utilidades.ReadObject(archivo, &TemporalSuperBloque, int64(MBRTemporal.MRBPartitions[index].PART_Start)); err != nil {
		fmt.Fprintf(buffer, "Error REP SB: Error al leer el SuperBloque.\n")
		return
	}

	// Leer el bitmap de bloques desde el archivo binario
	bitmapBlock := make([]byte, TemporalSuperBloque.S_Blocks_Count)
	if _, err := archivo.ReadAt(bitmapBlock, int64(TemporalSuperBloque.S_BM_Block_Start)); err != nil {
		fmt.Println("Error: No se pudo leer el bitmap de bloques:", err)
		return
	}

	// Crear el archivo de salida para el reporte
	outputFile, err := os.Create(path)
	if err != nil {
		fmt.Println("Error al crear el archivo de reporte:", err)
		return
	}
	defer outputFile.Close()

	// Escribir el reporte en el archivo de texto
	fmt.Fprintln(outputFile, "Reporte Bitmap de Bloques")
	fmt.Fprintln(outputFile, "")

	// Mostrar 20 bits por línea
	for i, bit := range bitmapBlock {
		if i > 0 && i%20 == 0 {
			// Nueva línea cada 20 bits
			fmt.Fprintln(outputFile)
		}
		fmt.Fprintf(outputFile, "%d ", bit)
	}

	fmt.Fprintf(buffer, "Reporte de BITMAP BLOCK de la partición:%s generado con éxito en la ruta: %s\n", id, path)
}
