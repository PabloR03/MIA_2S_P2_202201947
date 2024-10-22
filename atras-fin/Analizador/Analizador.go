package Analizador

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"proyecto1/ManejadorArchivo"
	"proyecto1/ManejadorDisco"
	"proyecto1/ManejoRoot"
	"proyecto1/Reportes"
	"proyecto1/Usuario"
	"regexp"
	"strconv"
	"strings"
)

var re = regexp.MustCompile(`-(\w+)=("[^"]+"|\S+)`)

func Analizar(texto string) string {
	var buffer bytes.Buffer

	scanner := bufio.NewScanner(strings.NewReader(texto))
	for scanner.Scan() {
		entrada := scanner.Text()
		if len(entrada) == 0 || entrada[0] == '#' {
			fmt.Fprintf(&buffer, "%s\n", entrada)
			continue
		}
		entrada = strings.TrimSpace(entrada)
		command, params := getCommandAndParams(entrada)
		println("Comando:", command, "Parametros:", params)
		// fmt.Fprintln(&buffer, "Comando:", command, "Parametros:", params)
		AnalyzeCommnad(command, params, &buffer)
	}

	return buffer.String()
}

func AnalyzeCommnad(command string, params string, buffer *bytes.Buffer) {
	// Pasa el buffer a las funciones
	if strings.Contains(command, "mkdisk") {
		Funcion_mkdisk(params, buffer)
	} else if strings.Contains(command, "fdisk") {
		Funcion_fdisk(params, buffer)
	} else if strings.Contains(command, "rmdisk") {
		Funcion_rmdisk(params, buffer)
	} else if strings.Contains(command, "mount") {
		Funcion_mount(params, buffer)
	} else if strings.Contains(command, "mkfs") {
		Funcion_mkfs(params, buffer)
	} else if strings.Contains(command, "login") {
		Funcion_login(params, buffer)
	} else if strings.Contains(command, "rep") {
		Funcion_Rep(params, buffer)
	} else if strings.Contains(command, "ldisk") {
		Funcion_ldisk(buffer)
	} else if strings.Contains(command, "logout") {
		Funcion_Logout(params, buffer)
	} else if strings.Contains(command, "cat") {
		Funcion_cat(params, command, buffer)
	} else if strings.Contains(command, "mkusr") {
		Funcion_Mkusr(params, buffer)
	} else if strings.Contains(command, "mkgrp") {
		Funcion_Mkgrp(params, buffer)
	} else if strings.Contains(command, "rmgrp") {
		Funcion_Rmgrp(params, buffer)
	} else if strings.Contains(command, "rmusr") {
		Funcion_Rmusr(params, buffer)
	} else if strings.Contains(command, "chgrp") {
		Funcion_Chgrp(params, buffer)
	} else {
		fmt.Fprintf(buffer, "Error: Comando no encontrado.\n")
	}
}

func getCommandAndParams(input string) (string, string) {
	parts := strings.Fields(input)
	if len(parts) > 0 {
		command := strings.ToLower(parts[0])
		for i := 1; i < len(parts); i++ {
			parts[i] = strings.ToLower(parts[i])
		}
		params := strings.Join(parts[1:], " ")
		return command, params
	}
	return "", input
}

// ya revisado
func Funcion_mkdisk(params string, writer io.Writer) {
	// Define flags
	fs := flag.NewFlagSet("mkdisk", flag.ExitOnError)
	size := fs.Int("size", 0, "Tamano")
	fit := fs.String("fit", "ff", "Ajuste")
	unit := fs.String("unit", "m", "Unidad")
	path := fs.String("path", "", "Ruta")

	// Parse los argumentos desde params en lugar de os.Args
	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		nombreFlag := match[1]
		valorFlag := strings.ToLower(match[2])
		valorFlag = strings.Trim(valorFlag, "\"")
		switch nombreFlag {
		case "size", "fit", "unit", "path":
			fs.Set(nombreFlag, valorFlag)
		default:
			println("Error: Parámetro no encontrado.")
			fmt.Fprint(writer, "Error: Parámetro no encontrado.\n")
			return
		}
	}

	fs.Parse([]string{})

	ManejadorDisco.Mkdisk(*size, *fit, *unit, *path, writer.(*bytes.Buffer))
}

