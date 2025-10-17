package handlers

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/EngSteven/pso-http-server/internal/server"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

// MandelbrotHandler genera el mapa de iteraciones de Mandelbrot.
// Query params:
//   width (int, default 256, max 2000)
//   height (int, default 256, max 2000)
//   max_iter (int, default 1000, max 10000)
//   outfile (optional) -> si se pasa, se escribe un archivo PGM/PNG en disco con ese nombre
func MandelbrotHandler(req *types.Request) *types.Response {
	// parse params con defaults y validación
	const (
		defaultW    = 256
		defaultH    = 256
		defaultIter = 1000
		maxW        = 2000
		maxH        = 2000
		maxIterCap  = 10000
	)

	w := defaultW
	h := defaultH
	it := defaultIter
	var err error

	if s := req.Query.Get("width"); s != "" {
		if w, err = strconv.Atoi(s); err != nil || w <= 0 {
			return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"invalid width"}`))
		}
	}
	if s := req.Query.Get("height"); s != "" {
		if h, err = strconv.Atoi(s); err != nil || h <= 0 {
			return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"invalid height"}`))
		}
	}
	if s := req.Query.Get("max_iter"); s != "" {
		if it, err = strconv.Atoi(s); err != nil || it <= 0 {
			return server.NewResponse(400, "Bad Request", "application/json", []byte(`{"error":"invalid max_iter"}`))
		}
	}

	// límites
	if w > maxW || h > maxH {
		return server.NewResponse(400, "Bad Request", "application/json", []byte(fmt.Sprintf(`{"error":"max width/height %d/%d"}`, maxW, maxH)))
	}
	if it > maxIterCap {
		return server.NewResponse(400, "Bad Request", "application/json", []byte(fmt.Sprintf(`{"error":"max_iter limit %d"}`, maxIterCap)))
	}

	outfile := req.Query.Get("outfile")

	// job que hará el cálculo. Usa la firma del JobFunc del pool: func(cancelCh <-chan struct{}) *types.Response
	job := func(cancelCh <-chan struct{}) *types.Response {
		start := time.Now()

		// mapa de iteraciones: height filas, width columnas
		iterations := make([][]int, h)
		for y := 0; y < h; y++ {
			// crear fila
			row := make([]int, w)
			for x := 0; x < w; x++ {
				// cancelación cooperativa: chequea cada píxel
				select {
				case <-cancelCh:
					return server.NewResponse(500, "Canceled", "application/json", []byte(`{"error":"cancelled"}`))
				default:
				}

				// Mapear (x,y) a coordenadas complejas (estándar área: -2.5..1.0 por real, -1.0..1.0 por imag)
				cr := -2.5 + (float64(x)/float64(w))*(1.0+2.5) // -2.5 .. 1.0
				ci := -1.0 + (float64(y)/float64(h))*2.0      // -1.0 .. 1.0

				// iteración clásica mandelbrot
				var zr, zi float64 = 0.0, 0.0
				var iter int
				for iter = 0; iter < it; iter++ {
					// z = z^2 + c
					// z^2: (zr^2 - zi^2) + 2*zr*zi i
					zr2 := zr*zr - zi*zi + cr
					zi = 2*zr*zi + ci
					zr = zr2
					if zr*zr+zi*zi > 4.0 {
						break
					}
				}
				row[x] = iter
			}
			iterations[y] = row
		}

		elapsed := time.Since(start).Milliseconds()

		// Si se pidió outfile: generar PGM/PNG y devolver file+checksum en JSON
		if outfile != "" {
			// generar imagen en escala de grises (0..255) según iter/max_iter
			img := image.NewGray(image.Rect(0, 0, w, h))
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					iter := iterations[y][x]
					var v uint8
					if iter >= it {
						v = 0 // dentro del conjunto -> negro (0)
					} else {
						// mapea iter a 1..255
						v = uint8(1 + (255*(iter))/it)
					}
					img.SetGray(x, y, color.Gray{Y: v})
				}
			}

			// decidir formato por extensión simple: .pgm -> PGM (write raw), else PNG
			if len(outfile) >= 4 && outfile[len(outfile)-4:] == ".pgm" {
				if err := writePGM(outfile, img); err != nil {
					msg := fmt.Sprintf(`{"error":"failed to write pgm: %v"}`, err)
					return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
				}
			} else {
				// default PNG
				f, err := os.Create(outfile)
				if err != nil {
					msg := fmt.Sprintf(`{"error":"failed to create file: %v"}`, err)
					return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
				}
				defer f.Close()
				if err := png.Encode(f, img); err != nil {
					msg := fmt.Sprintf(`{"error":"failed to encode png: %v"}`, err)
					return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
				}
			}

			// calcular checksum simple del archivo
			sum, err := fileChecksum(outfile)
			if err != nil {
				msg := fmt.Sprintf(`{"error":"wrote file but failed checksum: %v"}`, err)
				return server.NewResponse(500, "Internal Server Error", "application/json", []byte(msg))
			}

			respObj := map[string]interface{}{
				"file":       outfile,
				"width":      w,
				"height":     h,
				"max_iter":   it,
				"elapsed_ms": elapsed,
				"checksum":   sum,
			}
			b, _ := json.MarshalIndent(respObj, "", "  ")
			return server.NewResponse(200, "OK", "application/json", b)
		}

		// Por defecto: devolver JSON con la matriz de iteraciones
		respObj := map[string]interface{}{
			"width":      w,
			"height":     h,
			"max_iter":   it,
			"elapsed_ms": elapsed,
			"iterations": iterations,
		}
		b, _ := json.MarshalIndent(respObj, "", "  ")
		return server.NewResponse(200, "OK", "application/json", b)
	}

	// enviar al pool si existe (con prioridad normal)
	p := workers.GetPool("mandelbrot")
	if p == nil {
		// fallback inline
		return job(make(chan struct{}))
	}
	resp, err := p.SubmitAndWait(job, workers.PriorityNormal)
	if err != nil {
		// queue full o timeout
		return server.NewResponse(503, "Service Unavailable", "application/json", []byte(`{"error":"queue full or timeout"}`))
	}
	return resp
}

// writePGM escribe una imagen Gray en formato P5 (PGM binario)
func writePGM(path string, img *image.Gray) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	// cabecera PGM P5
	header := fmt.Sprintf("P5\n%d %d\n255\n", w, h)
	if _, err := io.WriteString(f, header); err != nil {
		return err
	}
	// datos (row-major)
	if _, err := f.Write(img.Pix); err != nil {
		return err
	}
	return nil
}

// fileChecksum calcula un hash FNV del archivo (retorna hex string)
func fileChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := fnv.New64a()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum64()), nil
}
