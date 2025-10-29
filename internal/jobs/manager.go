/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: manager.go
Descripcion: Gestor principal de jobs asincronos con colas por prioridad,
persistencia en journal y dispatcher para ejecutar tareas en pools de workers.
*/

package jobs

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/EngSteven/pso-http-server/internal/algorithms"
	"github.com/EngSteven/pso-http-server/internal/types"
	"github.com/EngSteven/pso-http-server/internal/util"
	"github.com/EngSteven/pso-http-server/internal/workers"
)

var (
	ErrJobNotFound  = errors.New("job not found")
	ErrJobQueueFull = errors.New("job manager queue full")
	ErrJobCancelled = errors.New("job cancelled")
)

// Default timeouts per command (ms)
var defaultTimeouts = map[string]int{
	"isprime":    5000,
	"factor":     8000,
	"pi":         15000,
	"matrixmul":  7000,
	"mandelbrot": 20000,
	"fibonacci":  3000,
	"createfile": 2000,
}

// timeoutForCommand obtiene timeout configurado para comando especifico.
// Entrada: cmd (string) - nombre del comando para buscar timeout
// Salida: int - timeout en milisegundos para el comando
// Descripcion: Busca timeout configurado en mapa defaultTimeouts para comando
//
//	especifico. Si no encuentra configuracion, retorna valor por
//	defecto de 5000ms como fallback para cualquier comando.
func timeoutForCommand(cmd string) int {
	// === BUSQUEDA DE TIMEOUT CONFIGURADO ===
	// Consultar mapa de timeouts por comando específico
	if v, ok := defaultTimeouts[cmd]; ok {
		return v
	}

	// === TIMEOUT POR DEFECTO ===
	// Retornar fallback de 5000ms para comandos sin configuración
	return 5000
}

// JobManager administra colas de jobs por prioridad y dispatch a pools.
// Mantiene persistencia en journal y canales de comunicacion para resultados.
type JobManager struct {
	mu sync.Mutex

	// queues per priority
	highQ, normalQ, lowQ chan *JobMeta

	// metadata
	store         map[string]*JobMeta
	resChMap      map[string]chan *types.Response
	cancelChMap   map[string]chan struct{}
	journalPath   string
	journalFile   *os.File
	stop          chan struct{}
	wg            sync.WaitGroup
	maxQueueTotal int
}

// NewJobManager crea e inicializa un nuevo gestor de jobs.
// Entrada: journalPath (string) - ruta del archivo journal para persistencia
//
//	qDepthPerPriority (int) - profundidad de cada cola de prioridad
//	maxQueueTotal (int) - limite total de jobs en todas las colas
//
// Salida: (*JobManager, error) - gestor inicializado o error de configuracion
// Descripcion: Crea nuevo JobManager con colas por prioridad (high, normal, low),
//
//	configura persistencia en journal JSONL, inicia dispatcher en
//	goroutine y rehidrata jobs desde journal existente si aplica.
func NewJobManager(journalPath string, qDepthPerPriority int, maxQueueTotal int) (*JobManager, error) {
	// === INICIALIZACION DE ESTRUCTURA ===
	// Crear JobManager con colas por prioridad y mapas de estado
	j := &JobManager{
		highQ:         make(chan *JobMeta, qDepthPerPriority),
		normalQ:       make(chan *JobMeta, qDepthPerPriority),
		lowQ:          make(chan *JobMeta, qDepthPerPriority),
		store:         make(map[string]*JobMeta),
		resChMap:      make(map[string]chan *types.Response),
		cancelChMap:   make(map[string]chan struct{}),
		journalPath:   journalPath,
		stop:          make(chan struct{}),
		maxQueueTotal: maxQueueTotal,
	}

	// === APERTURA DE ARCHIVO JOURNAL ===
	// Configurar persistencia para recuperación tras reinicio
	f, err := os.OpenFile(journalPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("open journal: %w", err)
	}
	j.journalFile = f

	// === REHIDRATACION DESDE JOURNAL ===
	// Recuperar jobs existentes desde archivo de persistencia
	if err := j.rehydrate(); err != nil {
		fmt.Printf("warning: unable to rehydrate jobs: %v\n", err)
	}

	// === INICIO DEL DISPATCHER ===
	// Lanzar goroutine para procesar colas de jobs
	j.wg.Add(1)
	go j.dispatcher()

	return j, nil
}

