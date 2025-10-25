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

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("============================================================")
	fmt.Println(" Generador de datasets de prueba (PSO_PY01b)")
	fmt.Println("============================================================")

	os.MkdirAll("data", 0755)

	generateNumbersFile("data/data_big.txt", 5_000_000)
	generateWordsFile("data/data_words.txt", 5_000_000)

	fmt.Println("Archivos generados en ./data/")
	fmt.Println(" - data_big.txt   (números aleatorios)")
	fmt.Println(" - data_words.txt (palabras aleatorias)")
	fmt.Println("============================================================")
}

// generateNumbersFile crea un archivo con N números aleatorios.
func generateNumbersFile(path string, lines int) {
	fmt.Printf("Generando %s con %d líneas...\n", path, lines)
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for i := 0; i < lines; i++ {
		fmt.Fprintf(f, "%d\n", rand.Intn(1_000_000))
	}
}

// generateWordsFile crea un archivo con N líneas de texto aleatorio.
func generateWordsFile(path string, lines int) {
	fmt.Printf("Generando %s con %d líneas...\n", path, lines)
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	words := []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta", "omega"}
	for i := 0; i < lines; i++ {
		line := []string{}
		for j := 0; j < 5; j++ {
			line = append(line, words[rand.Intn(len(words))])
		}
		f.WriteString(strings.Join(line, " ") + "\n")
	}
}
