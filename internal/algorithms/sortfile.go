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

// SortFile lee un archivo con números enteros (uno por línea), los ordena y guarda en un nuevo archivo.
// Parámetros:
//   name  = nombre del archivo a leer
//   algo  = algoritmo de ordenamiento ("quick" o "merge")
func SortFile(name, algo string, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	if name == "" {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"missing parameter: name"}`))
	}
	if algo == "" {
		algo = "quick"
	}

	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled before start"}`))
	default:
	}

	// --- 1️⃣ Leer archivo ---
	readStart := time.Now()
	file, err := os.Open(name)
	if err != nil {
		msg := fmt.Sprintf(`{"error":"failed to open file: %v"}`, err)
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}
	defer file.Close()

	var numbers []int
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // buffer de hasta 10MB por línea
	for scanner.Scan() {
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(`{"error":"operation cancelled while reading"}`))
		default:
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if n, err := strconv.Atoi(line); err == nil {
			numbers = append(numbers, n)
		}
	}
	readTime := time.Since(readStart)

	// --- 2️⃣ Ordenar ---
	sortStart := time.Now()
	switch algo {
	case "merge":
		numbers = mergeSort(numbers, cancelCh)
	case "quick":
		sort.Ints(numbers)
	default:
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid algorithm: must be merge or quick"}`))
	}
	sortTime := time.Since(sortStart)

	// --- 3️⃣ Escribir archivo de salida ---
	writeStart := time.Now()
	outName := fmt.Sprintf("%s.sorted", name)
	outFile, err := os.Create(outName)
	if err != nil {
		msg := fmt.Sprintf(`{"error":"failed to create output file: %v"}`, err)
		return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	for _, n := range numbers {
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(`{"error":"operation cancelled while writing"}`))
		default:
			fmt.Fprintln(writer, n)
		}
	}
	writer.Flush()
	writeTime := time.Since(writeStart)

	// --- 4️⃣ Métricas ---
	data, _ := json.MarshalIndent(map[string]interface{}{
		"file":           name,
		"output_file":    outName,
		"algorithm":      algo,
		"count":          len(numbers),
		"elapsed_ms":     time.Since(start).Milliseconds(),
		"read_ms":        readTime.Milliseconds(),
		"sort_ms":        sortTime.Milliseconds(),
		"write_ms":       writeTime.Milliseconds(),
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}

// --- MergeSort con soporte para cancelación ---
func mergeSort(arr []int, cancelCh <-chan struct{}) []int {
	if len(arr) <= 1 {
		return arr
	}
	mid := len(arr) / 2
	select {
	case <-cancelCh:
		return arr
	default:
	}
	left := mergeSort(arr[:mid], cancelCh)
	right := mergeSort(arr[mid:], cancelCh)
	return merge(left, right)
}

func merge(left, right []int) []int {
	result := make([]int, 0, len(left)+len(right))
	i, j := 0, 0
	for i < len(left) && j < len(right) {
		if left[i] <= right[j] {
			result = append(result, left[i])
			i++
		} else {
			result = append(result, right[j])
			j++
		}
	}
	result = append(result, left[i:]...)
	result = append(result, right[j:]...)
	return result
}