// rehydrate carga jobs desde journal para recuperar estado tras reinicio.
// Entrada: ninguna (metodo de JobManager)
// Salida: error - error de lectura o parsing, nil si es exitoso
// Descripcion: Lee archivo journal linea por linea, parsea cada entrada JSON
//
//	como JobMeta y reconstruye el estado de jobs en memoria.
//	Permite recuperacion tras reinicio del servidor.
func (j *JobManager) rehydrate() error {
	// === POSICIONAMIENTO EN INICIO DE ARCHIVO ===
	// Mover cursor al inicio para lectura completa del journal
	f := j.journalFile
	_, err := f.Seek(0, 0)
	if err != nil {
		return err
	}

	// === LECTURA LINEA POR LINEA ===
	// Procesar cada entrada JSON del journal
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// === DESERIALIZACION DE METADATA ===
		// Convertir JSON a JobMeta y agregarlo al store
		var meta JobMeta
		if err := json.Unmarshal(scanner.Bytes(), &meta); err == nil {
			j.store[meta.ID] = &meta
		}
	}
	return nil
}

// appendToJournal escribe metadata de job al archivo journal para persistencia.
// Entrada: meta (*JobMeta) - metadata del job a persistir
// Salida: ninguna
// Descripcion: Serializa JobMeta a JSON y lo escribe al archivo journal seguido
//
//	de newline. Ignora errores de escritura silenciosamente.
//	Permite recuperacion de jobs tras reinicio del servidor.
func (j *JobManager) appendToJournal(meta *JobMeta) {
	// === VERIFICACION DE ARCHIVO DISPONIBLE ===
	// Asegurar que el archivo journal este abierto
	if j.journalFile == nil {
		return
	}

	// === SERIALIZACION A JSON ===
	// Convertir metadata del job a formato JSON
	line, _ := json.Marshal(meta)

	// === ESCRITURA AL JOURNAL ===
	// Agregar línea JSON seguida de newline y sincronizar
	j.journalFile.Write(append(line, '\n'))
	j.journalFile.Sync()
}

// Submit crea metadata de job y lo encola segun prioridad especificada.
// Entrada: command (string) - comando/algoritmo a ejecutar
//
//	params (map[string]string) - parametros del comando
//	priority (Priority) - prioridad de ejecucion (high/normal/low)
//
// Salida: (string, error) - ID unico del job creado o error de validacion
// Descripcion: Crea JobMeta con ID unico, valida limites de cola total,
//
//	configura canales de respuesta y cancelacion, encola segun
//	prioridad y persiste en journal. Retorna ID para tracking.
func (j *JobManager) Submit(command string, params map[string]string, priority Priority) (string, error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	// === VERIFICACION DE LIMITE DE COLA ===
	// Calcular total de jobs en todas las colas y validar límite
	total := len(j.highQ) + len(j.normalQ) + len(j.lowQ)
	if total >= j.maxQueueTotal {
		// === BACKPRESSURE - RECHAZO CON RETRY ===
		// Rechazar job y sugerir tiempo de reintento
		retryAfter := workers.DefaultTimeoutFor(command)
		return "", fmt.Errorf("queue full: retry_after_ms=%d", retryAfter)
	}

	// === CREACION DE METADATA DEL JOB ===
	// Generar ID único y construir estructura JobMeta
	id := util.NewRequestID()
	meta := &JobMeta{
		ID:        id,
		Command:   command,
		Params:    params,
		Priority:  priority,
		Status:    StatusQueued,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		TimeoutMs: timeoutForCommand(command),
	}

	// === ALMACENAMIENTO Y PERSISTENCIA ===
	// Guardar en store y persistir en journal
	j.store[id] = meta
	j.appendToJournal(meta)

	// === ENCOLADO POR PRIORIDAD ===
	// Intentar encolar respetando prioridad con fallback a otras colas
	switch priority {
	case PriorityHigh:
		// === PRIORIDAD ALTA - ALTA > NORMAL > BAJA ===
		select {
		case j.highQ <- meta:
		default:
			select {
			case j.normalQ <- meta:
			default:
				select {
				case j.lowQ <- meta:
				default:
					delete(j.store, id)
					return "", ErrJobQueueFull
				}
			}
		}
	case PriorityNormal:
		// === PRIORIDAD NORMAL - NORMAL > BAJA > ALTA ===
		select {
		case j.normalQ <- meta:
		default:
			select {
			case j.lowQ <- meta:
			default:
				select {
				case j.highQ <- meta:
				default:
					delete(j.store, id)
					return "", ErrJobQueueFull
				}
			}
		}
	default:
		// === PRIORIDAD BAJA - BAJA > NORMAL ===
		select {
		case j.lowQ <- meta:
		default:
			select {
			case j.normalQ <- meta:
			default:
				delete(j.store, id)
				return "", ErrJobQueueFull
			}
		}
	}
	return id, nil
}

