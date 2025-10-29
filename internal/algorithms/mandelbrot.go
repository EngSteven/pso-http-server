/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: mandelbrot.go
Descripcion: Genera el conjunto de Mandelbrot calculando iteraciones para cada pixel
con opcion de guardar resultado en formato PGM y soporte para cancelacion.
*/

package algorithms

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
)

// Mandelbrot genera fractal del conjunto de Mandelbrot con parametros configurables.
// Entrada: width (int) - ancho de la imagen en pixeles (debe ser > 0)
//
//	height (int) - altura de la imagen en pixeles (debe ser > 0)
//	maxIter (int) - numero maximo de iteraciones por pixel (debe ser > 0)
//	saveFile (bool) - si guardar resultado como archivo PGM
//	cancelCh (<-chan struct{}) - canal para cancelacion de operacion
//
// Salida: *types.Response - respuesta HTTP con matriz de iteraciones o error
// Descripcion: Genera fractal de Mandelbrot calculando iteraciones para cada pixel
//
//	en region compleja [-2.5,1.0] x [-1.5,1.5]. Opcionalmente guarda
//	resultado como imagen PGM. Maneja cancelacion durante calculo.
func Mandelbrot(width, height, maxIter int, saveFile bool, cancelCh <-chan struct{}) *types.Response {
	start := time.Now()

	// === VALIDACION DE PARAMETROS ===
	// Verificar que dimensiones y iteraciones sean positivas
	if width <= 0 || height <= 0 || maxIter <= 0 {
		return server.NewResponse(400, "Bad Request", "application/json",
			[]byte(`{"error":"invalid parameters: width, height, max_iter must be > 0"}`))
	}

	// === VERIFICACION DE CANCELACION TEMPRANA ===
	// Chequear cancelacion antes de iniciar calculo intensivo
	select {
	case <-cancelCh:
		return server.NewResponse(499, "Client Closed Request", "application/json",
			[]byte(`{"error":"operation cancelled before start"}`))
	default:
	}

	// === CONFIGURACION DEL ESPACIO COMPLEJO ===
	// Definir region del plano complejo a visualizar
	xMin, xMax := -2.5, 1.0               // Rango horizontal
	yMin, yMax := -1.5, 1.5               // Rango vertical
	dx := (xMax - xMin) / float64(width)  // Incremento por pixel en X
	dy := (yMax - yMin) / float64(height) // Incremento por pixel en Y

	// === INICIALIZACION DE MATRIZ DE RESULTADOS ===
	// Crear matriz para almacenar iteraciones por pixel
	grid := make([][]int, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]int, width)
	}

	// === GENERACION DEL FRACTAL ===
	// Calcular iteraciones para cada pixel de la imagen
	for py := 0; py < height; py++ {
		// === VERIFICACION DE CANCELACION POR FILA ===
		// Permitir cancelacion entre cada fila procesada
		select {
		case <-cancelCh:
			return server.NewResponse(499, "Client Closed Request", "application/json",
				[]byte(fmt.Sprintf(`{"error":"cancelled at row %d"}`, py)))
		default:
			// === PROCESAMIENTO DE PIXELS EN FILA ===
			// Calcular valor Mandelbrot para cada pixel de la fila
			for px := 0; px < width; px++ {
				// === MAPEO DE PIXEL A COORDENADA COMPLEJA ===
				// Convertir coordenadas de pixel a coordenadas complejas
				x0 := xMin + float64(px)*dx
				y0 := yMin + float64(py)*dy

				// === INICIALIZACION DE ITERACION MANDELBROT ===
				// Comenzar iteracion con z = 0
				x, y := 0.0, 0.0
				iter := 0

				// === ITERACION MANDELBROT ===
				// z = z² + c hasta que |z| > 2 o se alcance maxIter
				for x*x+y*y <= 4 && iter < maxIter {
					// === VERIFICACION DE CANCELACION POR PIXEL ===
					// Permitir cancelacion durante calculos intensivos
					select {
					case <-cancelCh:
						return server.NewResponse(499, "Client Closed Request", "application/json",
							[]byte(fmt.Sprintf(`{"error":"cancelled at pixel (%d,%d)"}`, px, py)))
					default:
						// === FORMULA DE MANDELBROT ===
						// z_new = z² + c donde c = x0 + iy0
						xTemp := x*x - y*y + x0
						y = 2*x*y + y0
						x = xTemp
						iter++
					}
				}
				// === ALMACENAMIENTO DE RESULTADO ===
				// Guardar numero de iteraciones para este pixel
				grid[py][px] = iter
			}
		}
	}

	// === GENERACION DE ARCHIVO DE IMAGEN ===
	// Crear archivo de resultado si se solicito guardado
	filename := ""
	if saveFile {
		// === NOMBRADO DE ARCHIVO ===
		// Generar nombre descriptivo con parametros del fractal
		filename = fmt.Sprintf("mandelbrot_%dx%d_%d.pgm", width, height, maxIter)

		// === GUARDADO EN FORMATO PGM ===
		// Llamar funcion auxiliar para guardar en formato estandar
		savePGM(filename, grid, maxIter)
	}

	// === CONSTRUCCION DE RESPUESTA JSON ===
	// Preparar respuesta con metadatos y datos del fractal
	data, _ := json.MarshalIndent(map[string]interface{}{
		"width":      width,                            // Ancho de la imagen generada
		"height":     height,                           // Alto de la imagen generada
		"max_iter":   maxIter,                          // Numero maximo de iteraciones usado
		"saved_file": filename,                         // Nombre del archivo guardado (si aplica)
		"elapsed_ms": time.Since(start).Milliseconds(), // Tiempo total de procesamiento
		"iterations": grid,                             // Matriz completa de iteraciones por pixel
	}, "", "  ")

	// === RESPUESTA EXITOSA ===
	// Retornar resultado con codigo 200 y datos JSON
	return server.NewResponse(200, "OK", "application/json", data)
}

// savePGM guarda la matriz de iteraciones en formato PGM (escala de grises).
// Entrada: filename (string) - nombre del archivo a crear
//
//	grid ([][]int) - matriz de valores de iteraciones por pixel
//	maxIter (int) - valor maximo de iteraciones para normalizacion
//
// Salida: ninguna
// Descripcion: Guarda matriz de iteraciones como imagen PGM (Portable GrayMap).
//
//	Formato P2 ASCII con valores de 0 a maxIter representando intensidad
//	de gris. Compatible con visualizadores de imagenes estandar.
func savePGM(filename string, grid [][]int, maxIter int) {
	// === CREACION DE ARCHIVO ===
	// Crear archivo PGM en disco para escribir imagen
	f, err := os.Create(filename)
	if err != nil {
		return
	}
	defer f.Close()

	// === OBTENCION DE DIMENSIONES ===
	// Extraer dimensiones de la matriz de datos
	height := len(grid)
	width := len(grid[0])

	// === ESCRITURA DE HEADER PGM ===
	// Formato P2 (ASCII gris), dimensiones y valor maximo
	fmt.Fprintf(f, "P2\n%d %d\n%d\n", width, height, maxIter)

	// === ESCRITURA DE DATOS DE PIXEL ===
	// Convertir matriz de iteraciones a valores de gris
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// === ESCRITURA DE VALOR DE INTENSIDAD ===
			// Escribir numero de iteraciones como intensidad de gris
			fmt.Fprintf(f, "%d ", grid[y][x])
		}
		// === SALTO DE LINEA POR FILA ===
		// Terminar cada fila de pixels con nueva linea
		fmt.Fprintln(f)
	}
}
