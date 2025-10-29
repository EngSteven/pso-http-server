/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: sortfile.go
Descripcion: Ordena archivos de numeros enteros usando algoritmos quick sort o merge sort
con metricas detalladas de tiempo de lectura, ordenamiento y escritura.
*/

package algorithms

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// SortFile lee archivo de numeros enteros, los ordena y guarda resultado.
// Entrada: name (string) - ruta del archivo a ordenar
//
//	algo (string) - algoritmo de ordenamiento ("quick" o "merge")
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *types.Response - respuesta HTTP con metricas de ordenamiento o error
// Descripcion: Procesa archivo de enteros usando quicksort (Go nativo) o mergesort
//
//	(implementacion personalizada). Lee archivo con buffer de 10MB,
//	ordena en memoria y guarda en archivo .sorted con metricas detalladas.
func SortFile(name, algo string, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	// === VALIDACION DE PARAMETROS ===
	// Verificar que se proporcione nombre de archivo
	if name == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: name"}`))
	}

	// === ALGORITMO POR DEFECTO ===
	// Usar quicksort si no se especifica algoritmo
	if algo == "" {
		algo = "quick"
	}

	// === VERIFICACION DE CANCELACION INICIAL ===
	// Comprobar cancelacion antes de procesar archivo
	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled before start"}`))
	default:
	}

	// === LECTURA DE ARCHIVO ===
	// Abrir y leer archivo de numeros enteros
	readStart := time.Now()
	file, err := os.Open(name)
	if err != nil {
		msg := fmt.Sprintf(`{"error":"failed to open file: %v"}`, err)
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}
	defer file.Close()

	// === CONFIGURACION DE BUFFER ===
	// Configurar scanner con buffer de 10MB para archivos grandes
	var numbers []int
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // buffer de hasta 10MB por línea

	// === PROCESAMIENTO LINEA POR LINEA ===
	// Leer y convertir cada linea a entero
	for scanner.Scan() {
		// === VERIFICACION DE CANCELACION POR LINEA ===
		// Permitir cancelacion durante lectura intensiva
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(`{"error":"operation cancelled while reading"}`))
		default:
		}

		// === PROCESAMIENTO DE LINEA ===
		// Limpiar y convertir linea a numero entero
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // Saltar lineas vacias
		}
		if n, err := strconv.Atoi(line); err == nil {
			numbers = append(numbers, n)
		}
	}
	readTime := time.Since(readStart)

	// === VALIDACION DE DATOS ===
	// Verificar que se encontraron numeros validos en el archivo
	if len(numbers) == 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"no valid numeric data found in file"}`))
	}

	// === ORDENAMIENTO DE DATOS ===
	// Aplicar algoritmo de ordenamiento especificado
	sortStart := time.Now()
	switch algo {
	case "merge":
		// === MERGE SORT PERSONALIZADO ===
		// Usar implementacion propia de merge sort con cancelacion
		numbers = mergeSort(numbers, cancelCh)
	case "quick":
		// === QUICK SORT NATIVO ===
		// Usar implementacion nativa de Go (quicksort optimizado)
		sort.Ints(numbers)
	default:
		// === ALGORITMO INVALIDO ===
		// Retornar error si el algoritmo no es reconocido
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid algorithm: must be merge or quick"}`))
	}
	sortTime := time.Since(sortStart)

	// === ESCRITURA DE ARCHIVO RESULTADO ===
	// Crear archivo de salida con numeros ordenados
	writeStart := time.Now()
	outName := fmt.Sprintf("%s.sorted", name)
	outFile, err := os.Create(outName)
	if err != nil {
		msg := fmt.Sprintf(`{"error":"failed to create output file: %v"}`, err)
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}
	defer outFile.Close()

	// === ESCRITURA CON BUFFER ===
	// Usar buffer para escritura eficiente
	writer := bufio.NewWriter(outFile)
	for _, n := range numbers {
		// === VERIFICACION DE CANCELACION POR NUMERO ===
		// Permitir cancelacion durante escritura
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(`{"error":"operation cancelled while writing"}`))
		default:
			// === ESCRITURA DE NUMERO ===
			// Escribir cada numero en una linea
			fmt.Fprintln(writer, n)
		}
	}
	// === FLUSH DEL BUFFER ===
	// Asegurar que todos los datos se escriban al disco
	writer.Flush()
	writeTime := time.Since(writeStart)

	// === CONSTRUCCION DE RESPUESTA CON METRICAS ===
	// Preparar respuesta JSON con metricas detalladas de rendimiento
	data, _ := json.MarshalIndent(map[string]interface{}{
		"file":        name,                             // Archivo de entrada procesado
		"output_file": outName,                          // Archivo de salida generado
		"algorithm":   algo,                             // Algoritmo de ordenamiento usado
		"count":       len(numbers),                     // Cantidad de numeros procesados
		"elapsed_ms":  time.Since(start).Milliseconds(), // Tiempo total
		"read_ms":     readTime.Milliseconds(),          // Tiempo de lectura
		"sort_ms":     sortTime.Milliseconds(),          // Tiempo de ordenamiento
		"write_ms":    writeTime.Milliseconds(),         // Tiempo de escritura
	}, "", "  ")

	// === RESPUESTA EXITOSA ===
	// Retornar resultado con codigo 200 y metricas JSON
	return server.NewResponse(200, "OK", "application/json", data)
}