// ya revisado
func Funcion_rmdisk(params string, writer io.Writer) {
	fs := flag.NewFlagSet("rmdisk", flag.ExitOnError)
	path := fs.String("path", "", "Ruta")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		nombreFlag := match[1]
		valorFlag := strings.ToLower(match[2])
		valorFlag = strings.Trim(valorFlag, "\"")
		switch nombreFlag {
		case "path":
			fs.Set(nombreFlag, valorFlag)
		default:
			fmt.Fprint(writer, "Error: Parámetro no encontrado.\n")
			return
		}
	}
	ManejadorDisco.Rmdisk(*path, writer.(*bytes.Buffer))
}

// ya revisado
func Funcion_fdisk(input string, writer io.Writer) {
	fs := flag.NewFlagSet("fdisk", flag.ExitOnError)
	size := fs.Int("size", 0, "Tamaño")
	unit := fs.String("unit", "k", "Unidad")
	path := fs.String("path", "", "Ruta")
	type_ := fs.String("type", "p", "Tipo")
	fit := fs.String("fit", "wf", "Ajuste")
	delete_ := fs.String("delete", "", "Eliminar")
	name := fs.String("name", "", "Nombre")
	//add := fs.String("add", "", "Agregar")

	// Parsear los flags
	fs.Parse(os.Args[1:])

	// Encontrar los flags en el input
	matches := re.FindAllStringSubmatch(input, -1)

	// Procesar el input
	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.ToLower(match[2])

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "size", "fit", "unit", "path", "name", "type", "delete", "add":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Etiqueta no encontrada")
			return
		}
	}

	// Validaciones para la opción -delete
	if *delete_ != "" {
		if *path == "" || *name == "" {
			fmt.Println("Error: Para eliminar una partición, se requiere 'path' y 'name'.")
			return
		}
		// Llamar a la función que elimina la partición
		ManejadorDisco.DeletePartition(*path, *name, *delete_, writer.(*bytes.Buffer))
		return
	}

	// Validaciones
	if *size <= 0 {
		fmt.Println("Error: El tamaño debe ser mayor a 0")
		return
	}

	if *path == "" {
		fmt.Println("Error: Path/Ruta es obligatorio")
		return
	}

	// Si no se proporcionó un fit, usar el valor predeterminado "w"
	if *fit == "" {
		*fit = "wf"
	}

	// Validar fit (b/w/f)
	if *fit != "bf" && *fit != "ff" && *fit != "wf" {
		fmt.Println("Error: El ajuste debe ser 'bf', 'ff', o 'wf'")
		return
	}

	if *unit != "k" && *unit != "m" && *unit != "b" {
		fmt.Println("Error: Las unidades deben ser 'k' o 'm'")
		return
	}

	if *type_ != "p" && *type_ != "e" && *type_ != "l" {
		fmt.Println("Error: el tipo debe ser 'p', 'e', o 'l'")
		return
	}
	ManejadorDisco.Fdisk(*size, *path, *name, *unit, *type_, *fit, writer.(*bytes.Buffer))
}

// ya revisada
func Funcion_mount(input string, writer io.Writer) {
	fs := flag.NewFlagSet("mount", flag.ExitOnError)
	path := fs.String("path", "", "Ruta")
	name := fs.String("name", "", "Nombre de la partición")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.ToLower(match[2]) // Convertir todo a minúsculas
		flagValue = strings.Trim(flagValue, "\"")
		fs.Set(flagName, flagValue)
	}

	if *path == "" || *name == "" {
		fmt.Println("Error: Path y Name son obligatorios")
		return
	}

	// Convertir el nombre a minúsculas antes de pasarlo al Mount
	lowercaseName := strings.ToLower(*name)
	ManejadorDisco.Mount(*path, lowercaseName, writer.(*bytes.Buffer))
}