// dispatcher consume colas de jobs y los despacha a pools de workers.
// Entrada: ninguna (metodo de JobManager ejecutado en goroutine)
// Salida: ninguna
// Descripcion: Loop infinito que consume jobs de colas por prioridad (high > normal > low),
//
//	los envuelve en JobFunc, los envia a pools correspondientes y maneja
//	respuestas/timeouts. Se ejecuta en goroutine separada hasta recibir stop.
func (j *JobManager) dispatcher() {
	defer j.wg.Done()
	for {
		select {
		case <-j.stop:
			return
		default:
			var meta *JobMeta
			// === SELECCION ADAPTIVA DE PRIORIDAD ===
			// Algoritmo probabilístico para balancear carga y prioridad
			switch {
			case rand.Intn(100) < 50:
				// === ESTRATEGIA A: PRIORIDAD CON FALLBACK ===
				// 50% probabilidad: alta → normal → baja
				select {
				case meta = <-j.highQ:
				default:
					select {
					case meta = <-j.normalQ:
					default:
						select {
						case meta = <-j.lowQ:
						default:
							time.Sleep(50 * time.Millisecond)
							continue
						}
					}
				}
			default:
				// === ESTRATEGIA B: BALANCE NORMAL-ALTA ===
				// 50% probabilidad: normal → alta → baja
				select {
				case meta = <-j.normalQ:
				default:
					select {
					case meta = <-j.highQ:
					default:
						select {
						case meta = <-j.lowQ:
						default:
							time.Sleep(50 * time.Millisecond)
							continue
						}
					}
				}
			}

			if meta == nil {
				continue
			}

			// === ACTUALIZACION DE ESTADO A RUNNING ===
			// Marcar job como en ejecución y persistir
			j.mu.Lock()
			meta.Status = StatusRunning
			meta.UpdatedAt = time.Now()
			j.appendToJournal(meta)
			j.mu.Unlock()

			// === DELEGACION AL WORKER POOL ===
			// Intentar obtener pool específico para el comando
			pool := workers.GetPool(meta.Command)
			if pool == nil {
				// === EJECUCION INLINE - SIN POOL ===
				// Pool no disponible, ejecutar directamente
				res := j.executeCommandInline(meta)
				j.updateJobResult(meta, res)
				continue
			}

			// === ENVOLVER JOB PARA WORKER POOL ===
			// Crear función wrapper con metadata del job
			jobFn := j.wrapJob(meta)

			// === ENCOLADO EN WORKER POOL ===
			// Delegar a pool especializado con prioridad normal
			jobID, pResCh, cancelCh, err := pool.Enqueue(jobFn, workers.PriorityNormal)
			if err != nil {
				// === ERROR DE ENCOLADO - REVERTIR ESTADO ===
				j.mu.Lock()
				meta.Status = StatusQueued
				meta.UpdatedAt = time.Now()
				meta.Error = "pool full"
				j.appendToJournal(meta)
				j.mu.Unlock()
				// === REINTENTO AUTOMATICO CON DELAY ===
				// Reencolar job después de breve espera
				go func(m *JobMeta) {
					time.Sleep(200 * time.Millisecond)
					select {
					case j.normalQ <- m:
					default:
					}
				}(meta)
				continue
			}

			// === REGISTRO DE CANALES DE COMUNICACION ===
			// Almacenar canales para comunicación con worker pool
			j.mu.Lock()
			j.resChMap[jobID] = pResCh
			j.cancelChMap[jobID] = cancelCh
			j.mu.Unlock()

			// === DELEGACION DE MONITOREO ===
			// Lanzar goroutine para esperar resultado con timeout
			go j.waitForResult(meta, jobID, pResCh, cancelCh)
		}
	}
}

