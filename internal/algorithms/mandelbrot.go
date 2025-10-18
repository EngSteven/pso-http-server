package algorithms

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// Mandelbrot genera un mapa de iteraciones del conjunto de Mandelbrot.
// Si saveFile=true, guarda un archivo PGM en disco.
func Mandelbrot(width, height, maxIter int, saveFile bool, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	if width <= 0 || height <= 0 || maxIter <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameters: width, height, max_iter must be > 0"}`))
	}

	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled before start"}`))
	default:
	}

	// Parámetros visuales básicos
	xMin, xMax := -2.5, 1.0
	yMin, yMax := -1.5, 1.5
	dx := (xMax - xMin) / float64(width)
	dy := (yMax - yMin) / float64(height)

	grid := make([][]int, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]int, width)
	}

	// Generar el fractal
	for py := 0; py < height; py++ {
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(fmt.Sprintf(`{"error":"cancelled at row %d"}`, py)))
		default:
			for px := 0; px < width; px++ {
				x0 := xMin + float64(px)*dx
				y0 := yMin + float64(py)*dy
				x, y := 0.0, 0.0
				iter := 0
				for x*x+y*y <= 4 && iter < maxIter {
					select {
					case <-cancelCh:
						return server.NewResponse(499, "Client Closed Request", "application/json",
							[]byte(fmt.Sprintf(`{"error":"cancelled at pixel (%d,%d)"}`, px, py)))
					default:
						xTemp := x*x - y*y + x0
						y = 2*x*y + y0
						x = xTemp
						iter++
					}
				}
				grid[py][px] = iter
			}
		}
	}

	filename := ""
	if saveFile {
		filename = fmt.Sprintf("mandelbrot_%dx%d_%d.pgm", width, height, maxIter)
		savePGM(filename, grid, maxIter)
	}

	data, _ := json.MarshalIndent(map[string]interface{}{
		"width":       width,
		"height":      height,
		"max_iter":    maxIter,
		"saved_file":  filename,
		"elapsed_ms":  time.Since(start).Milliseconds(),
		"iterations":  grid,
	}, "", "  ")

	return server.NewResponse(200, "OK", "application/json", data)
}

// savePGM guarda la matriz en formato PGM (escala de grises)
func savePGM(filename string, grid [][]int, maxIter int) {
	f, err := os.Create(filename)
	if err != nil {
		return
	}
	defer f.Close()

	height := len(grid)
	width := len(grid[0])

	fmt.Fprintf(f, "P2\n%d %d\n%d\n", width, height, maxIter)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			fmt.Fprintf(f, "%d ", grid[y][x])
		}
		fmt.Fprintln(f)
	}
}