// ya revisada
func Funcion_mkfs(input string, writer io.Writer) {
	fs := flag.NewFlagSet("mkfs", flag.ExitOnError)
	id := fs.String("id", "", "ID")
	type_ := fs.String("type", "", "Tipo")
	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		nombreFlag := match[1]
		valorFlag := strings.ToLower(match[2])

		valorFlag = strings.Trim(valorFlag, "\"")

		switch nombreFlag {
		case "id", "type":
			fs.Set(nombreFlag, valorFlag)
		default:
			fmt.Fprint(writer, "Error: Parámetro no encontrado.\n")
			return
		}
	}
	ManejadorArchivo.Mkfs(*id, *type_, writer.(*bytes.Buffer))
}

// Función para ejecutar el comando LOGIN
func Funcion_login(input string, buffer io.Writer) {
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	user := fs.String("user", "", "Usuario")
	pass := fs.String("pass", "", "Contraseña")
	id := fs.String("id", "", "Id")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		flagName := match[1]
		flagValue := match[2]

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "user", "pass", "id":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
			return
		}
	}

	Usuario.Login(*user, *pass, *id, buffer.(*bytes.Buffer))

}

// Función para ejecutar el comando LOGOUT
func Funcion_Logout(entrada string, buffer io.Writer) {
	entrada = strings.TrimSpace(entrada)
	if len(entrada) > 0 {
		fmt.Fprintf(buffer, "Error: El comando 'LOGOUT' incluye parámetros no asociados.\n")
		return
	}
	Usuario.Logout(buffer.(*bytes.Buffer))
}

func Funcion_Rep(entrada string, buffer io.Writer) {
	fs := flag.NewFlagSet("rep", flag.ExitOnError)
	nombre := fs.String("name", "", "Nombre")
	ruta := fs.String("path", "full", "Ruta")
	ID := fs.String("id", "", "IDParticion")
	path_file_ls := fs.String("path_file_l", "", "PathFile")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(entrada, -1)

	for _, match := range matches {
		nombreFlag := match[1]
		valorFlag := strings.ToLower(match[2])

		valorFlag = strings.Trim(valorFlag, "\"")

		switch nombreFlag {
		case "name", "path", "id", "path_file_l":
			fs.Set(nombreFlag, valorFlag)
		default:
			fmt.Fprintf(buffer, "Error: El comando 'REP' incluye parámetros no asociados.\n")
			return
		}
	}
	Reportes.Rep(*nombre, *ruta, *ID, *path_file_ls, buffer.(*bytes.Buffer))
}

// Creacion de comando l disk para mostrar los discos montados
func Funcion_ldisk(writer io.Writer) {

	ManejadorDisco.Ldisk(writer.(*bytes.Buffer))
}

func Funcion_cat(params string, linea string, writer io.Writer) {
	//fs := flag.NewFlagSet("cat", flag.ExitOnError)

	// Usaremos un mapa para almacenar los archivos
	files := make(map[int]string)

	// Encontrar la flag en el input
	matches := re.FindAllStringSubmatch(params, -1)

	// Process the input
	for _, match := range matches {
		flagName := match[1]                   // match[1]: Captura y guarda el nombre del flag (por ejemplo, "file1", "file2", etc.)
		flagValue := strings.ToLower(match[2]) //strings.ToLower(match[2]): Captura y guarda el valor del flag, asegurándose de que esté en minúsculas

		flagValue = strings.Trim(flagValue, "\"")

		// Si el flagName empieza con "file" y es seguido por un número
		if strings.HasPrefix(flagName, "file") {
			// Extraer el número después de "file"
			fileNumber, err := strconv.Atoi(strings.TrimPrefix(flagName, "file"))
			if err != nil {
				fmt.Fprintln(writer, "Error: Nombre de archivo inválido")

				return
			}

			if flagValue == "" {
				fmt.Fprintln(writer, "Error: parametro -file"+string(fileNumber)+" no contiene ninguna ruta")
			}

			files[fileNumber] = flagValue
		} else {
			fmt.Println("Error: Flag not found")
		}
	}

	// Convertir el mapa a un slice ordenado
	var orderedFiles []string
	for i := 1; i <= len(files); i++ {
		if file, exists := files[i]; exists {
			orderedFiles = append(orderedFiles, file)
		} else {
			fmt.Fprintln(writer, "Error: Falta un archivo en la secuencia")
			return
		}
	}

	if len(orderedFiles) == 0 {
		fmt.Fprintln(writer, "Error: No se encontraron archivos")
		return
	}

	// Llamar a la función para manejar los archivos en orden
	ManejadorArchivo.Cat(orderedFiles, writer.(*bytes.Buffer))
}
func Funcion_Rmusr(params string, writer io.Writer) {
	fs := flag.NewFlagSet("rmusr", flag.ExitOnError)
	user := fs.String("user", "", "user")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		flagName := strings.ToLower(match[1])
		flagValue := strings.ToLower(match[2])

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "user":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
		}
	}

	if *user == "" {
		fmt.Fprintf(writer, "Error: Name is obligatory")
		return
	}

	//RootUser.Rmusr(*user)
	ManejoRoot.Rmusr(*user, writer.(*bytes.Buffer))
}