// waitForResult espera resultado de job con timeout y maneja actualizacion de estado.
// Entrada: meta (*JobMeta) - metadata del job
//
//	id (string) - ID del job
//	pch (chan *types.Response) - canal de respuesta del pool
//	cancelCh (chan struct{}) - canal de cancelacion
//
// Salida: ninguna
// Descripcion: Goroutine que espera resultado del job o timeout, actualiza metadata
//
//	con resultado/error, limpia canales de mapas y persiste estado final.
func (j *JobManager) waitForResult(meta *JobMeta, id string, pch chan *types.Response, cancelCh chan struct{}) {
	// === CONFIGURACION DE TIMEOUT ===
	// Calcular timeout basado en metadata del job
	timeout := time.Duration(meta.TimeoutMs) * time.Millisecond
	select {
	case res := <-pch:
		// === RESULTADO EXITOSO RECIBIDO ===
		// Actualizar job con resultado del worker pool
		j.updateJobResult(meta, res)
	case <-time.After(timeout):
		// === TIMEOUT - CANCELACION Y LIMPIEZA ===
		// Enviar señal de cancelación al worker
		close(cancelCh)
		j.mu.Lock()
		meta.Status = StatusTimeout
		meta.Error = fmt.Sprintf("timed out after %d ms", meta.TimeoutMs)
		meta.UpdatedAt = time.Now()
		j.appendToJournal(meta)
		// === LIMPIEZA DE CANALES ===
		// Remover canales de mapas de seguimiento
		delete(j.resChMap, id)
		delete(j.cancelChMap, id)
		j.mu.Unlock()
	}
}

/*Algoritmos de los jobs*/

// wrapJob envuelve metadata de job en JobFunc ejecutable por workers.
// Entrada: meta (*JobMeta) - metadata del job con comando y parametros
// Salida: workers.JobFunc - funcion ejecutable que retorna Response
// Descripcion: Convierte JobMeta en funcion ejecutable por workers, mapea
//
//	comando a algoritmo correspondiente, extrae y convierte parametros.
//	Incluye manejo de cancelacion y fallback para comandos desconocidos.
func (j *JobManager) wrapJob(meta *JobMeta) workers.JobFunc {
	return func(cancelCh <-chan struct{}) *types.Response {
		// === DISPATCH POR COMANDO ===
		// Router para mapear comando a algoritmo específico
		switch meta.Command {
		// === ALGORITMOS MATEMATICOS ===
		case "fibonacci":
			// === FIBONACCI - CONVERSION DE PARAMETROS ===
			n, _ := strconv.Atoi(meta.Params["num"])
			return algorithms.CalculateFibonacci(n, cancelCh)

		// === ALGORITMOS DE ARCHIVOS ===
		case "createfile":
			// === CREAR ARCHIVO - PARAMETROS CON REPETICION ===
			name := meta.Params["name"]
			content := meta.Params["content"]
			repeat := 1
			if r, ok := meta.Params["repeat"]; ok {
				if v, err := strconv.Atoi(r); err == nil && v > 0 {
					repeat = v
				}
			}
			return algorithms.CreateFile(name, content, repeat, cancelCh)

		case "deletefile":
			// === ELIMINAR ARCHIVO ===
			name := meta.Params["name"]
			return algorithms.DeleteFile(name, cancelCh)

		// === ALGORITMOS DE TEXTO ===
		case "reverse":
			// === REVERSO DE TEXTO ===
			text := meta.Params["text"]
			return algorithms.ReverseText(text, cancelCh)

		case "toupper":
			// === CONVERTIR A MAYUSCULAS ===
			text := meta.Params["text"]
			return algorithms.ToUpper(text, cancelCh)

		// === ALGORITMOS DE GENERACION ===
		case "random":
			// === NUMEROS ALEATORIOS - PARAMETROS NUMERICOS ===
			count, _ := strconv.Atoi(meta.Params["count"])
			min, _ := strconv.Atoi(meta.Params["min"])
			max, _ := strconv.Atoi(meta.Params["max"])
			return algorithms.GenerateRandom(count, min, max, cancelCh)

		case "timestamp":
			// === TIMESTAMP ACTUAL ===
			return algorithms.GetTimestamp(cancelCh)

		// === ALGORITMOS DE HASH ===
		case "hash":
			// === HASH DE TEXTO ===
			text := meta.Params["text"]
			return algorithms.HashText(text, cancelCh)

		// === ALGORITMOS DE SIMULACION ===
		case "simulate":
			// === SIMULACION DE TRABAJO ===
			seconds, _ := strconv.Atoi(meta.Params["seconds"])
			taskName := meta.Params["task"]
			return algorithms.SimulateWork(seconds, taskName, cancelCh)

		case "sleep":
			// === PAUSA/SLEEP ===
			seconds, _ := strconv.Atoi(meta.Params["seconds"])
			return algorithms.Sleep(seconds, cancelCh)

		case "loadtest":
			// === PRUEBA DE CARGA ===
			taskCount, _ := strconv.Atoi(meta.Params["tasks"])
			sleepSeconds, _ := strconv.Atoi(meta.Params["sleep"])
			return algorithms.LoadTest(taskCount, sleepSeconds, cancelCh)

		//-------------------------------CPU Bound---------------------------

		// === ALGORITMOS INTENSIVOS EN CPU ===
		case "isprime":
			// === PRUEBA DE PRIMALIDAD ===
			n, _ := strconv.ParseInt(meta.Params["n"], 10, 64)
			method := meta.Params["method"]
			return algorithms.IsPrime(n, method, cancelCh)

		case "factor":
			// === FACTORIZACION ===
			n, _ := strconv.ParseInt(meta.Params["n"], 10, 64)
			return algorithms.Factorize(n, cancelCh)

		case "pi":
			// === CALCULO DE PI ===
			digits, _ := strconv.Atoi(meta.Params["digits"])
			return algorithms.CalculatePi(digits, cancelCh)

		case "mandelbrot":
			// === FRACTAL MANDELBROT - PARAMETROS DE RENDERING ===
			width, _ := strconv.Atoi(meta.Params["width"])
			height, _ := strconv.Atoi(meta.Params["height"])
			maxIter, _ := strconv.Atoi(meta.Params["max_iter"])
			save := meta.Params["save"] == "true" || meta.Params["save"] == "1"
			return algorithms.Mandelbrot(width, height, maxIter, save, cancelCh)

		case "matrixmul":
			// === MULTIPLICACION DE MATRICES ===
			size, _ := strconv.Atoi(meta.Params["size"])
			seed, _ := strconv.ParseInt(meta.Params["seed"], 10, 64)
			return algorithms.MatrixMultiply(size, seed, cancelCh)

		//-------------------------------IO Bound---------------------------

		// === ALGORITMOS INTENSIVOS EN IO ===
		case "sortfile":
			// === ORDENAMIENTO DE ARCHIVO ===
			name := meta.Params["name"]
			algo := meta.Params["algo"]
			return algorithms.SortFile(name, algo, cancelCh)

		case "wordcount":
			// === CONTEO DE PALABRAS ===
			name := meta.Params["name"]
			return algorithms.WordCount(name, cancelCh)

		case "grep":
			// === BUSQUEDA EN ARCHIVO ===
			name := meta.Params["name"]
			pattern := meta.Params["pattern"]
			return algorithms.Grep(name, pattern, cancelCh)

		case "hashfile":
			// === HASH DE ARCHIVO ===
			name := meta.Params["name"]
			algo := meta.Params["algo"]
			return algorithms.HashFile(name, algo, cancelCh)

		case "compress":
			// === COMPRESION DE ARCHIVO ===
			name := meta.Params["name"]
			codec := meta.Params["codec"]
			return algorithms.CompressFile(name, codec, cancelCh)

		default:
			// === COMANDO DESCONOCIDO - ERROR ===
			return j.newResponse(400, "Bad Request", "application/json", []byte(`{"error":"unknown command"}`))
		}
	}
}

