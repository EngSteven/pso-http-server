/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: gen_dataset.go
Descripcion: Script utilitario para generar datasets grandes de prueba
con numeros aleatorios y texto para testing de algoritmos IO-intensivos.
*/

// scripts/gen_dataset.go
//
// Script utilitario para generar datasets grandes de prueba.
// Uso:
//   go run scripts/gen_dataset.go
//
// Genera:
//   data/data_big.txt    (~50 MB de números aleatorios)
//   data/data_words.txt  (~50 MB de texto aleatorio)

package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

// main ejecuta la generacion completa de datasets de prueba.
// Entrada: ninguna
// Salida: ninguna (void)
// Descripcion: Punto de entrada del script. Inicializa generador aleatorio,
//
//	crea directorio data/ y genera dos archivos de prueba:
//	numeros aleatorios (5M lineas) y texto aleatorio (5M lineas).
//	Muestra progreso y confirmacion al usuario.
func main() {
	// === INICIALIZACION DEL GENERADOR ALEATORIO ===
	// Sembrar con timestamp para variabilidad
	rand.Seed(time.Now().UnixNano())

	fmt.Println("============================================================")
	fmt.Println(" Generador de datasets de prueba (PSO_PY01b)")
	fmt.Println("============================================================")

	// === CREACION DEL DIRECTORIO DE DATOS ===
	// Asegurar que directorio data/ existe
	os.MkdirAll("data", 0755)

	// === GENERACION DE DATASETS ===
	// Crear archivos de números y palabras para testing
	generateNumbersFile("data/data_big.txt", 5_000_000)
	generateWordsFile("data/data_words.txt", 5_000_000)

	// === CONFIRMACION AL USUARIO ===
	fmt.Println("Archivos generados en ./data/")
	fmt.Println(" - data_big.txt   (números aleatorios)")
	fmt.Println(" - data_words.txt (palabras aleatorias)")
	fmt.Println("============================================================")
}

// generateNumbersFile crea archivo con N numeros aleatorios para pruebas.
// Entrada: path (string) - ruta del archivo a crear
//
//	lines (int) - cantidad de numeros a generar
//
// Salida: ninguna (void)
// Descripcion: Genera archivo de texto con numeros aleatorios del 0 al 999,999.
//
//	Cada numero en linea separada. Usado para testing de algoritmos
//	de ordenamiento y procesamiento numerico. Maneja errores fatales.
func generateNumbersFile(path string, lines int) {
	fmt.Printf("Generando %s con %d líneas...\n", path, lines)
	// === CREACION DEL ARCHIVO ===
	// Abrir archivo para escritura con manejo de errores
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// === GENERACION DE NUMEROS ALEATORIOS ===
	// Escribir números del 0 al 999,999 en líneas separadas
	for i := 0; i < lines; i++ {
		fmt.Fprintf(f, "%d\n", rand.Intn(1_000_000))
	}
}

// generateWordsFile crea archivo con N lineas de texto aleatorio para pruebas.
// Entrada: path (string) - ruta del archivo a crear
//
//	lines (int) - cantidad de lineas de texto a generar
//
// Salida: ninguna (void)
// Descripcion: Genera archivo con lineas de 5 palabras aleatorias cada una.
//
//	Palabras seleccionadas de vocabulario griego predefinido.
//	Usado para testing de algoritmos de procesamiento de texto,
//	busqueda y conteo de palabras. Maneja errores fatales.
func generateWordsFile(path string, lines int) {
	fmt.Printf("Generando %s con %d líneas...\n", path, lines)
	// === CREACION DEL ARCHIVO ===
	// Abrir archivo para escritura con manejo de errores
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// === VOCABULARIO GRIEGO PREDEFINIDO ===
	// Lista de palabras para generación aleatoria
	words := []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta", "omega"}
	for i := 0; i < lines; i++ {
		// === CONSTRUCCION DE LINEA ===
		// Crear línea con 5 palabras aleatorias
		line := []string{}
		for j := 0; j < 5; j++ {
			line = append(line, words[rand.Intn(len(words))])
		}
		// === ESCRITURA DE LINEA ===
		// Unir palabras con espacios y agregar salto de línea
		f.WriteString(strings.Join(line, " ") + "\n")
	}
}