func Funcion_Rmgrp(params string, writer io.Writer) {
	fs := flag.NewFlagSet("rmgrp", flag.ExitOnError)
	user := fs.String("name", "", "Name")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		flagName := strings.ToLower(match[1])
		flagValue := strings.ToLower(match[2])

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "name":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
		}
	}

	if *user == "" {
		fmt.Fprintf(writer, "Error: Name is obligatory")
		return
	}

	//RootUser.Rmusr(*user)
	ManejoRoot.Rmgrp(*user, writer.(*bytes.Buffer))
}

func Funcion_Mkgrp(params string, writer io.Writer) {
	fs := flag.NewFlagSet("mkgrp", flag.ExitOnError)
	name := fs.String("name", "", "Nombre")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		flagName := strings.ToLower(match[1])
		flagValue := strings.ToLower(match[2])

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "name":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
		}
	}

	if *name == "" {
		fmt.Fprintf(writer, "Error: Name is obligatory")
		return
	}

	ManejoRoot.Mkgrp(*name, writer.(*bytes.Buffer))

}

func Funcion_Chgrp(params string, writer io.Writer) {
	fs := flag.NewFlagSet("chgrp", flag.ExitOnError)
	user := fs.String("user", "", "Usuario")
	grp := fs.String("grp", "", "Grupo")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		flagName := strings.ToLower(match[1])
		flagValue := strings.ToLower(match[2])

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "user", "grp":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
		}
	}

	if *user == "" {
		fmt.Fprintf(writer, "Error: Name is obligatory")
		return
	}

	if *grp == "" {
		fmt.Fprintf(writer, "Error: Name is obligatory")
		return
	}

	//User.Chgrp(*user, *grp)
	ManejoRoot.Chgrp(*user, *grp, writer.(*bytes.Buffer))
}

func Funcion_Mkusr(params string, writer io.Writer) {
	fs := flag.NewFlagSet("mkusr", flag.ExitOnError)
	user := fs.String("user", "", "Usuario")
	pass := fs.String("pass", "", "Contraseña")
	grp := fs.String("grp", "", "Grupo")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		flagName := strings.ToLower(match[1])
		flagValue := strings.ToLower(match[2])

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "user", "pass", "grp":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
		}
	}

	if *user == "" {
		fmt.Fprintf(writer, "Error: Name is obligatory")
		return
	}

	if *pass == "" {
		fmt.Fprintf(writer, "Error: Name is obligatory")
		return
	}

	if *grp == "" {
		fmt.Fprintf(writer, "Error: Name is obligatory")
		return
	}

	//User.Mkusr(*user, *pass, *grp)
	//ManejoRoot.Mkusr(*user, *pass, *grp)
	ManejoRoot.Mkusr(*user, *pass, *grp, writer.(*bytes.Buffer))
}