// updateJobResult actualiza metadata de job con resultado de ejecucion.
// Entrada: meta (*JobMeta) - metadata del job a actualizar
//
//	res (*types.Response) - respuesta del algoritmo ejecutado
//
// Salida: ninguna
// Descripcion: Actualiza estado del job a done/error segun codigo de respuesta,
//
//	serializa respuesta completa a JSON, actualiza timestamp y
//	persiste cambios en journal para recuperacion.
func (j *JobManager) updateJobResult(meta *JobMeta, res *types.Response) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if res != nil {
		// === RESULTADO EXITOSO - SERIALIZACION ===
		// Convertir respuesta completa a JSON para almacenamiento
		b, _ := json.Marshal(res)
		meta.Result = string(b)
		meta.Status = StatusDone
	} else {
		// === RESULTADO NULO - ERROR ===
		// Manejar caso de respuesta nula como error
		meta.Error = "nil response"
		meta.Status = StatusError
	}
	// === ACTUALIZACION DE TIMESTAMP Y PERSISTENCIA ===
	// Marcar tiempo de finalización y persistir en journal
	meta.UpdatedAt = time.Now()
	j.appendToJournal(meta)
	// === LIMPIEZA DE CANALES ===
	// Remover canales de comunicación del mapa
	delete(j.resChMap, meta.ID)
	delete(j.cancelChMap, meta.ID)
}