// mergeSort implementa algoritmo de ordenamiento merge sort recursivo.
// Entrada: arr ([]int) - array de enteros a ordenar
//
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: []int - array ordenado o parcialmente ordenado si se cancelo
// Descripcion: Implementa merge sort dividiendo recursivamente el array hasta
//
//	elementos individuales y combinandolos ordenadamente. Maneja
//	cancelacion en cada nivel de recursion para interrumpir proceso.
func mergeSort(arr []int, cancelCh <-chan struct{}) []int {
	// === CASO BASE ===
	// Arrays de 0 o 1 elemento ya estan ordenados
	if len(arr) <= 1 {
		return arr
	}

	// === CALCULO DE PUNTO MEDIO ===
	// Dividir array en dos mitades
	mid := len(arr) / 2

	// === VERIFICACION DE CANCELACION ===
	// Permitir cancelacion en cada nivel de recursion
	select {
	case <-cancelCh:
		return arr // Retornar array sin ordenar si se cancela
	default:
	}

	// === DIVISION RECURSIVA ===
	// Ordenar recursivamente cada mitad
	left := mergeSort(arr[:mid], cancelCh)
	right := mergeSort(arr[mid:], cancelCh)

	// === COMBINACION ===
	// Combinar las dos mitades ordenadas
	return merge(left, right)
}

// merge combina dos arrays ordenados en uno solo manteniendo el orden.
// Entrada: left ([]int) - primer array ordenado
//
//	right ([]int) - segundo array ordenado
//
// Salida: []int - array resultante con elementos de ambos arrays ordenados
// Descripcion: Combina dos arrays ya ordenados en uno solo preservando orden.
//
//	Utiliza dos indices para recorrer ambos arrays simultaneamente
//	comparando elementos y agregando el menor al resultado.
func merge(left, right []int) []int {
	// === INICIALIZACION DE RESULTADO ===
	// Crear array resultado con capacidad total de ambos arrays
	result := make([]int, 0, len(left)+len(right))
	i, j := 0, 0

	// === COMPARACION Y COMBINACION ===
	// Recorrer ambos arrays comparando elementos
	for i < len(left) && j < len(right) {
		// === SELECCION DEL MENOR ELEMENTO ===
		// Agregar el menor elemento al resultado
		if left[i] <= right[j] {
			result = append(result, left[i])
			i++
		} else {
			result = append(result, right[j])
			j++
		}
	}

	// === AGREGADO DE ELEMENTOS RESTANTES ===
	// Agregar elementos que quedaron en cada array
	result = append(result, left[i:]...)  // Resto del array izquierdo
	result = append(result, right[j:]...) // Resto del array derecho

	return result
}