// newResponse crea nueva instancia de Response con parametros especificados.
// Entrada: statusCode (int) - codigo de estado HTTP
//
//	status (string) - texto del estado HTTP
//	ctype (string) - tipo de contenido
//	body ([]byte) - cuerpo de la respuesta
//
// Salida: *types.Response - respuesta HTTP construida
// Descripcion: Factory function para crear respuestas HTTP con headers basicos.
//
//	Establece Content-Type segun parametro y construye estructura Response.
func (j *JobManager) newResponse(statusCode int, status, ctype string, body []byte) *types.Response {
	// === CONSTRUCCION DE RESPUESTA HTTP ===
	// Factory para crear respuesta con headers básicos
	return &types.Response{
		StatusCode: statusCode,
		StatusText: status,
		Headers:    map[string]string{"Content-Type": ctype},
		Body:       body,
	}
}

// executeCommandInline ejecuta comando directamente sin usar pools de workers.
// Entrada: meta (*JobMeta) - metadata del job con comando y parametros
// Salida: *types.Response - respuesta del comando ejecutado
// Descripcion: Fallback para ejecutar comandos cuando pools no estan disponibles.
//
//	Implementa algoritmos basicos inline (fibonacci, reverse, etc.)
//	sin cancelacion ni timeouts. Usado como backup del sistema.
func (j *JobManager) executeCommandInline(meta *JobMeta) *types.Response {
	// === DISPATCH DE COMANDOS INLINE ===
	// Ejecutar algoritmos básicos sin pools ni cancelación
	switch meta.Command {
	case "fibonacci":
		// === FIBONACCI INLINE - IMPLEMENTACION SIMPLE ===
		// Calcular serie completa sin optimizaciones
		n, _ := strconv.Atoi(meta.Params["num"])
		series := make([]int, n)
		if n > 0 {
			series[0] = 0
		}
		if n > 1 {
			series[1] = 1
			// === CALCULO ITERATIVO ===
			// Generar serie usando suma de términos anteriores
			for i := 2; i < n; i++ {
				series[i] = series[i-1] + series[i-2]
			}
		}
		// === SERIALIZACION Y RESPUESTA ===
		// Convertir resultado a JSON y crear respuesta
		data, _ := json.Marshal(map[string]interface{}{"n": n, "series": series})
		return j.newResponse(200, "OK", "application/json", data)
	default:
		// === COMANDO NO SOPORTADO ===
		// Solo fibonacci está implementado inline
		return j.newResponse(400, "Bad Request", "application/json", []byte(`{"error":"unknown command"}`))
	}
}

func (j *JobManager) GetMeta(id string) (*JobMeta, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	// === BUSQUEDA EN STORE ===
	// Verificar existencia del job en el mapa de metadata
	if meta, ok := j.store[id]; ok {
		// === COPIA DEFENSIVA ===
		// Crear copia para evitar modificaciones concurrentes
		c := *meta
		return &c, nil
	}
	// === JOB NO ENCONTRADO ===
	return nil, ErrJobNotFound
}

func (j *JobManager) Cancel(id string) error {
	j.mu.Lock()
	// === VERIFICACION DE EXISTENCIA ===
	// Buscar job en store de metadata
	meta, ok := j.store[id]
	if !ok {
		j.mu.Unlock()
		return ErrJobNotFound
	}
	// === VERIFICACION DE ESTADO CANCELABLE ===
	// Jobs finalizados no pueden ser cancelados
	if meta.Status == StatusDone || meta.Status == StatusError || meta.Status == StatusCanceled {
		j.mu.Unlock()
		return ErrJobCancelled
	}
	// === CANCELACION DE JOB EN COLA ===
	// Job aún no ejecutándose, cancelar directamente
	if meta.Status == StatusQueued {
		meta.Status = StatusCanceled
		meta.UpdatedAt = time.Now()
		meta.Error = "canceled before dispatch"
		j.appendToJournal(meta)
		j.mu.Unlock()
		return nil
	}
	// === CANCELACION DE JOB EN EJECUCION ===
	// Enviar señal de cancelación al worker
	if cancelCh, ok := j.cancelChMap[id]; ok {
		close(cancelCh)
		meta.Status = StatusCanceled
		meta.UpdatedAt = time.Now()
		j.appendToJournal(meta)
		j.mu.Unlock()
		return nil
	}
	// === JOB NO CANCELABLE ===
	// Estado inconsistente o ya procesado
	j.mu.Unlock()
	return ErrJobCancelled
}
