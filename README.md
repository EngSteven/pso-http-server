# PSO HTTP Server (Proyecto 1 - Sistemas Operativos)

Servidor HTTP/1.0 concurrente escrito en Go para el curso de Principios de Sistemas Operativos.

## Autores

Steven Sequeira Araya 

Jefferson Salas Cordero

## Documentación de la API:

https://documenter.getpostman.com/view/37666062/2sB3QQK8Eh

## Video explicación:

https://youtu.be/a47ozGUjfz4

## Compilación y Ejecución

Importante, para compilar y ejecutar el proyecto debe hacerlo desde una terminal que soporte Linux

1. Desde el root del proyecto ejecute para compilar el servidor:

```bash 
make build
```

2. Desde el root del proyecto ejecute para correr el servidor:

```bash 
make run
```

## Configuración del Servidor

El servidor HTTP permite ajustar su comportamiento mediante **variables de entorno**.  
Todas las configuraciones son **opcionales** y pueden **combinarse libremente**.  
Si no se especifican, el servidor utilizará los **valores por defecto** indicados en la tabla.

| Variable                 | Descripción                                                               | Valor por defecto |
|---------------------------|---------------------------------------------------------------------------|-------------------|
| `PORT`                    | Puerto HTTP de escucha del servidor                                       | `8080`            |
| `DATA_DIR` *(opcional)*   | Ruta donde se guardan archivos, datasets y el journal de jobs             | `data/`           |
| `QUEUE_DEPTH`             | Profundidad global de las colas del *Job Manager*                         | `50`              |
| `MAX_TOTAL`               | Máximo total de trabajos concurrentes aceptados por el servidor           | `150`             |
| `WORKERS_<COMANDO>`       | Número de workers para un comando específico (ej. `WORKERS_PI=4`)         | Valor definido en código |
| `QUEUE_<COMANDO>`         | Profundidad de la cola del comando (ej. `QUEUE_SORTFILE=3`)               | Valor definido en código |
| `TIMEOUT_<COMANDO>`       | Tiempo máximo de ejecución en milisegundos (ej. `TIMEOUT_HASHFILE=8000`)  | Valor definido en código |

>  Cualquier comando definido en el servidor (CPU-bound o IO-bound) puede personalizarse:
> - `/fibonacci`, `/isprime`, `/pi`, `/matrixmul`, `/mandelbrot`
> - `/sortfile`, `/wordcount`, `/grep`, `/hashfile`, `/compress`
> - `/reverse`, `/toupper`, `/createfile`, `/deletefile`, etc.

---

### Ejemplos de uso

1. Cambiar el puerto del servidor:
```bash
export PORT=9090
make run
```

2. Personalizar el pool del comando pi:
```bash
export WORKERS_PI=4
export QUEUE_PI=3
export TIMEOUT_PI=8000
make run
```

3. Ajustar la profundidad global de colas y el máximo total:
```bash
export QUEUE_DEPTH=100
export MAX_TOTAL=300
make run
```

4. Redefinir la ruta base de datos:
```bash
export DATA_DIR=/tmp/pso_data
make run
```


5. Personalizar múltiples comandos a la vez:
```bash
export WORKERS_SORTFILE=2
export QUEUE_SORTFILE=5
export TIMEOUT_SORTFILE=7000
export WORKERS_ISPRIME=3
make run
```

## Correr todas las pruebas

1. Genere los archivos grandes para las pruebas de rendimiento:

```bash
go run scripts/gen_dataset.go
```

2. Desde el root del proyecto ejecute para correr las pruebas

```bash 
make test-all
```

3. Opción extra para incrementar la carga en las pruebas de rendimiento

```bash 
PERF_LEVEL=heavy make test-all
```

## Arquitectura 

### Estructura General del Proyecto

```bash
PSO-HTTP-SERVER/
│
├── cmd/               → Punto de entrada principal (main.go)
├── internal/
│   ├── algorithms/    → Implementaciones de algoritmos (CPU-bound e IO-bound)
│   ├── config/        → Carga de parámetros y variables de entorno
│   ├── handlers/      → Controladores de rutas HTTP (fibonacci, reverse, status, etc.)
│   ├── jobs/          → Job Manager y colas asincrónicas para tareas largas
│   ├── metrics/       → Registro y cálculo de métricas del sistema
│   ├── router/        → Enrutamiento manual HTTP/1.0 (GET → handler)
│   ├── server/        → Núcleo del servidor (socket, accept, request parsing, response)
│   ├── types/         → Definiciones comunes (Request, Response, Job, WorkerInfo)
│   ├── util/          → Funciones auxiliares (logs, validaciones, helpers)
│   └── workers/       → Pools de goroutines/threads por comando
│
├── data/              → Archivos generados o usados por IO-bound (/sortfile, /compress, etc.)
├── scripts/           → Scripts de prueba y benchmarking
├── test/              → Pruebas unitarias e integración (≥90% coverage)
├── go.mod / go.sum    → Dependencias del proyecto
└── Makefile           → Comandos de build, test y run automatizados

```

### Capas Principales

#### 1. Capa de Servidor (`internal/server`)
Encargada de la gestión de sockets y protocolo HTTP/1.0.  
Implementa:
- Recepción de conexiones (`listen`, `accept`).
- Parsing manual de peticiones.
- Envío de respuestas con cabeceras y códigos HTTP.
- Multiplexación de clientes concurrentes.

---

#### 2. Capa de Enrutamiento (`internal/router`)
- Determina a qué *handler* se debe redirigir cada solicitud.
- Implementa el mapeo de rutas `/help`, `/status`, `/fibonacci`, etc.
- Se comunica con `handlers` y mantiene la lógica central del flujo de solicitudes.

---

#### 3. Capa de Handlers (`internal/handlers`)
Contiene los controladores para cada endpoint HTTP.  
Cada handler:
- Valida los parámetros.
- Llama al algoritmo o job correspondiente.
- Retorna respuestas JSON estandarizadas.


---

#### 4. Capa de Workers (`internal/workers`)
- Implementa *pools* de goroutines por tipo de comando.
- Asegura concurrencia controlada y evita bloqueos.
- Gestiona la ejecución directa de tareas cortas (CPU o IO).

---

#### 5. Capa de Jobs (`internal/jobs`)
- Maneja trabajos asincrónicos de larga duración.
- Implementa los endpoints `/jobs/submit`, `/jobs/status`, `/jobs/result`, `/jobs/cancel`.
- Gestiona colas, prioridades, progreso y cancelaciones.
- Incluye persistencia temporal para reanudación tras un reinicio.

---

#### 6. Capa de Algoritmos (`internal/algorithms`)
- Contiene las funciones de procesamiento real:
  - CPU-bound: `isprime`, `factor`, `pi`, `matrixmul`, etc.
  - IO-bound: `sortfile`, `grep`, `compress`, `hashfile`, etc.
- Cada algoritmo retorna estructuras JSON con métricas de ejecución.

---

#### 7. Capa de Métricas (`internal/metrics`)
- Registra tiempos de ejecución, tiempos en cola y uso de workers.
- Expone datos a través del endpoint `/metrics`.
- Permite análisis de rendimiento (p50/p95/p99).

---

#### 8. Capa de Utilidades (`internal/util` y `internal/config`)
- `util`: Funciones auxiliares para logs, validación, formatos, etc.  
- `config`: Maneja variables de entorno y parámetros CLI (`--port`, `--workers.isprime`, `--queue.size`, etc.).

---

### Diseño Modular

El sistema sigue un patrón **modular desacoplado**, donde cada paquete puede evolucionar sin afectar al resto.  
Esta organización facilita:
- Reutilización de componentes.
- Pruebas unitarias independientes.
- Extensión de nuevos comandos o algoritmos sin modificar el núcleo.

---

### Flujo General de Ejecución

1. **Inicio:** `cmd/main.go` inicializa configuración, workers y servidor.  
2. **Servidor:** `server` escucha conexiones y construye `types.Request`.  
3. **Router:** Enruta la petición al handler correcto.  
4. **Handler:** Ejecuta el algoritmo o job.  
5. **Workers/Jobs:** Procesan concurrentemente según el tipo.  
6. **Métricas:** Registran tiempos y estados.  
7. **Respuesta:** `server` envía JSON con código HTTP adecuado.

---

### Conclusión

La arquitectura de **PSO HTTP Server** refleja los principios de:
- **Separación de responsabilidades.**
- **Concurrencia controlada.**
- **Escalabilidad horizontal por comando.**
- **Observabilidad y trazabilidad.**

Con esta estructura, el sistema es extensible, testeable y cumple con los requerimientos técnicos y funcionales del proyecto.


## Anexos

### Salida esperada al ejecutar las pruebas automatizadas

```bash
====================================================
Running all tests with coverage
====================================================
go test ./tests -v -count=1 \
    -coverpkg=github.com/EngSteven/pso-http-server/internal/algorithms \
    -coverprofile=reports/coverage.out

============================================================
TEST SUITE — Algoritmos del Servidor HTTP (PSO_PY01b)
Descripción: Pruebas unitarias por categorías (básicos, CPU, IO)
============================================================
=== RUN   TestBenchmark_Profiles
    integration_test.go:61: 
        ============================================================
    integration_test.go:62: SETUP: Iniciando servidor embebido para pruebas de integración
    integration_test.go:63: ============================================================
2025/11/02 09:26:11 Servidor escuchando en :8080
2025/11/02 09:26:11 [d600aecd-605d-4104-b7f9-69810ec285de] GET /status -> 200 (OK) [PID=1211] [1.44 ms]
    integration_test.go:71: Servidor embebido disponible en: http://localhost:8080
    benchmark_test.go:184: 
        ============================================================
    benchmark_test.go:185: BENCHMARK DE CARGA — Iniciando perfiles (low, medium, high)
    benchmark_test.go:186: ============================================================
    benchmark_test.go:89: 
        --- [RUNNING] Perfil de carga: low (10 solicitudes, 2 concurrentes) ---
2025/11/02 09:26:11 [87f7aba3-3254-4870-9864-b25e7cd19c53] GET /pi -> 200 (OK) [PID=1211] [67.30 ms]
2025/11/02 09:26:11 [3c76985a-b948-45a2-b825-eb7fb8febe0a] GET /pi -> 200 (OK) [PID=1211] [119.73 ms]
2025/11/02 09:26:11 [1a633272-26da-4c37-8cf7-1b477502a67e] GET /pi -> 200 (OK) [PID=1211] [98.84 ms]
2025/11/02 09:26:12 [6c34e6cb-2f91-460e-9993-fb21c985df08] GET /pi -> 200 (OK) [PID=1211] [96.64 ms]
2025/11/02 09:26:12 [d98cabfd-62ff-426d-9eef-fd4e8500a712] GET /pi -> 200 (OK) [PID=1211] [102.38 ms]
2025/11/02 09:26:12 [1556bdc7-a79f-40db-8e58-4fdb0ee5951f] GET /pi -> 200 (OK) [PID=1211] [103.11 ms]
2025/11/02 09:26:12 [af8eccb4-791e-47ac-af5f-04e3b74864dd] GET /pi -> 200 (OK) [PID=1211] [89.09 ms]
2025/11/02 09:26:12 [4044300e-b296-41c7-a83d-f25246c3f62c] GET /pi -> 200 (OK) [PID=1211] [73.34 ms]
2025/11/02 09:26:12 [0b563b49-bcf6-496f-b511-dfb308f816d1] GET /pi -> 200 (OK) [PID=1211] [63.32 ms]
2025/11/02 09:26:12 [45f95227-4555-4fad-8635-5f48b6867f8e] GET /pi -> 200 (OK) [PID=1211] [53.35 ms]
    benchmark_test.go:156: 
        [RESULTADOS] Perfil: low
          Requests totales: 10 (Errores: 0)
          Duración total:   0.45s
          Throughput:       22.17 req/s
          Latencia p50:     90.00 ms
          Latencia p95:     103.00 ms
          Latencia p99:     103.00 ms
    benchmark_test.go:166: [OK] Perfil 'low' completado correctamente
    benchmark_test.go:89: 
        --- [RUNNING] Perfil de carga: medium (50 solicitudes, 5 concurrentes) ---
2025/11/02 09:26:12 [abab9e70-4abb-4e05-a34c-ce5a2f028fe0] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.20 ms]
2025/11/02 09:26:12 [c0406bca-0d40-49b1-951d-288ff2352248] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.32 ms]
2025/11/02 09:26:12 [064da146-c1b4-44f0-aa51-7f26c61d0c66] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.16 ms]
2025/11/02 09:26:12 [475b7736-db9b-45b8-bf17-4dcb2b075610] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [216067f5-a5bc-4f44-b389-9dffbc2c36fa] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.15 ms]
2025/11/02 09:26:12 [7d87df68-e8b4-4a9a-b36a-0c948b31c93d] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.07 ms]
2025/11/02 09:26:12 [1b351c7d-8e0b-4fa3-af01-28f0d1829c2c] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.14 ms]
2025/11/02 09:26:12 [20478a70-9674-43b1-9f0a-5ca92446dba7] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.17 ms]
2025/11/02 09:26:12 [00c74d82-aef9-4737-8250-f411d43ad8ff] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [fb9b37a3-b1c7-4eaf-bf9d-6d3ce33d250c] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.09 ms]
2025/11/02 09:26:12 [8ac8fef9-0285-4109-b78b-77ab95f95908] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.07 ms]
2025/11/02 09:26:12 [ed7e66d2-3990-4768-89c1-3d7370519294] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.07 ms]
2025/11/02 09:26:12 [8e8da9c4-ce7b-4065-9020-b4a623379339] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.07 ms]
2025/11/02 09:26:12 [3b6f9ae9-94f6-4873-8e8e-d9525cda7c50] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.14 ms]
2025/11/02 09:26:12 [0fa0acb0-dca8-4276-9561-31792991d04d] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.12 ms]
2025/11/02 09:26:12 [9951d4dc-b747-4d4c-a2af-627b13a83ea5] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.18 ms]
2025/11/02 09:26:12 [007d29c2-e152-4db1-a09b-92e316649765] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.50 ms]
2025/11/02 09:26:12 [3c38ced9-42d4-48f6-9496-e15f4e94213d] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.18 ms]
2025/11/02 09:26:12 [a0c531f9-0e34-4901-8673-45a06ed3ff8a] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.15 ms]
2025/11/02 09:26:12 [b8306425-5164-45d6-be63-317fe6d3a5de] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.15 ms]
2025/11/02 09:26:12 [271a7c4d-3944-4a1d-8537-fa634126d365] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.04 ms]
2025/11/02 09:26:12 [b1c82af0-f9f4-4d5b-bfe4-71d472cfb402] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.07 ms]
2025/11/02 09:26:12 [002dfbb6-e90d-4e8d-ac4d-e90fe3b330ad] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.09 ms]
2025/11/02 09:26:12 [da014a35-0849-49bf-8fde-c0f1c25ec63b] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.09 ms]
2025/11/02 09:26:12 [62eddbb8-6b79-45c4-987a-f89489f9f529] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.12 ms]
2025/11/02 09:26:12 [8fd37f36-0519-4f39-9dbd-ba850f7a41c9] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.13 ms]
2025/11/02 09:26:12 [eb11d9ba-fb84-4b08-a656-488e91d52c4a] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [10263f8a-57f3-4edf-9d83-be89ebc66986] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.06 ms]
2025/11/02 09:26:12 [e3b10e8e-4ccf-418a-886b-19e2d5fbd081] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.04 ms]
2025/11/02 09:26:12 [c2bd4689-f5ba-42bb-9ad1-26b6a703af7c] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [e1ea8576-bd7f-453c-bf23-b4b89e27f552] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.59 ms]
2025/11/02 09:26:12 [9a88c968-5679-4545-84fd-38eb35560bb7] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.15 ms]
2025/11/02 09:26:12 [e9e6302c-23e8-454b-84a4-578f276d9349] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.10 ms]
2025/11/02 09:26:12 [9a28779d-4d06-4612-a699-2625ba06afad] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.06 ms]
2025/11/02 09:26:12 [15733777-d032-4e14-a010-afc8cc22b7ed] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [ffc00c3f-bd3c-4526-a175-2ae536f1fcad] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.30 ms]
2025/11/02 09:26:12 [d90aa8ca-8b31-4aa7-8df7-cae0a1869033] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.04 ms]
2025/11/02 09:26:12 [76cb48fa-b474-432a-a046-4ba29f8752cb] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.26 ms]
2025/11/02 09:26:12 [7845d3b7-ba5b-4896-814a-37b1f600ecbf] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.37 ms]
2025/11/02 09:26:12 [465c7be4-3a42-44bd-b5a0-4ee593206930] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.06 ms]
2025/11/02 09:26:12 [964777a0-1799-4770-8ae0-67b424687cdc] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.29 ms]
2025/11/02 09:26:12 [c768f248-cea5-40d8-a1a2-aab73f13e623] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.06 ms]
2025/11/02 09:26:12 [6c70bd2f-3a1d-4d40-84dd-34e30255d00e] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.04 ms]
2025/11/02 09:26:12 [f7810cb7-723c-458c-b6cf-e94855ab7fff] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.12 ms]
2025/11/02 09:26:12 [57170ac0-afd7-469c-9f09-d8fd0de4170d] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.03 ms]
2025/11/02 09:26:12 [463f6f22-3bb2-496e-8122-12c2ee8716fa] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.10 ms]
2025/11/02 09:26:12 [9b6c84ff-2d54-4fd4-bc74-85a62fe0d2af] GET /matrixmul -> 503 (Service Unavailable) [PID=1211] [0.14 ms]
2025/11/02 09:26:12 [604937c4-a3fc-489f-ac77-6d20f1bb9db3] GET /matrixmul -> 200 (OK) [PID=1211] [58.45 ms]
2025/11/02 09:26:12 [1e25c128-e909-4253-80dd-8f819a28e588] GET /matrixmul -> 200 (OK) [PID=1211] [103.14 ms]
2025/11/02 09:26:12 [edc2c617-1de6-4cbf-8f88-d141d7c2ad91] GET /matrixmul -> 200 (OK) [PID=1211] [157.65 ms]
    benchmark_test.go:156: 
        [RESULTADOS] Perfil: medium
          Requests totales: 50 (Errores: 0)
          Duración total:   0.16s
          Throughput:       315.94 req/s
          Latencia p50:     0.00 ms
          Latencia p95:     1.00 ms
          Latencia p99:     103.00 ms
    benchmark_test.go:166: [OK] Perfil 'medium' completado correctamente
    benchmark_test.go:89: 
        --- [RUNNING] Perfil de carga: high (100 solicitudes, 10 concurrentes) ---
2025/11/02 09:26:12 [41b4227a-29ba-4dd2-8d03-446b07cb34e6] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.75 ms]
2025/11/02 09:26:12 [56c5c0b6-81f9-4692-8a3e-b82793a3ba78] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.78 ms]
2025/11/02 09:26:12 [d381387e-c983-4a1f-836c-b8b144557f45] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.92 ms]
2025/11/02 09:26:12 [912bd034-651a-40ee-a196-03f261f906ce] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.95 ms]
2025/11/02 09:26:12 [34cfbf5e-0f16-4011-840b-7d688ff2c85d] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [1.06 ms]
2025/11/02 09:26:12 [d71ddf3d-0838-44ff-9506-d4c15a727d08] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.06 ms]
2025/11/02 09:26:12 [4124f676-631e-4ecb-8a7a-cbb034b541f6] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.10 ms]
2025/11/02 09:26:12 [618103d4-33c1-494b-b1b7-66a7d4c08166] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.06 ms]
2025/11/02 09:26:12 [6cb7db1d-2db3-485b-baae-fc67232673e6] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.10 ms]
2025/11/02 09:26:12 [6d19c679-7410-403a-9ba5-c8af02cd5802] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.06 ms]
2025/11/02 09:26:12 [91a99fe6-7f1d-448d-aed0-6501cd750c21] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [cf0b19f7-12c3-47c9-a6a2-cd59d1cfa82c] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.10 ms]
2025/11/02 09:26:12 [038fe341-7753-434e-a1d4-eb54b13ac0b8] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.12 ms]
2025/11/02 09:26:12 [6a12b951-31c8-4665-8b27-12145245bc7c] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.04 ms]
2025/11/02 09:26:12 [6f3461b2-3a8b-4f65-9a7f-845692bed9e7] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [f5a275de-ebf3-45f5-b1ab-f2c4e73c9e03] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [3a195332-f40f-4293-b385-d8a4eafd1d52] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.10 ms]
2025/11/02 09:26:12 [76086679-5117-4179-a06f-6e1aa6be8c2a] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.06 ms]
2025/11/02 09:26:12 [f6e88dc0-e3f7-415f-bfbc-a93f21fb6a02] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.08 ms]
2025/11/02 09:26:12 [6011fd91-1ce1-4bd8-b668-d27e279ca019] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.07 ms]
2025/11/02 09:26:12 [4a84993a-8e38-4968-a0a3-658eb12586d4] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.11 ms]
2025/11/02 09:26:12 [89bcd7d4-6efb-4a24-a4a1-3bb87c519a57] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.14 ms]
2025/11/02 09:26:12 [388a23e3-a3f8-4422-b210-a88cde57c51a] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.20 ms]
2025/11/02 09:26:12 [a495511a-2062-4abb-a8c0-565f5e656cf8] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [2f151d99-7acb-440d-9073-1dfb3b84662e] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [1d9a1115-ca6f-49fd-8d7a-b9145d53e87e] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.10 ms]
2025/11/02 09:26:12 [9eae042f-84b4-46cf-8538-110fcd73b545] GET /sortfile -> 503 (Service Unavailable) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [d806a138-3c1e-4792-adc7-260b874ecc53] GET /sortfile -> 503 (Service Unavailable) [PID=1211] [0.04 ms]
2025/11/02 09:26:12 [b91794c2-a5f2-4bd7-aaa3-ffac636c8bb5] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.14 ms]
2025/11/02 09:26:12 [5966dabc-89ba-4e7f-bf94-b14aacbd2a38] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.17 ms]
2025/11/02 09:26:12 [3bbc20f1-bcbe-4c25-a987-955eacbab6f4] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.18 ms]
2025/11/02 09:26:12 [4f1f03d3-8242-4723-ba98-99f6c871ed2c] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.22 ms]
2025/11/02 09:26:12 [5a5ce95a-0230-466f-98d0-5bf8c5d6f749] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.08 ms]
2025/11/02 09:26:12 [693980da-6dbb-498e-89e2-a01a803b55b0] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.06 ms]
2025/11/02 09:26:12 [7ccecb1e-f96d-47e6-91b1-6ef2cef21cbf] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.04 ms]
2025/11/02 09:26:12 [66fdb21e-9bf3-44e5-b724-5cd7e32ca14f] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [2.87 ms]
2025/11/02 09:26:12 [e9c9b2a8-ccd6-4255-accd-9cb30dc36f88] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.06 ms]
2025/11/02 09:26:12 [abb4dcbd-2b7e-4741-a7f3-e554a9422921] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [7068ff0c-c003-4de9-8328-d5fa6d9aa34b] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.08 ms]
2025/11/02 09:26:12 [11e6e696-3e11-4d21-b441-95545c29338f] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.07 ms]
2025/11/02 09:26:12 [eb739f32-1678-4495-b19d-82cd9dad6762] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [d6ca6336-b85c-41ff-a8ad-5b68c881c6b2] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.19 ms]
2025/11/02 09:26:12 [a84dce44-797c-4c3a-907e-1b596900b031] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [bb73bdad-2a54-4338-9326-0b82ce2516db] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.03 ms]
2025/11/02 09:26:12 [d664e673-bc21-45c7-b25b-59f99e18daf3] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.16 ms]
2025/11/02 09:26:12 [b648cdb2-ff48-4e21-a9ea-5e3e21e48474] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.14 ms]
2025/11/02 09:26:12 [c60ddd34-1267-4362-b52d-95a45beefe3a] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.06 ms]
2025/11/02 09:26:12 [1211c0fe-367d-4c34-b326-cf3c5b68e862] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.33 ms]
2025/11/02 09:26:12 [3606979f-7965-4aee-abe6-2e6bc698569f] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.04 ms]
2025/11/02 09:26:12 [360e6f79-8e57-4a11-88dc-f70fda73e02d] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.12 ms]
2025/11/02 09:26:12 [32f3d61e-ebbb-4266-97cb-b15cbe9f7601] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [01e64e92-c784-4c30-9188-d64c89ec3bea] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.04 ms]
2025/11/02 09:26:12 [40cc89e6-dac4-4d93-8222-66bc96898c1b] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.11 ms]
2025/11/02 09:26:12 [1fc51386-af40-48c1-a043-f53066c9904f] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.14 ms]
2025/11/02 09:26:12 [f48280f3-ba2d-48a0-af10-70f0538cc2fe] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.10 ms]
2025/11/02 09:26:12 [37b642df-e2ce-4c2a-84cb-9e0885ef2da4] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.07 ms]
2025/11/02 09:26:12 [ed0efc3f-71e9-478b-8839-121f796ed4b6] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.04 ms]
2025/11/02 09:26:12 [a744440a-e44d-40c8-98a9-dc2cd0c6205a] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [fc4d7df6-7e08-4580-90c0-d6bbb03fcb34] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.06 ms]
2025/11/02 09:26:12 [ca344433-f680-4862-8fbe-e9cb40faadc7] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.04 ms]
2025/11/02 09:26:12 [891c2229-1ded-4cd2-b2a5-00dac3b93a61] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.09 ms]
2025/11/02 09:26:12 [847f9d2b-4f22-44a5-9141-a4ae0a8828b1] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.08 ms]
2025/11/02 09:26:12 [3090cb5c-4b5c-4304-9be8-5f574d09e150] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.10 ms]
2025/11/02 09:26:12 [4f3c138e-9323-46ba-b6e6-fb2a32a3bb04] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.04 ms]
2025/11/02 09:26:12 [e8d9d7d6-bb17-4c53-9ed6-9f4e609e2219] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.12 ms]
2025/11/02 09:26:12 [88378c76-6336-453e-a576-3759c7a3a5b9] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.68 ms]
2025/11/02 09:26:12 [c65c92c1-1919-495c-b67e-467d68a4c057] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [1.12 ms]
2025/11/02 09:26:12 [1da67ad6-7fd8-4400-9671-826efb386f51] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.30 ms]
2025/11/02 09:26:12 [53754e75-0230-46cd-81fd-bfc64cec0816] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.94 ms]
2025/11/02 09:26:12 [ec7166fd-fe8d-4f75-a662-183717e26499] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [1.02 ms]
2025/11/02 09:26:12 [0640d4e4-e458-4985-9b6e-5e91e508472d] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [1.06 ms]
2025/11/02 09:26:12 [5eaf33ae-8fac-4d88-a47d-a8acbe1794e7] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [1.25 ms]
2025/11/02 09:26:12 [9b1f460e-6ae3-4b4e-9588-73b5547ee8bc] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [1.27 ms]
2025/11/02 09:26:12 [d73dc1d3-4d91-4a28-960c-5e5a530a879d] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [1.43 ms]
2025/11/02 09:26:12 [f81f8af0-c66a-4786-988b-fabbee06d987] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.11 ms]
2025/11/02 09:26:12 [06d90a2a-bd64-4781-9cd0-235f82ddb46e] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.09 ms]
2025/11/02 09:26:12 [285b524d-ae01-4598-af5a-a2e1d9d0405e] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.21 ms]
2025/11/02 09:26:12 [55d12bcd-fbcf-4eba-9df6-81d20434b0e4] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [1.72 ms]
2025/11/02 09:26:12 [199de165-72a2-475a-8181-a759b925251f] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.17 ms]
2025/11/02 09:26:12 [90906418-0f79-4d3f-980c-4ca8a8089d59] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.04 ms]
2025/11/02 09:26:12 [b5a74f9b-8b20-4ed8-8aa8-ad117b23e18f] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.74 ms]
2025/11/02 09:26:12 [ebfa8201-08eb-427a-9d97-1ef2f00d9647] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [b12d07ca-454e-40ea-80b9-2c795faea033] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.13 ms]
2025/11/02 09:26:12 [1e7591e1-7e17-4cc7-a2c3-6266fd15fa94] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.06 ms]
2025/11/02 09:26:12 [0299a2cb-d54f-48a9-9cf5-26165b47f617] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [a65d06cf-b228-4cbd-8f66-a7b111ac846f] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.16 ms]
2025/11/02 09:26:12 [469c5ec0-f3be-4324-b85a-4b7ac9d01296] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.06 ms]
2025/11/02 09:26:12 [8a3973f8-4bb7-469a-9022-ae0dafbe5d77] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.06 ms]
2025/11/02 09:26:12 [660f8443-3b42-4e27-a05e-ff7247c49e84] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.24 ms]
2025/11/02 09:26:12 [511371a7-8f6c-45f8-90b2-0e236e5a41a1] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [f3a0030c-9e91-4db3-86f8-0f182f3bfcd3] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [75ea8c82-f370-4ce3-8667-65df373a5b00] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.14 ms]
2025/11/02 09:26:12 [e18aa35c-e33c-4914-8a42-f7c0df8c8ed0] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.17 ms]
2025/11/02 09:26:12 [bf388020-0e8b-4878-a5d2-715628d841a9] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.04 ms]
2025/11/02 09:26:12 [27ef1d6d-39b4-4444-b3a8-92897e70dad0] GET /sortfile -> 503 (Service Unavailable) [PID=1211] [0.07 ms]
2025/11/02 09:26:12 [9f2bca7e-baa9-47be-b050-149afdf8e1c6] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.16 ms]
2025/11/02 09:26:12 [817379dd-a775-4e65-9ed0-a1d90fa1217d] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.21 ms]
2025/11/02 09:26:12 [3580c07e-15ac-43c8-b319-a02129fc3aea] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.07 ms]
2025/11/02 09:26:12 [b60c0741-0f59-4116-944d-2bddc9fa4f7d] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.17 ms]
2025/11/02 09:26:12 [5f51655c-71ac-4cbd-8dc5-2580bff2256d] GET /sortfile -> 500 (Internal Server Error) [PID=1211] [0.44 ms]
    benchmark_test.go:156: 
        [RESULTADOS] Perfil: high
          Requests totales: 100 (Errores: 0)
          Duración total:   0.01s
          Throughput:       11395.01 req/s
          Latencia p50:     0.00 ms
          Latencia p95:     2.00 ms
          Latencia p99:     2.00 ms
    benchmark_test.go:166: [OK] Perfil 'high' completado correctamente
    benchmark_test.go:194: 
        [OK] Todos los perfiles de carga completados exitosamente.
--- PASS: TestBenchmark_Profiles (0.82s)
=== RUN   TestConcurrentClients
    concurrency_test.go:50: 
        --- [RUNNING] 20 clientes concurrentes sobre /reverse ---
2025/11/02 09:26:12 [71988c34-ad7c-4926-bc41-0d2919005751] GET /reverse -> 200 (OK) [PID=1211] [0.10 ms]
2025/11/02 09:26:12 [ec457a35-7311-48fc-a204-c40964585644] GET /reverse -> 200 (OK) [PID=1211] [0.13 ms]
2025/11/02 09:26:12 [e7ab4d47-4a10-45be-89ec-115a418bdd4e] GET /reverse -> 200 (OK) [PID=1211] [0.40 ms]
2025/11/02 09:26:12 [e24acf35-228c-4ad3-af6b-2e174e8c28c5] GET /reverse -> 200 (OK) [PID=1211] [0.24 ms]
2025/11/02 09:26:12 [ab7145e1-5614-4efe-9f6a-4b6493e1dd8b] GET /reverse -> 200 (OK) [PID=1211] [0.08 ms]
2025/11/02 09:26:12 [1ee3b9b1-9d2a-4d51-9f7c-020e54329719] GET /reverse -> 200 (OK) [PID=1211] [0.09 ms]
2025/11/02 09:26:12 [bcfeaf76-dbe8-48ad-ba0c-fc6efb81b5c3] GET /reverse -> 200 (OK) [PID=1211] [0.06 ms]
2025/11/02 09:26:12 [7c5bb70f-753d-4671-ad2c-9d9389d6d811] GET /reverse -> 200 (OK) [PID=1211] [0.21 ms]
2025/11/02 09:26:12 [6f28f1d1-58e8-4d42-9f0a-80f009decb6a] GET /reverse -> 200 (OK) [PID=1211] [0.23 ms]
2025/11/02 09:26:12 [f9d18871-803f-4117-945f-4d3d9a98eec0] GET /reverse -> 200 (OK) [PID=1211] [0.59 ms]
2025/11/02 09:26:12 [d0a01cc6-42e7-4914-8b1f-b53d01bac5b9] GET /reverse -> 200 (OK) [PID=1211] [0.14 ms]
2025/11/02 09:26:12 [40e36b37-75a3-4ab9-9131-52991e31214c] GET /reverse -> 200 (OK) [PID=1211] [0.04 ms]
2025/11/02 09:26:12 [b9ed645b-ba72-455d-9476-b51a4dfa2257] GET /reverse -> 200 (OK) [PID=1211] [0.24 ms]
2025/11/02 09:26:12 [c9983270-ffca-4862-bad4-1ca1b8d17d1d] GET /reverse -> 200 (OK) [PID=1211] [0.08 ms]
2025/11/02 09:26:12 [3298da43-2f92-4569-8a61-ed0c0520499d] GET /reverse -> 200 (OK) [PID=1211] [0.09 ms]
2025/11/02 09:26:12 [6ef21f75-1ad2-4f8d-920f-b121a3b47dfe] GET /reverse -> 200 (OK) [PID=1211] [0.11 ms]
2025/11/02 09:26:12 [e4a57d5a-cd39-4020-a156-151557425d2d] GET /reverse -> 200 (OK) [PID=1211] [0.05 ms]
2025/11/02 09:26:12 [c08eeae1-36dc-4bad-9f55-b40f97e9a3e7] GET /reverse -> 200 (OK) [PID=1211] [0.12 ms]
2025/11/02 09:26:12 [0192a6c7-60f3-4f64-8d23-32eb9fbf9c01] GET /reverse -> 200 (OK) [PID=1211] [1.00 ms]
2025/11/02 09:26:12 [e2b6667a-76df-4237-9400-375be1116b75] GET /reverse -> 200 (OK) [PID=1211] [1.26 ms]
    concurrency_test.go:90: [RESULT] 20 concurrent requests completed in 1.961563ms (errors: 0)
    concurrency_test.go:92: [OK] No se detectaron errores de concurrencia
--- PASS: TestConcurrentClients (0.00s)
=== RUN   TestJobQueuePressure
    concurrency_test.go:117: 
        --- [RUNNING] Saturación de cola con 30 jobs ---
2025/11/02 09:26:12 [39313394-014e-4140-91d1-6c2805cc6fde] GET /jobs/submit -> 200 (OK) [PID=1211] [22.43 ms]
2025/11/02 09:26:12 [d56833b4-9138-4fda-a571-ea77318488ac] GET /jobs/submit -> 200 (OK) [PID=1211] [4.20 ms]
2025/11/02 09:26:12 [e5131a93-e1e3-48bb-8855-682d0d8e3ac1] GET /jobs/submit -> 200 (OK) [PID=1211] [5.38 ms]
2025/11/02 09:26:12 [99c261c8-7d57-4a93-bb62-179f1a6d8978] GET /jobs/submit -> 200 (OK) [PID=1211] [14.99 ms]
2025/11/02 09:26:12 [5726190f-08e2-48a3-b271-5dfb4bcb81b9] GET /jobs/submit -> 200 (OK) [PID=1211] [17.60 ms]
2025/11/02 09:26:12 [54cb171d-f2ca-4dab-b3ab-7e6db6a2b2d5] GET /jobs/submit -> 200 (OK) [PID=1211] [8.23 ms]
2025/11/02 09:26:12 [1ed506f3-a697-42ae-8e60-e1e100030a26] GET /jobs/submit -> 200 (OK) [PID=1211] [4.48 ms]
2025/11/02 09:26:12 [0bd3b096-4024-46cf-b597-9f18a9e2a5bf] GET /jobs/submit -> 200 (OK) [PID=1211] [5.58 ms]
2025/11/02 09:26:12 [ac4898a9-1479-4b60-8a28-4f1f13e45214] GET /jobs/submit -> 200 (OK) [PID=1211] [5.49 ms]
2025/11/02 09:26:12 [68536e34-b9c2-4d5d-9b25-30ce7176da26] GET /jobs/submit -> 200 (OK) [PID=1211] [4.96 ms]
2025/11/02 09:26:12 [6d7ddf96-0ed6-452e-946d-a7e2d83e056c] GET /jobs/submit -> 200 (OK) [PID=1211] [6.98 ms]
2025/11/02 09:26:12 [30aace59-14c2-4dcf-90aa-c6f3045b5880] GET /jobs/submit -> 200 (OK) [PID=1211] [4.93 ms]
2025/11/02 09:26:12 [b6686b19-35b6-414b-8fea-23fad87548c7] GET /jobs/submit -> 200 (OK) [PID=1211] [4.74 ms]
2025/11/02 09:26:12 [71b3f476-d88f-4475-88a8-e5e71f0ce64e] GET /jobs/submit -> 200 (OK) [PID=1211] [12.63 ms]
2025/11/02 09:26:12 [cd05ae07-c1fe-460e-8c5b-ee2d473ac5e4] GET /jobs/submit -> 200 (OK) [PID=1211] [11.32 ms]
2025/11/02 09:26:12 [c327cdcc-b353-4a31-affe-4dc5f7179378] GET /jobs/submit -> 200 (OK) [PID=1211] [20.92 ms]
2025/11/02 09:26:12 [7765d572-6819-4d49-853f-b5f2e73a0c8f] GET /jobs/submit -> 200 (OK) [PID=1211] [13.58 ms]
2025/11/02 09:26:12 [f25d121f-24a9-4804-bd1f-b3f346289117] GET /jobs/submit -> 200 (OK) [PID=1211] [10.61 ms]
2025/11/02 09:26:12 [c45a13f7-edac-4c26-801a-ba484450a46b] GET /jobs/submit -> 200 (OK) [PID=1211] [10.98 ms]
2025/11/02 09:26:12 [c0cef81d-253e-4da4-a6a9-2f0ed4f1f133] GET /jobs/submit -> 200 (OK) [PID=1211] [10.78 ms]
2025/11/02 09:26:12 [e643ba3e-f68d-4893-a498-42112605d8ed] GET /jobs/submit -> 200 (OK) [PID=1211] [9.63 ms]
2025/11/02 09:26:12 [17af03e1-2a50-4936-a4d7-badf41662de0] GET /jobs/submit -> 200 (OK) [PID=1211] [11.61 ms]
2025/11/02 09:26:12 [c117ad0e-958e-456f-86ff-be96dcc0b3a6] GET /jobs/submit -> 200 (OK) [PID=1211] [11.07 ms]
2025/11/02 09:26:12 [764f39b1-405b-4c62-9bca-9063f3551e07] GET /jobs/submit -> 200 (OK) [PID=1211] [9.85 ms]
2025/11/02 09:26:12 [1b1b7436-5d87-4cf9-b134-1a5dcbc4af04] GET /jobs/submit -> 200 (OK) [PID=1211] [10.32 ms]
2025/11/02 09:26:12 [daf03e7d-93ba-411a-9e65-53519e9a007f] GET /jobs/submit -> 200 (OK) [PID=1211] [11.31 ms]
2025/11/02 09:26:12 [54679001-40c6-4cfa-bffb-23b6b10d910c] GET /jobs/submit -> 200 (OK) [PID=1211] [10.45 ms]
2025/11/02 09:26:12 [59dbb6d6-b7e4-46fd-9bdf-40eaa974ab61] GET /jobs/submit -> 200 (OK) [PID=1211] [9.81 ms]
2025/11/02 09:26:12 [547cc854-c2b6-40c0-aed2-c80c5e3c5c54] GET /jobs/submit -> 200 (OK) [PID=1211] [10.78 ms]
2025/11/02 09:26:12 [70643b55-8ba8-4272-a7d2-869161e96f91] GET /jobs/submit -> 200 (OK) [PID=1211] [9.95 ms]
    concurrency_test.go:142: [RESULT] Jobs OK: 30, Queue full (503): 0
    concurrency_test.go:149: [WARN] Cola no saturada — ajustar MAX_TOTAL o reducir límites para pruebas realistas
--- PASS: TestJobQueuePressure (0.34s)
=== RUN   TestMetricsAndStatusDuringLoad
    concurrency_test.go:170: 
        --- [RUNNING] Validando /metrics y /status durante ejecución de jobs ---
2025/11/02 09:26:12 [1fadad6f-a649-465a-809c-5158534d2b9a] GET /jobs/submit -> 200 (OK) [PID=1211] [9.65 ms]
2025/11/02 09:26:12 [d88367fa-19a2-46cb-83ef-0d4e809aad53] GET /jobs/submit -> 200 (OK) [PID=1211] [11.29 ms]
2025/11/02 09:26:12 [14960970-0efa-44ff-ba77-d931591bad7c] GET /jobs/submit -> 200 (OK) [PID=1211] [6.69 ms]
2025/11/02 09:26:12 [e12a89f9-a920-4c7e-947d-0e2078d46803] GET /jobs/submit -> 200 (OK) [PID=1211] [11.86 ms]
2025/11/02 09:26:12 [79449cfe-d7b4-4164-8148-b23fde2e03a5] GET /jobs/submit -> 200 (OK) [PID=1211] [11.76 ms]
2025/11/02 09:26:12 [96f2c7c6-e364-47d7-a01d-56e4fe5450d5] GET /metrics -> 200 (OK) [PID=1211] [0.96 ms]
2025/11/02 09:26:12 [7128fa54-a3fd-4b17-9c4c-fae1f914d994] GET /status -> 200 (OK) [PID=1211] [0.61 ms]
2025/11/02 09:26:13 [cde76899-d2f2-43bd-9b58-39f657c720d3] GET /metrics -> 200 (OK) [PID=1211] [0.53 ms]
2025/11/02 09:26:13 [f627bd9d-3997-4285-a4dd-8ab208b790be] GET /status -> 200 (OK) [PID=1211] [0.31 ms]
2025/11/02 09:26:13 [068e04b2-31b9-44c4-a5fc-a09ace1a7fc7] GET /metrics -> 200 (OK) [PID=1211] [0.49 ms]
2025/11/02 09:26:13 [82357d64-69b0-4d8a-a320-89e18565f15e] GET /status -> 200 (OK) [PID=1211] [0.59 ms]
    concurrency_test.go:198: [OK] /metrics y /status permanecieron operativos bajo carga
--- PASS: TestMetricsAndStatusDuringLoad (1.57s)
=== RUN   TestCancelMultipleJobs
    concurrency_test.go:221: 
        --- [RUNNING] Cancelación simultánea de 5 jobs ---
2025/11/02 09:26:14 [8359ccf1-ea75-41db-bfe1-f0d1056032ec] GET /jobs/submit -> 200 (OK) [PID=1211] [10.99 ms]
2025/11/02 09:26:14 [bfec08e9-8e88-4995-9146-74440e0d497c] GET /jobs/submit -> 200 (OK) [PID=1211] [11.42 ms]
2025/11/02 09:26:14 [7ece9f0a-8aef-4624-9571-ac2119f81ca6] GET /jobs/submit -> 200 (OK) [PID=1211] [11.24 ms]
2025/11/02 09:26:14 [10fe086b-438f-4de6-9f9b-529be92c2a7f] GET /jobs/submit -> 200 (OK) [PID=1211] [8.51 ms]
2025/11/02 09:26:14 [50269c03-9f72-4aa1-9fbb-0f0bf8f371ef] GET /jobs/submit -> 200 (OK) [PID=1211] [9.67 ms]
2025/11/02 09:26:14 [fba76b94-d7bd-407c-84a0-34d4b52471bd] GET /jobs/cancel -> 200 (OK) [PID=1211] [7.31 ms]
2025/11/02 09:26:14 [f2c5121f-4119-4283-a854-559cb3c15913] GET /jobs/cancel -> 200 (OK) [PID=1211] [10.39 ms]
2025/11/02 09:26:14 [e34b62fc-0f28-4ceb-9582-9eb571672a9d] GET /jobs/cancel -> 200 (OK) [PID=1211] [14.41 ms]
2025/11/02 09:26:14 [1f23145b-1dad-40a8-bf50-188e7b49552d] GET /jobs/cancel -> 200 (OK) [PID=1211] [18.30 ms]
2025/11/02 09:26:14 [f808674f-83ba-475a-a7e6-44bbfbb6613c] GET /jobs/cancel -> 200 (OK) [PID=1211] [21.39 ms]
2025/11/02 09:26:14 [458da6c7-daf5-494a-8dd2-0f40fbed1081] GET /jobs/status -> 200 (OK) [PID=1211] [5.55 ms]
2025/11/02 09:26:14 [05c4d785-55c3-4dd1-978b-a84bbfea1e71] GET /jobs/status -> 200 (OK) [PID=1211] [6.11 ms]
2025/11/02 09:26:14 [af351b22-3cc3-45e0-99a3-17be915e567f] GET /jobs/status -> 200 (OK) [PID=1211] [5.65 ms]
2025/11/02 09:26:14 [cabded6f-f6c2-4671-b79e-0c0e4ec4c298] GET /jobs/status -> 200 (OK) [PID=1211] [5.44 ms]
2025/11/02 09:26:14 [c7fe9ee0-8d05-47a8-92a1-951bf3db3562] GET /jobs/status -> 200 (OK) [PID=1211] [4.87 ms]
    concurrency_test.go:261: [OK] Múltiples cancelaciones concurrentes manejadas sin errores
--- PASS: TestCancelMultipleJobs (0.41s)
=== RUN   TestServer_ReverseEndpoint
2025/11/02 09:26:14 [9c7d76b1-bbf9-4300-8934-a3205e57e239] GET /reverse -> 200 (OK) [PID=1211] [0.27 ms]
--- PASS: TestServer_ReverseEndpoint (0.00s)
=== RUN   TestServer_ToUpper
2025/11/02 09:26:14 [e14365d7-f2b1-4c80-b650-3c54a16c4963] GET /toupper -> 200 (OK) [PID=1211] [0.34 ms]
--- PASS: TestServer_ToUpper (0.00s)
=== RUN   TestServer_StatusMetricsHelp
2025/11/02 09:26:14 [2d7c58b5-7566-4e36-be7e-7d80fb913fdd] GET /status -> 200 (OK) [PID=1211] [0.30 ms]
2025/11/02 09:26:14 [8b36387d-18cd-4444-a7d2-68b03174a4ed] GET /metrics -> 200 (OK) [PID=1211] [0.26 ms]
2025/11/02 09:26:14 [dc0e52d9-5d08-46e7-b1c4-a33953c24aa4] GET /help -> 200 (OK) [PID=1211] [0.40 ms]
--- PASS: TestServer_StatusMetricsHelp (0.00s)
=== RUN   TestServer_CreateDeleteFile
2025/11/02 09:26:14 [d7a1b287-640e-4b7b-802c-8e621b54b9f2] GET /createfile -> 200 (OK) [PID=1211] [0.34 ms]
2025/11/02 09:26:14 [5ac97a85-1622-4414-8e59-256a6e44220c] GET /deletefile -> 200 (OK) [PID=1211] [0.39 ms]
--- PASS: TestServer_CreateDeleteFile (0.00s)
=== RUN   TestServer_FibonacciIntegration
2025/11/02 09:26:14 [a970b994-c726-47df-9650-547b540237ec] GET /fibonacci -> 200 (OK) [PID=1211] [0.21 ms]
--- PASS: TestServer_FibonacciIntegration (0.00s)
=== RUN   TestServer_InvalidPath
2025/11/02 09:26:14 [0f92215c-3ab2-4c7f-b6a1-8165f43ce9ca] GET /notfound -> 404 (0.18 ms)
--- PASS: TestServer_InvalidPath (0.00s)
=== RUN   TestJobs_SubmitStatusResultFlow
2025/11/02 09:26:14 [c281184f-d363-451c-86f8-caff9a592511] GET /jobs/submit -> 200 (OK) [PID=1211] [8.10 ms]
2025/11/02 09:26:15 [21ae19b7-23ab-444e-9354-afc5d046d9e6] GET /jobs/status -> 200 (OK) [PID=1211] [5.30 ms]
2025/11/02 09:26:15 [8ef8eceb-ff75-43fc-874f-f9652b0bdc2f] GET /jobs/result -> 200 (OK) [PID=1211] [0.32 ms]
--- PASS: TestJobs_SubmitStatusResultFlow (1.02s)
=== RUN   TestJobs_Cancel
2025/11/02 09:26:15 [5296ccde-a38c-4b94-ae84-7270266a7881] GET /jobs/submit -> 200 (OK) [PID=1211] [3.78 ms]
2025/11/02 09:26:16 [4c766bd0-762a-40c7-af8e-442bfa016499] GET /jobs/cancel -> 200 (OK) [PID=1211] [11.46 ms]
2025/11/02 09:26:16 [2513490e-f055-4a94-835e-b1dbc4dddfec] GET /jobs/status -> 200 (OK) [PID=1211] [7.30 ms]
--- PASS: TestJobs_Cancel (0.63s)
=== RUN   TestServer_ConcurrentRequests
    integration_test.go:378: 
        --- [RUNNING] 10 concurrent requests to /reverse ---
2025/11/02 09:26:16 [76451a55-33c3-4cac-bde6-f51c36cff313] GET /reverse -> 200 (OK) [PID=1211] [0.06 ms]
2025/11/02 09:26:16 [8a04f615-ebcd-44bf-9384-1e861ddf4830] GET /reverse -> 200 (OK) [PID=1211] [0.06 ms]
2025/11/02 09:26:16 [436da8f1-84a7-4a91-9da6-a90cb5cfd9f8] GET /reverse -> 200 (OK) [PID=1211] [0.05 ms]
2025/11/02 09:26:16 [51482414-e35a-4a57-83af-369eafe2cbdc] GET /reverse -> 200 (OK) [PID=1211] [0.06 ms]
2025/11/02 09:26:16 [9b044aa4-0bcc-4fd8-a034-166d02f51078] GET /reverse -> 200 (OK) [PID=1211] [0.17 ms]
2025/11/02 09:26:16 [ab5eee17-4d52-4da4-9929-0df3d2cd8d44] GET /reverse -> 200 (OK) [PID=1211] [0.05 ms]
2025/11/02 09:26:16 [6cda98f4-ae5c-4107-af70-52cbc72b9456] GET /reverse -> 200 (OK) [PID=1211] [0.06 ms]
2025/11/02 09:26:16 [358b9814-0a5f-42a6-98a4-8ef20e7bf2be] GET /reverse -> 200 (OK) [PID=1211] [0.07 ms]
2025/11/02 09:26:16 [b7b22378-b6c6-4b81-b977-a2fed5aedb00] GET /reverse -> 200 (OK) [PID=1211] [0.05 ms]
2025/11/02 09:26:16 [f45747ed-dac9-49ea-b652-0f3db4857625] GET /reverse -> 200 (OK) [PID=1211] [0.05 ms]
    integration_test.go:419: [OK] All concurrent requests handled correctly (200/503)
--- PASS: TestServer_ConcurrentRequests (0.00s)
=== RUN   TestPool_GetPoolInfo_And_DefaultTimeout
--- PASS: TestPool_GetPoolInfo_And_DefaultTimeout (0.00s)
=== RUN   TestPool_SubmitAndWait_SuccessAndTimeout
--- PASS: TestPool_SubmitAndWait_SuccessAndTimeout (30.00s)
=== RUN   TestPool_InitPool_Duplicate
--- PASS: TestPool_InitPool_Duplicate (0.00s)
=== RUN   TestSetTimeout_ModifiesDefault
--- PASS: TestSetTimeout_ModifiesDefault (0.00s)
=== RUN   TestPerformance_PiHeavy
    performance_test.go:108: 
        --- [RUNNING] Pi(8000 dígitos) ---
2025/11/02 09:26:46 [343a7983-1d9d-4ac8-abe4-181ad5685f68] GET /pi -> 200 (OK) [PID=1211] [123.51 ms]
    performance_test.go:117: [DONE] Pi(8000 dígitos) completado en 124.674365ms
    performance_test.go:122: [WARN] Pi(8000 dígitos) se ejecutó demasiado rápido (posible entorno limitado)
--- PASS: TestPerformance_PiHeavy (0.12s)
=== RUN   TestPerformance_MatrixMulHeavy
    performance_test.go:108: 
        --- [RUNNING] MatrixMul(size=300, seed=7) ---
2025/11/02 09:26:46 [ed68f4bb-9ef2-4fa8-b56d-4fafca03e33d] GET /matrixmul -> 200 (OK) [PID=1211] [153.07 ms]
    performance_test.go:117: [DONE] MatrixMul(size=300, seed=7) completado en 153.963319ms
    performance_test.go:122: [WARN] MatrixMul(size=300, seed=7) se ejecutó demasiado rápido (posible entorno limitado)
--- PASS: TestPerformance_MatrixMulHeavy (0.15s)
=== RUN   TestPerformance_MandelbrotHeavy
    performance_test.go:108: 
        --- [RUNNING] Mandelbrot(1000x800, max_iter=300) ---
    performance_test.go:117: [DONE] Mandelbrot(1000x800, max_iter=300) completado en 579.685442ms
    performance_test.go:122: [WARN] Mandelbrot(1000x800, max_iter=300) se ejecutó demasiado rápido (posible entorno limitado)
--- PASS: TestPerformance_MandelbrotHeavy (0.58s)
=== RUN   TestPerformance_SortFile
2025/11/02 09:26:47 [a7046070-b439-4c3e-b1ba-67f5dfb63bc5] GET /mandelbrot -> 200 (OK) [PID=1211] [579.22 ms]
    performance_test.go:108: 
        --- [RUNNING] SortFile(≈5000000 líneas) ---
2025/11/02 09:26:59 [0c9c1988-7d25-414c-8c0d-7f21e1126ea3] GET /sortfile -> 200 (OK) [PID=1211] [2267.12 ms]
    performance_test.go:117: [DONE] SortFile(≈5000000 líneas) completado en 2.268612585s
--- PASS: TestPerformance_SortFile (12.26s)
=== RUN   TestPerformance_Compress
    performance_test.go:108: 
        --- [RUNNING] Compress(gzip, ≈1000000 repeticiones) ---
2025/11/02 09:26:59 [cfd44af6-f862-4900-ada9-7d97f7505c59] GET /compress -> 200 (OK) [PID=1211] [146.02 ms]
    performance_test.go:117: [DONE] Compress(gzip, ≈1000000 repeticiones) completado en 146.883087ms
    performance_test.go:122: [WARN] Compress(gzip, ≈1000000 repeticiones) se ejecutó demasiado rápido (posible entorno limitado)
--- PASS: TestPerformance_Compress (0.21s)
=== RUN   TestPerformance_HashFile
    performance_test.go:108: 
        --- [RUNNING] HashFile(sha256, ≈5500000 líneas) ---
2025/11/02 09:26:59 [b7c22fa4-6c5e-462f-adf8-33fdfb72687e] GET /hashfile -> 200 (OK) [PID=1211] [51.86 ms]
    performance_test.go:117: [DONE] HashFile(sha256, ≈5500000 líneas) completado en 53.021934ms
    performance_test.go:122: [WARN] HashFile(sha256, ≈5500000 líneas) se ejecutó demasiado rápido (posible entorno limitado)
--- PASS: TestPerformance_HashFile (0.12s)
=== RUN   TestPerformance_MetricsAfterHeavyLoad
2025/11/02 09:26:59 [f9420c0e-ec89-4065-9e22-19eaac5b2951] GET /metrics -> 200 (OK) [PID=1211] [0.69 ms]
    performance_test.go:374: [OK] /metrics respondió correctamente tras pruebas de carga pesada (CPU & IO)
--- PASS: TestPerformance_MetricsAfterHeavyLoad (0.00s)
=== RUN   TestReverse_Success
    unit_test.go:78: 
        --- [TestReverse_Success] Reversa de texto válido ---
--- PASS: TestReverse_Success (0.00s)
=== RUN   TestReverse_EmptyText
    unit_test.go:78: 
        --- [TestReverse_EmptyText] Reversa con texto vacío debe devolver 400 ---
--- PASS: TestReverse_EmptyText (0.00s)
=== RUN   TestToUpper_Success
    unit_test.go:78: 
        --- [TestToUpper_Success] Conversión a mayúsculas con entrada válida ---
--- PASS: TestToUpper_Success (0.00s)
=== RUN   TestToUpper_MissingParam
    unit_test.go:78: 
        --- [TestToUpper_MissingParam] Falta parámetro 'text' debe devolver 400 ---
--- PASS: TestToUpper_MissingParam (0.00s)
=== RUN   TestHash_Success
    unit_test.go:78: 
        --- [TestHash_Success] Hash de texto válido produce sha256 ---
--- PASS: TestHash_Success (0.00s)
=== RUN   TestHash_Empty
    unit_test.go:78: 
        --- [TestHash_Empty] Hash con texto vacío debe devolver 400 ---
--- PASS: TestHash_Empty (0.00s)
=== RUN   TestRandom_Success
    unit_test.go:78: 
        --- [TestRandom_Success] Generación de números aleatorios válida ---
--- PASS: TestRandom_Success (0.00s)
=== RUN   TestRandom_InvalidRange
    unit_test.go:78: 
        --- [TestRandom_InvalidRange] Rango inválido (min>max) debe devolver 400 ---
--- PASS: TestRandom_InvalidRange (0.00s)
=== RUN   TestTimestamp_Success
    unit_test.go:78: 
        --- [TestTimestamp_Success] Timestamp debe incluir campo 'iso' ---
--- PASS: TestTimestamp_Success (0.00s)
=== RUN   TestTimestamp_Cancelled
    unit_test.go:78: 
        --- [TestTimestamp_Cancelled] Cancelación previa al inicio (sin invocación) ---
--- PASS: TestTimestamp_Cancelled (0.00s)
=== RUN   TestSimulate_Success
    unit_test.go:78: 
        --- [TestSimulate_Success] Simulación de trabajo exitosa ---
--- PASS: TestSimulate_Success (1.00s)
=== RUN   TestSimulate_Cancelled
    unit_test.go:78: 
        --- [TestSimulate_Cancelled] Simulación cancelada debe devolver 499 ---
--- PASS: TestSimulate_Cancelled (1.00s)
=== RUN   TestSleep_Success
    unit_test.go:78: 
        --- [TestSleep_Success] Sleep con segundos válidos ---
--- PASS: TestSleep_Success (1.00s)
=== RUN   TestSleep_Cancel
    unit_test.go:78: 
        --- [TestSleep_Cancel] Sleep cancelado debe devolver 499 ---
--- PASS: TestSleep_Cancel (1.00s)
=== RUN   TestSleep_InvalidSeconds
    unit_test.go:78: 
        --- [TestSleep_InvalidSeconds] Segundos inválidos (<=0) debe devolver 400 ---
--- PASS: TestSleep_InvalidSeconds (0.00s)
=== RUN   TestSleep_CancelledBeforeStart
    unit_test.go:78: 
        --- [TestSleep_CancelledBeforeStart] Cancelación previa al inicio debe devolver 499 ---
--- PASS: TestSleep_CancelledBeforeStart (0.00s)
=== RUN   TestLoadTest_Success
    unit_test.go:78: 
        --- [TestLoadTest_Success] Ejecución de tareas concurrentes válida ---
--- PASS: TestLoadTest_Success (1.00s)
=== RUN   TestLoadTest_InvalidParams
    unit_test.go:78: 
        --- [TestLoadTest_InvalidParams] Parámetros inválidos en LoadTest (tasks=0) ---
--- PASS: TestLoadTest_InvalidParams (0.00s)
=== RUN   TestLoadTest_NegativeSleep
    unit_test.go:78: 
        --- [TestLoadTest_NegativeSleep] Parámetro sleep negativo debe devolver 400 ---
--- PASS: TestLoadTest_NegativeSleep (0.00s)
=== RUN   TestLoadTest_CancelledBeforeStart
    unit_test.go:78: 
        --- [TestLoadTest_CancelledBeforeStart] Cancelación antes de iniciar debe devolver 499 ---
--- PASS: TestLoadTest_CancelledBeforeStart (0.00s)
=== RUN   TestLoadTest_CancelledDuringExecution
    unit_test.go:78: 
        --- [TestLoadTest_CancelledDuringExecution] Cancelación durante ejecución (sin llamada, conserva intención original) ---
--- PASS: TestLoadTest_CancelledDuringExecution (0.00s)
=== RUN   TestHashText_Consistency
    unit_test.go:78: 
        --- [TestHashText_Consistency] Mismo input debe producir mismo hash ---
--- PASS: TestHashText_Consistency (0.00s)
=== RUN   TestTimestamp_NotEmpty
    unit_test.go:78: 
        --- [TestTimestamp_NotEmpty] Body del timestamp no debe estar vacío ---
--- PASS: TestTimestamp_NotEmpty (0.00s)
=== RUN   TestReverse_Cancelled
    unit_test.go:78: 
        --- [TestReverse_Cancelled] Cancelación previa o bad request ---
--- PASS: TestReverse_Cancelled (0.00s)
=== RUN   TestCreateTempFileForFutureIO
    unit_test.go:78: 
        --- [TestCreateTempFileForFutureIO] Crear/eliminar archivo temporal ---
--- PASS: TestCreateTempFileForFutureIO (0.00s)
=== RUN   TestFibonacci_Valid
    unit_test.go:78: 
        --- [TestFibonacci_Valid] Serie de Fibonacci válida ---
--- PASS: TestFibonacci_Valid (0.00s)
=== RUN   TestFibonacci_Invalid
    unit_test.go:78: 
        --- [TestFibonacci_Invalid] n inválido debe devolver 400 ---
--- PASS: TestFibonacci_Invalid (0.00s)
=== RUN   TestIsPrime_TrialMethod
    unit_test.go:78: 
        --- [TestIsPrime_TrialMethod] Primalidad con método 'trial' ---
--- PASS: TestIsPrime_TrialMethod (0.00s)
=== RUN   TestIsPrime_TrialMethod2
    unit_test.go:78: 
        --- [TestIsPrime_TrialMethod2] Primalidad para 2 con método 'trial' ---
--- PASS: TestIsPrime_TrialMethod2 (0.00s)
=== RUN   TestIsPrime_TrialMethod3
    unit_test.go:78: 
        --- [TestIsPrime_TrialMethod3] No primalidad para 4 con 'trial' ---
--- PASS: TestIsPrime_TrialMethod3 (0.00s)
=== RUN   TestIsPrime_TrialMethod4
    unit_test.go:78: 
        --- [TestIsPrime_TrialMethod4] Valores <2 no son primos (trial) ---
--- PASS: TestIsPrime_TrialMethod4 (0.00s)
=== RUN   TestIsPrime_TrialMethod5
    unit_test.go:78: 
        --- [TestIsPrime_TrialMethod5] No primalidad para 35 (trial) ---
--- PASS: TestIsPrime_TrialMethod5 (0.00s)
=== RUN   TestIsPrime_MillerMethod
    unit_test.go:78: 
        --- [TestIsPrime_MillerMethod] Primalidad con método 'miller' ---
--- PASS: TestIsPrime_MillerMethod (0.00s)
=== RUN   TestIsPrime_InvalidMethod
    unit_test.go:78: 
        --- [TestIsPrime_InvalidMethod] Método inválido debe devolver 400 ---
--- PASS: TestIsPrime_InvalidMethod (0.00s)
=== RUN   TestIsPrime_DefaultMethod
    unit_test.go:78: 
        --- [TestIsPrime_DefaultMethod] Sin método => usa 'trial' ---
--- PASS: TestIsPrime_DefaultMethod (0.00s)
=== RUN   TestIsPrime_CancelledImmediately
    unit_test.go:78: 
        --- [TestIsPrime_CancelledImmediately] Cancelación inmediata debe devolver 499 ---
--- PASS: TestIsPrime_CancelledImmediately (0.00s)
=== RUN   TestIsPrime_CancelDuringTrial
    unit_test.go:78: 
        --- [TestIsPrime_CancelDuringTrial] Cancel durante trial en número grande ---
--- PASS: TestIsPrime_CancelDuringTrial (0.00s)
=== RUN   TestIsPrime_CancelDuringMiller
    unit_test.go:78: 
        --- [TestIsPrime_CancelDuringMiller] Cancel durante Miller-Rabin ---
--- PASS: TestIsPrime_CancelDuringMiller (0.00s)
=== RUN   TestIsPrime_MillerCompositeBranches
    unit_test.go:78: 
        --- [TestIsPrime_MillerCompositeBranches] Cobertura de ramas compuestas (Miller) ---
--- PASS: TestIsPrime_MillerCompositeBranches (0.00s)
=== RUN   TestFactorNumber_Valid
    unit_test.go:78: 
        --- [TestFactorNumber_Valid] Factorización de 84 ---
--- PASS: TestFactorNumber_Valid (0.00s)
=== RUN   TestFactorNumber_Valid2
    unit_test.go:78: 
        --- [TestFactorNumber_Valid2] Factorización de 15 ---
--- PASS: TestFactorNumber_Valid2 (0.00s)
=== RUN   TestFactorNumber_PrimeInput
    unit_test.go:78: 
        --- [TestFactorNumber_PrimeInput] Factorización de primo debe ser [n] ---
--- PASS: TestFactorNumber_PrimeInput (0.00s)
=== RUN   TestFactorNumber_Invalid
    unit_test.go:78: 
        --- [TestFactorNumber_Invalid] Entrada negativa debe devolver 400 ---
--- PASS: TestFactorNumber_Invalid (0.00s)
=== RUN   TestMatrixMultiply_Valid
    unit_test.go:78: 
        --- [TestMatrixMultiply_Valid] Multiplicación de matrices válida ---
--- PASS: TestMatrixMultiply_Valid (0.00s)
=== RUN   TestMatrixMultiply_InvalidSize
    unit_test.go:78: 
        --- [TestMatrixMultiply_InvalidSize] Size inválido debe devolver 400 ---
--- PASS: TestMatrixMultiply_InvalidSize (0.00s)
=== RUN   TestMandelbrot_Valid
    unit_test.go:78: 
        --- [TestMandelbrot_Valid] Generación válida de fractal ---
--- PASS: TestMandelbrot_Valid (0.00s)
=== RUN   TestMandelbrot_InvalidParams
    unit_test.go:78: 
        --- [TestMandelbrot_InvalidParams] Parámetros inválidos deben devolver 400 ---
--- PASS: TestMandelbrot_InvalidParams (0.00s)
=== RUN   TestPi_Valid
    unit_test.go:78: 
        --- [TestPi_Valid] Cálculo de PI con dígitos válidos ---
--- PASS: TestPi_Valid (0.00s)
=== RUN   TestPi_InvalidDigits
    unit_test.go:78: 
        --- [TestPi_InvalidDigits] Dígitos inválidos deben devolver 400 ---
--- PASS: TestPi_InvalidDigits (0.00s)
=== RUN   TestFibonacci_Cancelled
    unit_test.go:78: 
        --- [TestFibonacci_Cancelled] Cancelación durante Fibonacci ---
--- PASS: TestFibonacci_Cancelled (0.00s)
=== RUN   TestMatrixMultiply_Cancelled
    unit_test.go:78: 
        --- [TestMatrixMultiply_Cancelled] Cancelación durante multiplicación de matrices grande ---
--- PASS: TestMatrixMultiply_Cancelled (0.02s)
=== RUN   TestIsPrime_LargeNumber
    unit_test.go:78: 
        --- [TestIsPrime_LargeNumber] Prueba con número grande (bordes) ---
--- PASS: TestIsPrime_LargeNumber (0.00s)
=== RUN   TestMandelbrot_SaveFileTrue
    unit_test.go:78: 
        --- [TestMandelbrot_SaveFileTrue] Generación con volcado a archivo ---
--- PASS: TestMandelbrot_SaveFileTrue (0.00s)
=== RUN   TestPi_Consistency
    unit_test.go:78: 
        --- [TestPi_Consistency] Determinismo en cálculo de PI ---
--- PASS: TestPi_Consistency (0.00s)
=== RUN   TestMatrixMul_SeedEffect
    unit_test.go:78: 
        --- [TestMatrixMul_SeedEffect] Seeds diferentes deben producir hashes distintos ---
--- PASS: TestMatrixMul_SeedEffect (0.00s)
=== RUN   TestFactorNumber_StringConversion
    unit_test.go:78: 
        --- [TestFactorNumber_StringConversion] Sanity check de strconv.Itoa ---
--- PASS: TestFactorNumber_StringConversion (0.00s)
=== RUN   TestCreateFile_Success
    unit_test.go:78: 
        --- [TestCreateFile_Success] Creación de archivo válida ---
--- PASS: TestCreateFile_Success (0.00s)
=== RUN   TestCreateFile_Invalid
    unit_test.go:78: 
        --- [TestCreateFile_Invalid] Falta de nombre de archivo debe devolver 400 ---
--- PASS: TestCreateFile_Invalid (0.00s)
=== RUN   TestCreateFile_DefaultRepeat
    unit_test.go:78: 
        --- [TestCreateFile_DefaultRepeat] Repeat<=0 debe usar default=1 ---
--- PASS: TestCreateFile_DefaultRepeat (0.00s)
=== RUN   TestCreateFile_CancelledBeforeStart
    unit_test.go:78: 
        --- [TestCreateFile_CancelledBeforeStart] Cancelado antes de iniciar no debe crear archivo ---
--- PASS: TestCreateFile_CancelledBeforeStart (0.00s)
=== RUN   TestCreateFile_WriteError
    unit_test.go:78: 
        --- [TestCreateFile_WriteError] Error de escritura (ruta inválida) debe devolver 500 ---
--- PASS: TestCreateFile_WriteError (0.00s)
=== RUN   TestDeleteFile_Success
    unit_test.go:78: 
        --- [TestDeleteFile_Success] Borrado de archivo existente ---
--- PASS: TestDeleteFile_Success (0.00s)
=== RUN   TestDeleteFile_NotFound
    unit_test.go:78: 
        --- [TestDeleteFile_NotFound] Borrado de archivo inexistente debe devolver 500 ---
--- PASS: TestDeleteFile_NotFound (0.00s)
=== RUN   TestDeleteFile_MissingName
    unit_test.go:78: 
        --- [TestDeleteFile_MissingName] Falta de nombre debe devolver 400 ---
--- PASS: TestDeleteFile_MissingName (0.00s)
=== RUN   TestDeleteFile_CancelledBeforeStart
    unit_test.go:78: 
        --- [TestDeleteFile_CancelledBeforeStart] Cancelado antes de iniciar no debe eliminar archivo ---
--- PASS: TestDeleteFile_CancelledBeforeStart (0.00s)
=== RUN   TestSortFile_Success
    unit_test.go:78: 
        --- [TestSortFile_Success] Ordenamiento de números por merge sort ---
--- PASS: TestSortFile_Success (0.00s)
=== RUN   TestSortFile_InvalidAlgo
    unit_test.go:78: 
        --- [TestSortFile_InvalidAlgo] Algoritmo inválido debe devolver 400 ---
--- PASS: TestSortFile_InvalidAlgo (0.00s)
=== RUN   TestSortFile_MissingName
    unit_test.go:78: 
        --- [TestSortFile_MissingName] Falta de nombre de archivo debe devolver 400 ---
--- PASS: TestSortFile_MissingName (0.00s)
=== RUN   TestSortFile_FileNotFound
    unit_test.go:78: 
        --- [TestSortFile_FileNotFound] Archivo inexistente debe devolver 500 ---
--- PASS: TestSortFile_FileNotFound (0.00s)
=== RUN   TestSortFile_EmptyFile
    unit_test.go:78: 
        --- [TestSortFile_EmptyFile] Archivo vacío debe devolver 400 ---
--- PASS: TestSortFile_EmptyFile (0.00s)
=== RUN   TestSortFile_CancelledBeforeStart
    unit_test.go:78: 
        --- [TestSortFile_CancelledBeforeStart] Cancelación antes de iniciar debe devolver 499 ---
--- PASS: TestSortFile_CancelledBeforeStart (0.00s)
=== RUN   TestSortFile_CancelDuringRead
    unit_test.go:78: 
        --- [TestSortFile_CancelDuringRead] Cancelación durante lectura (sin llamada, se elimina el archivo) ---
--- PASS: TestSortFile_CancelDuringRead (0.00s)
=== RUN   TestSortFile_CancelDuringWrite
    unit_test.go:78: 
        --- [TestSortFile_CancelDuringWrite] Cancelación durante escritura del .sorted ---
--- PASS: TestSortFile_CancelDuringWrite (0.00s)
=== RUN   TestSortFile_DefaultQuickSort
    unit_test.go:78: 
        --- [TestSortFile_DefaultQuickSort] Algoritmo por defecto quicksort cuando 'algo' está vacío ---
--- PASS: TestSortFile_DefaultQuickSort (0.00s)
=== RUN   TestWordCount_Success
    unit_test.go:78: 
        --- [TestWordCount_Success] Recuento de un archivo simple ---
--- PASS: TestWordCount_Success (0.00s)
=== RUN   TestWordCount_MissingFile
    unit_test.go:78: 
        --- [TestWordCount_MissingFile] Archivo inexistente debe devolver 500 ---
--- PASS: TestWordCount_MissingFile (0.00s)
=== RUN   TestGrep_Success
    unit_test.go:78: 
        --- [TestGrep_Success] Búsqueda de patrón con coincidencias ---
--- PASS: TestGrep_Success (0.00s)
=== RUN   TestGrep_InvalidRegex
    unit_test.go:78: 
        --- [TestGrep_InvalidRegex] Regex inválido debe devolver 400 ---
--- PASS: TestGrep_InvalidRegex (0.00s)
=== RUN   TestGrep_MissingParams
    unit_test.go:78: 
        --- [TestGrep_MissingParams] Falta de parámetros debe devolver 400 ---
--- PASS: TestGrep_MissingParams (0.00s)
=== RUN   TestGrep_FileNotFound
    unit_test.go:78: 
        --- [TestGrep_FileNotFound] Archivo inexistente debe devolver 500 ---
--- PASS: TestGrep_FileNotFound (0.00s)
=== RUN   TestGrep_CancelledDuringRead
    unit_test.go:78: 
        --- [TestGrep_CancelledDuringRead] Cancelación durante lectura (sin llamada, conserva intención) ---
--- PASS: TestGrep_CancelledDuringRead (0.00s)
=== RUN   TestGrep_NoMatches
    unit_test.go:78: 
        --- [TestGrep_NoMatches] Búsqueda sin coincidencias debe retornar matches=0 ---
--- PASS: TestGrep_NoMatches (0.00s)
=== RUN   TestGrep_ScannerError
    unit_test.go:78: 
        --- [TestGrep_ScannerError] Pasar directorio debe provocar error (500) ---
--- PASS: TestGrep_ScannerError (0.00s)
=== RUN   TestHashFile_Success
    unit_test.go:78: 
        --- [TestHashFile_Success] Hash de archivo válido ---
--- PASS: TestHashFile_Success (0.00s)
=== RUN   TestHashFile_InvalidAlgo
    unit_test.go:78: 
        --- [TestHashFile_InvalidAlgo] Algoritmo inválido debe devolver 400 ---
--- PASS: TestHashFile_InvalidAlgo (0.00s)
=== RUN   TestCompressFile_GzipSuccess
    unit_test.go:78: 
        --- [TestCompressFile_GzipSuccess] Compresión gzip con archivo válido ---
--- PASS: TestCompressFile_GzipSuccess (0.00s)
=== RUN   TestCompressFile_InvalidCodec
    unit_test.go:78: 
        --- [TestCompressFile_InvalidCodec] Codec inválido debe devolver 400 ---
--- PASS: TestCompressFile_InvalidCodec (0.00s)
=== RUN   TestCompressFile_Cancelled
    unit_test.go:78: 
        --- [TestCompressFile_Cancelled] Cancelación durante compresión gzip ---
--- PASS: TestCompressFile_Cancelled (0.00s)
=== RUN   TestCompressFile_MissingName
    unit_test.go:78: 
        --- [TestCompressFile_MissingName] Falta de nombre debe devolver 400 ---
--- PASS: TestCompressFile_MissingName (0.00s)
=== RUN   TestCompressFile_FileNotFound
    unit_test.go:78: 
        --- [TestCompressFile_FileNotFound] Archivo inexistente debe devolver 500 ---
--- PASS: TestCompressFile_FileNotFound (0.00s)
=== RUN   TestCompressFile_DefaultCodec
    unit_test.go:78: 
        --- [TestCompressFile_DefaultCodec] Codec por defecto gzip cuando string vacío ---
--- PASS: TestCompressFile_DefaultCodec (0.00s)
=== RUN   TestCompressFile_XZCodec
    unit_test.go:78: 
        --- [TestCompressFile_XZCodec] Compresión con xz (resultado dependiente del entorno) ---
--- PASS: TestCompressFile_XZCodec (0.01s)
=== RUN   TestCompressFile_CreateOutputError
    unit_test.go:78: 
        --- [TestCompressFile_CreateOutputError] Error al crear archivo de salida debe devolver 500 ---
--- PASS: TestCompressFile_CreateOutputError (0.00s)
=== RUN   TestCompressFile_EmptyFile
    unit_test.go:78: 
        --- [TestCompressFile_EmptyFile] Compresión de archivo vacío (200) ---
--- PASS: TestCompressFile_EmptyFile (0.00s)
=== RUN   TestCompressFile_FileMissingAndXZ
    unit_test.go:78: 
        --- [TestCompressFile_FileMissingAndXZ] Missing file (500) y luego xz (200) ---
--- PASS: TestCompressFile_FileMissingAndXZ (0.00s)
=== RUN   TestSortFile_QuickSort
    unit_test.go:78: 
        --- [TestSortFile_QuickSort] Ordenamiento quicksort explícito ---
--- PASS: TestSortFile_QuickSort (0.00s)
=== RUN   TestIsPrime_EdgeCases
    unit_test.go:78: 
        --- [TestIsPrime_EdgeCases] Casos límite: 0,1,2,4 ---
--- PASS: TestIsPrime_EdgeCases (0.00s)
=== RUN   TestLoadTest_Cancelled
    unit_test.go:78: 
        --- [TestLoadTest_Cancelled] Cancelación durante LoadTest ---
--- PASS: TestLoadTest_Cancelled (1.00s)
=== RUN   TestOverall_Consistency
    unit_test.go:78: 
        --- [TestOverall_Consistency] Resumen de verificación general ---
    unit_test.go:1830: Todos los algoritmos básicos, CPU-bound e IO-bound verificados con éxito
--- PASS: TestOverall_Consistency (0.00s)
=== RUN   TestLoadTest_InvalidAndCancel
    unit_test.go:78: 
        --- [TestLoadTest_InvalidAndCancel] Params inválidos y cancelación en LoadTest ---
--- PASS: TestLoadTest_InvalidAndCancel (1.00s)
=== RUN   TestHashFile_FileMissingAndCancel
    unit_test.go:78: 
        --- [TestHashFile_FileMissingAndCancel] Missing file (500) y cancelación inmediata ---
--- PASS: TestHashFile_FileMissingAndCancel (0.00s)
=== RUN   TestSimulateWork_EdgeCases
    unit_test.go:78: 
        --- [TestSimulateWork_EdgeCases] Segundos=0 (400) y cancelación inmediata ---
--- PASS: TestSimulateWork_EdgeCases (0.00s)
=== RUN   TestSortFile_EmptyAndBadNumbers
    unit_test.go:78: 
        --- [TestSortFile_EmptyAndBadNumbers] Archivo vacío (400) y líneas no numéricas (400) ---
--- PASS: TestSortFile_EmptyAndBadNumbers (0.00s)
=== RUN   TestLoadTest_EdgeAndCancel
    unit_test.go:78: 
        --- [TestLoadTest_EdgeAndCancel] Tasks=0 (400), sleep<0 (400) y cancelación ---
--- PASS: TestLoadTest_EdgeAndCancel (1.00s)
=== RUN   TestGrep_NoMatchesAndCancel
    unit_test.go:78: 
        --- [TestGrep_NoMatchesAndCancel] Sin coincidencias y cancelación inmediata ---
--- PASS: TestGrep_NoMatchesAndCancel (0.00s)
=== RUN   TestRandom_EdgeCases
    unit_test.go:78: 
        --- [TestRandom_EdgeCases] count=0 (400), min=max (400) y cancelación ---
--- PASS: TestRandom_EdgeCases (0.00s)
PASS
coverage: 90.3% of statements in github.com/EngSteven/pso-http-server/internal/algorithms
ok      github.com/EngSteven/pso-http-server/tests      56.365s coverage: 90.3% of statements in github.com/EngSteven/pso-http-server/internal/algorithms

```
### Pruebas pesadas
```bash
====================================================
Running all tests with coverage
====================================================
go test ./tests -v -count=1 \
    -coverpkg=github.com/EngSteven/pso-http-server/internal/algorithms \
    -coverprofile=reports/coverage.out
[INFO] PERF_LEVEL=heavy → pruebas intensivas activadas

============================================================
TEST SUITE — Algoritmos del Servidor HTTP (PSO_PY01b)
Descripción: Pruebas unitarias por categorías (básicos, CPU, IO)
============================================================
=== RUN   TestBenchmark_Profiles
    integration_test.go:61: 
        ============================================================
    integration_test.go:62: SETUP: Iniciando servidor embebido para pruebas de integración
    integration_test.go:63: ============================================================
2025/11/02 09:29:59 Servidor escuchando en :8080
2025/11/02 09:29:59 [83e94764-b1d8-4fc4-9989-ebcf4f490bb2] GET /status -> 200 (OK) [PID=2199] [0.71 ms]
    integration_test.go:71: Servidor embebido disponible en: http://localhost:8080
    benchmark_test.go:184: 
        ============================================================
    benchmark_test.go:185: BENCHMARK DE CARGA — Iniciando perfiles (low, medium, high)
    benchmark_test.go:186: ============================================================
    benchmark_test.go:89: 
        --- [RUNNING] Perfil de carga: low (10 solicitudes, 2 concurrentes) ---
2025/11/02 09:29:59 [3a50200a-5d90-4914-b27a-5a923300c47f] GET /pi -> 200 (OK) [PID=2199] [35.71 ms]
2025/11/02 09:29:59 [89e9b011-df97-4ba6-a0a9-72cd33353c1f] GET /pi -> 200 (OK) [PID=2199] [65.72 ms]
2025/11/02 09:29:59 [b1fbe66b-b085-48b0-94ef-818b05479fa4] GET /pi -> 200 (OK) [PID=2199] [60.76 ms]
2025/11/02 09:29:59 [757db77d-d348-4410-b864-44cf198f9e60] GET /pi -> 200 (OK) [PID=2199] [87.08 ms]
2025/11/02 09:29:59 [df78e676-8b5e-4b70-8f7f-6bbb349a385c] GET /pi -> 200 (OK) [PID=2199] [105.82 ms]
2025/11/02 09:29:59 [88f7e721-9889-4fe6-b234-a49a1205a891] GET /pi -> 200 (OK) [PID=2199] [103.90 ms]
2025/11/02 09:29:59 [2d205599-5f8c-4c3a-ae6f-7e06bb7592a2] GET /pi -> 200 (OK) [PID=2199] [104.70 ms]
2025/11/02 09:29:59 [e9352646-a846-4948-bad3-a118dab0c359] GET /pi -> 200 (OK) [PID=2199] [106.27 ms]
2025/11/02 09:29:59 [9a76aad1-8cb1-45dc-8f98-8a1ba16067f4] GET /pi -> 200 (OK) [PID=2199] [115.94 ms]
2025/11/02 09:29:59 [21908d18-df19-4c4d-9812-0e7e8168009e] GET /pi -> 200 (OK) [PID=2199] [107.77 ms]
    benchmark_test.go:156: 
        [RESULTADOS] Perfil: low
          Requests totales: 10 (Errores: 0)
          Duración total:   0.48s
          Throughput:       20.97 req/s
          Latencia p50:     105.00 ms
          Latencia p95:     109.00 ms
          Latencia p99:     109.00 ms
    benchmark_test.go:166: [OK] Perfil 'low' completado correctamente
    benchmark_test.go:89: 
        --- [RUNNING] Perfil de carga: medium (50 solicitudes, 5 concurrentes) ---
2025/11/02 09:29:59 [c5baa591-bd6d-4311-8f69-32aa409f6523] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.36 ms]
2025/11/02 09:29:59 [51c75e26-83b2-407b-a192-cba6f44ae4a3] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [1.69 ms]
2025/11/02 09:29:59 [e3500cc4-6b59-4d6b-9228-eb43869b12b3] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.17 ms]
2025/11/02 09:29:59 [6d485abd-785a-4c01-baeb-64908111f5d8] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.16 ms]
2025/11/02 09:29:59 [f87828a1-6ab2-418e-8305-fb57d3328cde] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.20 ms]
2025/11/02 09:29:59 [7043d6b3-09bd-41fb-9a20-bd11bf658671] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.36 ms]
2025/11/02 09:29:59 [0130da95-3d4e-436b-93e7-6bcbbde2a787] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [2.59 ms]
2025/11/02 09:29:59 [ca8a1318-2a3d-4940-92a7-b1338962d159] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [1.03 ms]
2025/11/02 09:29:59 [d7fdf3c5-0be6-4073-8b17-819215748b9a] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.39 ms]
2025/11/02 09:29:59 [4dbf173e-a823-46b0-8484-d45f4acf3147] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.17 ms]
2025/11/02 09:29:59 [651a2534-6b4b-42ce-a33e-aec3c81c0408] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.20 ms]
2025/11/02 09:29:59 [95631db7-c089-4642-b829-7bceb730d066] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.24 ms]
2025/11/02 09:29:59 [ca955573-9b5d-4156-a1d8-be28d7e2d131] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.42 ms]
2025/11/02 09:29:59 [a8d95254-2f16-460b-b916-4f9fb71a89aa] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.33 ms]
2025/11/02 09:29:59 [ab32c848-e5fb-45fd-8f2e-bbfcc0d27799] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.31 ms]
2025/11/02 09:29:59 [3d7cd304-cace-49cb-a781-f0ec26e98cd1] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.35 ms]
2025/11/02 09:29:59 [81af4e8e-5c4f-4854-bdee-b742bd95270a] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.29 ms]
2025/11/02 09:29:59 [10c42b13-89f8-435d-957d-160115c0501b] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.16 ms]
2025/11/02 09:29:59 [04fca415-b98f-4280-8e3f-3d30689786b3] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.25 ms]
2025/11/02 09:29:59 [1e5d015f-1af2-4e28-bbac-37cf6f7461f4] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.33 ms]
2025/11/02 09:29:59 [c836da35-96d0-4eeb-bcc8-f61637dc67c1] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.32 ms]
2025/11/02 09:29:59 [fbe86440-08a5-4b80-8f68-1d52c70a4588] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.26 ms]
2025/11/02 09:29:59 [3b315ffc-8cea-4dd5-a01a-a58f3d2c6c71] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.25 ms]
2025/11/02 09:29:59 [2f7ce9fe-dee4-4427-8b82-2fd5623ceb0c] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.12 ms]
2025/11/02 09:29:59 [3c35e3bd-0d1c-4522-80b5-e7fb8da2f5c5] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.14 ms]
2025/11/02 09:29:59 [162ae580-3cf5-4a20-bced-ba83637d5030] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.29 ms]
2025/11/02 09:29:59 [3b1b64e6-5442-4af9-aeca-77eecf646d4d] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.15 ms]
2025/11/02 09:29:59 [77ca907a-e5f5-484f-bd0c-791803d5ad41] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.16 ms]
2025/11/02 09:29:59 [3ccf22b0-a86c-46fd-b6f9-c799ef35176e] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [1.35 ms]
2025/11/02 09:29:59 [31fdd37a-6c1c-4def-a5aa-f7bf79e1f9a5] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.14 ms]
2025/11/02 09:29:59 [97512f16-e947-4a0e-9d84-4b50ab149923] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.20 ms]
2025/11/02 09:29:59 [5e67c590-7b0b-41fb-b893-0f038e730a16] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.22 ms]
2025/11/02 09:29:59 [c208a71d-dc79-497e-9857-b4061bb0a774] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.21 ms]
2025/11/02 09:29:59 [4a4a93ac-1850-4fad-997d-a0efabfe19c0] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.38 ms]
2025/11/02 09:29:59 [b1ad98dc-d6ff-4e7d-8049-2f8789323ff8] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.35 ms]
2025/11/02 09:29:59 [ceac614c-451e-494e-ad19-9035ab691a7d] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.32 ms]
2025/11/02 09:29:59 [d67d6365-1626-4e02-b193-bc3cc20987d7] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.37 ms]
2025/11/02 09:29:59 [9a485aa1-7690-45e9-ad06-b5b448de85a2] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.12 ms]
2025/11/02 09:29:59 [3b261eae-bfac-4a8a-bc02-2cdd8bbe08cd] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.10 ms]
2025/11/02 09:29:59 [e31fc3e9-66ef-4035-bc9a-03365a0b1207] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.32 ms]
2025/11/02 09:29:59 [8d549120-8054-4cc7-b6ff-a8963aa2d43c] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.84 ms]
2025/11/02 09:29:59 [08239003-6ead-427a-bad3-853563d1a3da] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.28 ms]
2025/11/02 09:29:59 [0a4909ae-939f-4ded-acfc-0eeeab656898] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.14 ms]
2025/11/02 09:29:59 [4eb49f8f-4a85-4c2b-96ba-3116464bbb08] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.26 ms]
2025/11/02 09:29:59 [8ad25cd5-ed17-4810-963a-6fe676396c03] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.21 ms]
2025/11/02 09:29:59 [948d4db1-23cf-4b08-8033-865a29e9d0c5] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.11 ms]
2025/11/02 09:29:59 [22a3da64-1c48-49de-87f0-8000bd70f25c] GET /matrixmul -> 503 (Service Unavailable) [PID=2199] [0.23 ms]
2025/11/02 09:30:00 [90f00532-e50f-4c58-ac89-8f11bf3f884b] GET /matrixmul -> 200 (OK) [PID=2199] [114.71 ms]
2025/11/02 09:30:00 [354c9a50-07d1-45b4-ae5b-5c6a376aaa86] GET /matrixmul -> 200 (OK) [PID=2199] [218.33 ms]
2025/11/02 09:30:00 [41953d87-1a2c-431f-a350-d9aadb26f861] GET /matrixmul -> 200 (OK) [PID=2199] [327.76 ms]
    benchmark_test.go:156: 
        [RESULTADOS] Perfil: medium
          Requests totales: 50 (Errores: 0)
          Duración total:   0.33s
          Throughput:       151.63 req/s
          Latencia p50:     0.00 ms
          Latencia p95:     3.00 ms
          Latencia p99:     219.00 ms
    benchmark_test.go:166: [OK] Perfil 'medium' completado correctamente
    benchmark_test.go:89: 
        --- [RUNNING] Perfil de carga: high (100 solicitudes, 10 concurrentes) ---
2025/11/02 09:30:00 [dbcd84a7-49b6-4331-97db-7f36d6841578] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [1.39 ms]
2025/11/02 09:30:00 [f6165c16-67f4-4e1d-80fa-41e4bb35bb33] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [1.82 ms]
2025/11/02 09:30:00 [7ae3ba9f-712a-4856-894f-03fff956bbbe] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [2.10 ms]
2025/11/02 09:30:00 [8e76783d-069d-428e-86ea-4e86be68b725] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.24 ms]
2025/11/02 09:30:00 [33ab32f9-a9c9-4ab3-95cd-7812724d06a9] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [1.54 ms]
2025/11/02 09:30:00 [938b0af2-e55f-4b59-b06a-9cf053a76a74] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [1.74 ms]
2025/11/02 09:30:00 [50128b9a-7a77-4ca6-afa5-30c2245f7d68] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [1.85 ms]
2025/11/02 09:30:00 [38020a5c-39d2-4b2c-af33-21d7be754e43] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [1.91 ms]
2025/11/02 09:30:00 [61bba0f8-51ca-4eec-ac26-ccbb2759c343] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [2.09 ms]
2025/11/02 09:30:00 [65158893-74d3-480f-9673-541e226caa14] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [2.09 ms]
2025/11/02 09:30:00 [0ae1070f-0988-4bd4-af8f-ff3e91565d74] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.13 ms]
2025/11/02 09:30:00 [ad40781a-e81c-40dc-9b63-85c2fd0bcbef] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.18 ms]
2025/11/02 09:30:00 [16648368-722d-4207-8716-d63af940d1ff] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.18 ms]
2025/11/02 09:30:00 [1cf98304-18bd-47b9-ad89-c47c6a566fe1] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.09 ms]
2025/11/02 09:30:00 [1ffe2c54-a891-4ba0-863b-ef8e35924104] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.15 ms]
2025/11/02 09:30:00 [bbf97e1a-ee87-4f5c-a32d-00252ce76aee] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.28 ms]
2025/11/02 09:30:00 [e8f3a886-53a2-4f4d-8fe3-e3eee690513b] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.30 ms]
2025/11/02 09:30:00 [c2345f9e-a2f9-47a3-8a23-54d6f0952110] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.14 ms]
2025/11/02 09:30:00 [590f7c1f-2eaf-4f27-823b-aefd60661ee4] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.08 ms]
2025/11/02 09:30:00 [af5cb2d9-acf3-43df-bdf2-961f8beb4c03] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.33 ms]
2025/11/02 09:30:00 [c2441e7e-7ca4-46a1-bdd6-d9f2219c4986] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.17 ms]
2025/11/02 09:30:00 [946a486e-2d60-4181-ba89-9b63a21abafb] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.14 ms]
2025/11/02 09:30:00 [60db74f5-e472-4d46-bd23-11a0e9a1417f] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.09 ms]
2025/11/02 09:30:00 [7dafa6c7-e7f9-4413-8a78-9dc96ac1667a] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.21 ms]
2025/11/02 09:30:00 [7904f127-1aab-4bac-bdc8-4289f1c80aa3] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.25 ms]
2025/11/02 09:30:00 [6b7f36a6-118e-4a89-9e24-036bdb9cc361] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.29 ms]
2025/11/02 09:30:00 [48bc8cf0-d48e-499b-ad62-9958fafd54ee] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.36 ms]
2025/11/02 09:30:00 [ad39e55f-a0c3-4f39-9473-1313c775ec96] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.43 ms]
2025/11/02 09:30:00 [7bed2a44-b2a2-4396-a013-67b8b25d7af8] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.12 ms]
2025/11/02 09:30:00 [ad5810b0-972d-4f78-b365-4f81ce506a23] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.57 ms]
2025/11/02 09:30:00 [3fd10403-a115-40b9-b96b-ada837254ae4] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.32 ms]
2025/11/02 09:30:00 [205af366-9c97-409a-99d3-789db91d158e] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.17 ms]
2025/11/02 09:30:00 [fa6c8ff5-7b21-4656-91f0-9b6eec6cb080] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.09 ms]
2025/11/02 09:30:00 [8dafead9-fe64-41ef-a93d-8482b1d80555] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.17 ms]
2025/11/02 09:30:00 [85a1567b-8ec1-4a96-8c25-fab2bbb3235b] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.13 ms]
2025/11/02 09:30:00 [d3445015-8b3f-4013-aa29-7905a0dcf445] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.28 ms]
2025/11/02 09:30:00 [6a104b0e-7813-47d3-ba97-89e6a7836c42] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.18 ms]
2025/11/02 09:30:00 [8b35c2a4-df1b-469d-a4b7-37415af3e8d0] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.39 ms]
2025/11/02 09:30:00 [80fa0110-a475-487c-9204-5d50ec5f8a4a] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.31 ms]
2025/11/02 09:30:00 [0da0c76c-9258-4239-a8c2-792b5a9c6f5a] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.68 ms]
2025/11/02 09:30:00 [a5d9d481-316b-4a6e-b193-80497806a52f] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.12 ms]
2025/11/02 09:30:00 [9a4224fc-5751-4279-9ec0-d8777c871004] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.35 ms]
2025/11/02 09:30:00 [7070a94f-72dc-441c-975b-f3670acd76aa] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.15 ms]
2025/11/02 09:30:00 [9691fd25-2367-461b-9ea9-b7d6192d54d7] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.09 ms]
2025/11/02 09:30:00 [cc79ffa6-6e60-4693-a461-e0e61db03d0f] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.08 ms]
2025/11/02 09:30:00 [bf55db4b-bda0-49f7-b34b-a075877ccb06] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.24 ms]
2025/11/02 09:30:00 [4d653eeb-5261-432a-861d-5a5b646486dd] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.76 ms]
2025/11/02 09:30:00 [ec60dd28-3601-4dbe-876b-e8742fda9a51] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.26 ms]
2025/11/02 09:30:00 [0d47de55-9023-4fad-80ac-fb647ad37287] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.25 ms]
2025/11/02 09:30:00 [2e9b6eb4-9236-49cd-8f09-da09becd60e8] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.21 ms]
2025/11/02 09:30:00 [4b41141c-625f-4cc3-9923-fa35ed079e11] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.47 ms]
2025/11/02 09:30:00 [1d9be829-f1de-48a6-af35-d65fef952416] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.09 ms]
2025/11/02 09:30:00 [7aea3bf5-5688-4481-a8f8-e09f7a0598d3] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.10 ms]
2025/11/02 09:30:00 [a9433992-3800-417e-93c5-25bab2a1e9a7] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.36 ms]
2025/11/02 09:30:00 [b1443768-da1a-433c-a580-fe241e22ee20] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.37 ms]
2025/11/02 09:30:00 [ccbd2d7a-44f0-49bb-b72b-129937f80e0b] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.09 ms]
2025/11/02 09:30:00 [27cde878-670d-477f-b89c-8d697ddbb5a8] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.55 ms]
2025/11/02 09:30:00 [89ed5a7e-ae3a-4650-bc2d-ef515f4ea369] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.11 ms]
2025/11/02 09:30:00 [3b4c7fe8-edd1-484e-aee6-3bcf80984e08] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.08 ms]
2025/11/02 09:30:00 [3078e30e-8c1e-40bb-b210-242b1bad7769] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.06 ms]
2025/11/02 09:30:00 [729d73b2-ddbc-4023-933e-9c11612357f3] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.06 ms]
2025/11/02 09:30:00 [13fb34b4-3acb-47ac-a4cb-a99e0ff42920] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.19 ms]
2025/11/02 09:30:00 [ee586a25-c7d8-4f15-ad42-bd9b68dd8376] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.10 ms]
2025/11/02 09:30:00 [237fe42f-c2c9-46c0-ae15-bf3602aecf74] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.38 ms]
2025/11/02 09:30:00 [b1c3365e-1a80-4757-ada8-febbdb302263] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.31 ms]
2025/11/02 09:30:00 [8b4472ab-33d1-4b2f-8895-aa146c2be9c3] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.09 ms]
2025/11/02 09:30:00 [bf121fa3-bf78-4534-a329-d18b6a6b8bb8] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.12 ms]
2025/11/02 09:30:00 [192980b9-4d05-4ede-b31b-ae7adf68b500] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.11 ms]
2025/11/02 09:30:00 [8cceb842-d770-4763-9eea-add31aac0d09] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.11 ms]
2025/11/02 09:30:00 [04b2d9e0-bc44-4796-b406-ebc21eab9e72] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.07 ms]
2025/11/02 09:30:00 [9644270f-e2ca-4cb4-9b99-3beb4654fcf5] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.19 ms]
2025/11/02 09:30:00 [47be8ea5-84a9-4745-924c-0a3da1f4782f] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.19 ms]
2025/11/02 09:30:00 [c1344377-abba-49f4-b48b-55fe3b739ea5] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.08 ms]
2025/11/02 09:30:00 [3fbb41bf-222f-40fc-969f-41baaaa28347] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.42 ms]
2025/11/02 09:30:00 [228cfaf9-8e97-4ef0-8f98-b36851716b46] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.12 ms]
2025/11/02 09:30:00 [01eb288d-5efd-4d28-9983-2ef5b111df3e] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.08 ms]
2025/11/02 09:30:00 [e008c25e-8fd7-41c9-b3d6-f65b3211939e] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.07 ms]
2025/11/02 09:30:00 [afbeafad-e5b1-49f2-9c20-a355e2d7e84e] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.08 ms]
2025/11/02 09:30:00 [8c1c0b36-28ba-4534-90ad-ba3a30cfc018] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.14 ms]
2025/11/02 09:30:00 [fb4af372-1bdf-4be7-ba40-3e641c3572c3] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.17 ms]
2025/11/02 09:30:00 [6b919c67-3dbc-4553-96b6-354e43996956] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.50 ms]
2025/11/02 09:30:00 [5822dd89-ac83-4b8c-83f3-01fb378d9725] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.08 ms]
2025/11/02 09:30:00 [89596348-0702-41ea-bb45-39d9682d22d8] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [1.63 ms]
2025/11/02 09:30:00 [6bcc9d30-953a-42a0-857f-33629e263cd5] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [1.87 ms]
2025/11/02 09:30:00 [1ac68a7b-02e6-4532-b041-0703766092d1] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [1.87 ms]
2025/11/02 09:30:00 [54ad4436-3ea8-4483-8d0c-c66a2a089a60] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.60 ms]
2025/11/02 09:30:00 [6ac57688-4996-445a-8dfc-7f2780f0a4df] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.20 ms]
2025/11/02 09:30:00 [9034bdd1-1320-4a98-9528-526b6a297876] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.27 ms]
2025/11/02 09:30:00 [5ee76fe3-1e90-45c9-bf39-daa5f0ff265c] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.11 ms]
2025/11/02 09:30:00 [cc04dd7c-d22e-4086-a774-74b9bc3650d1] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.43 ms]
2025/11/02 09:30:00 [5129e715-2fd8-47fd-935f-fdd323cd19d0] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.19 ms]
2025/11/02 09:30:00 [f96baf61-27c9-40c1-b4f3-b89deaaafc97] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.08 ms]
2025/11/02 09:30:00 [ebe8e942-718c-4045-a136-ef8b0f46e6c2] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.27 ms]
2025/11/02 09:30:00 [6cc54acb-9cc4-42ca-b035-042869d1dbef] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.24 ms]
2025/11/02 09:30:00 [acf68959-be28-4a91-a775-b9a10843d996] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.06 ms]
2025/11/02 09:30:00 [5c9e87b3-2f21-41ca-998d-4395a239f998] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.14 ms]
2025/11/02 09:30:00 [3bac189d-7c26-42d8-b159-6bbe7ab8ca16] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.12 ms]
2025/11/02 09:30:00 [db56f0e1-570e-4d65-90ae-ff5f4381fd5f] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [1.47 ms]
2025/11/02 09:30:00 [0b04e6d8-e0aa-41a0-a855-bfafe0940967] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [0.15 ms]
2025/11/02 09:30:00 [39817350-83ee-496a-9094-a6f323ae0231] GET /sortfile -> 500 (Internal Server Error) [PID=2199] [1.94 ms]
    benchmark_test.go:156: 
        [RESULTADOS] Perfil: high
          Requests totales: 100 (Errores: 0)
          Duración total:   0.02s
          Throughput:       4271.16 req/s
          Latencia p50:     1.00 ms
          Latencia p95:     5.00 ms
          Latencia p99:     6.00 ms
    benchmark_test.go:166: [OK] Perfil 'high' completado correctamente
    benchmark_test.go:194: 
        [OK] Todos los perfiles de carga completados exitosamente.
--- PASS: TestBenchmark_Profiles (1.03s)
=== RUN   TestConcurrentClients
    concurrency_test.go:50: 
        --- [RUNNING] 20 clientes concurrentes sobre /reverse ---
2025/11/02 09:30:00 [06e4415a-706a-4026-9a5e-179c576dec09] GET /reverse -> 200 (OK) [PID=2199] [0.13 ms]
2025/11/02 09:30:00 [e4e76dfb-b1bf-4c8d-b381-e9aa23bec6c4] GET /reverse -> 200 (OK) [PID=2199] [0.10 ms]
2025/11/02 09:30:00 [966a8d1a-ff27-4f09-ac10-86f4bd675956] GET /reverse -> 200 (OK) [PID=2199] [0.11 ms]
2025/11/02 09:30:00 [7a0aea0a-7ade-4eff-93dc-58e0f53d7832] GET /reverse -> 200 (OK) [PID=2199] [0.08 ms]
2025/11/02 09:30:00 [92b61c82-a134-421f-b854-16c96d0e6974] GET /reverse -> 200 (OK) [PID=2199] [0.06 ms]
2025/11/02 09:30:00 [cd6ffa3f-c4c6-497a-bcc3-e15380a8432d] GET /reverse -> 200 (OK) [PID=2199] [0.07 ms]
2025/11/02 09:30:00 [d3a8bade-69c6-4c3a-9d5b-608f191826d5] GET /reverse -> 200 (OK) [PID=2199] [0.98 ms]
2025/11/02 09:30:00 [31d7db30-6511-4d42-82b8-33f006939416] GET /reverse -> 200 (OK) [PID=2199] [0.06 ms]
2025/11/02 09:30:00 [1e1646f1-6d55-4215-9f65-7fb7a8bfa063] GET /reverse -> 200 (OK) [PID=2199] [0.33 ms]
2025/11/02 09:30:00 [48a7049a-b610-4cbc-8777-9b69784a2a2e] GET /reverse -> 200 (OK) [PID=2199] [0.31 ms]
2025/11/02 09:30:00 [a4b14df4-9e5e-49c0-b65c-1e1276a463d4] GET /reverse -> 200 (OK) [PID=2199] [0.08 ms]
2025/11/02 09:30:00 [3c1dbbd2-76cf-4540-a614-15b071cf44af] GET /reverse -> 200 (OK) [PID=2199] [0.09 ms]
2025/11/02 09:30:00 [4c4444e8-c8df-41a4-98b8-61ea2a0964fa] GET /reverse -> 200 (OK) [PID=2199] [0.06 ms]
2025/11/02 09:30:00 [de9dc45d-71d2-449e-930e-235c9d3c7f41] GET /reverse -> 200 (OK) [PID=2199] [0.11 ms]
2025/11/02 09:30:00 [b35ff877-851c-4c4a-bc0a-121818d73765] GET /reverse -> 200 (OK) [PID=2199] [0.60 ms]
2025/11/02 09:30:00 [0f02a63e-77de-46aa-907f-d0c9f8fa6527] GET /reverse -> 200 (OK) [PID=2199] [0.09 ms]
2025/11/02 09:30:00 [0e10047b-bfd2-4b65-8cae-b8a42777393e] GET /reverse -> 200 (OK) [PID=2199] [1.68 ms]
2025/11/02 09:30:00 [5790e4d1-9ef3-4002-8966-82a8216aba35] GET /reverse -> 200 (OK) [PID=2199] [0.06 ms]
2025/11/02 09:30:00 [3579122b-68dc-4787-aa1e-56e0c100eefe] GET /reverse -> 200 (OK) [PID=2199] [2.44 ms]
2025/11/02 09:30:00 [e230d914-861c-45c7-9032-851f4c5bfce7] GET /reverse -> 200 (OK) [PID=2199] [1.50 ms]
    concurrency_test.go:90: [RESULT] 20 concurrent requests completed in 3.928626ms (errors: 0)
    concurrency_test.go:92: [OK] No se detectaron errores de concurrencia
--- PASS: TestConcurrentClients (0.00s)
=== RUN   TestJobQueuePressure
    concurrency_test.go:117: 
        --- [RUNNING] Saturación de cola con 30 jobs ---
2025/11/02 09:30:00 [0e197f12-45a8-4a95-9608-4268ef426c6c] GET /jobs/submit -> 200 (OK) [PID=2199] [13.28 ms]
2025/11/02 09:30:00 [37a4ae70-d9a1-48d5-88e8-2f14210263c0] GET /jobs/submit -> 200 (OK) [PID=2199] [2.86 ms]
2025/11/02 09:30:00 [fa3ca564-0972-4472-81df-8073fc1b644c] GET /jobs/submit -> 200 (OK) [PID=2199] [3.71 ms]
2025/11/02 09:30:00 [38257da9-994b-426e-a366-c6ded97115de] GET /jobs/submit -> 200 (OK) [PID=2199] [3.30 ms]
2025/11/02 09:30:00 [0a028ffc-871e-4b54-81c0-8a330f8f717b] GET /jobs/submit -> 200 (OK) [PID=2199] [3.74 ms]
2025/11/02 09:30:00 [97d9b78a-ada1-4335-a380-5171e11d00dd] GET /jobs/submit -> 200 (OK) [PID=2199] [2.45 ms]
2025/11/02 09:30:00 [9fe5178e-ad44-46b7-aec5-abe57ab78536] GET /jobs/submit -> 200 (OK) [PID=2199] [2.98 ms]
2025/11/02 09:30:00 [e7ac276c-c1f5-4ccd-9fea-2c026801a19f] GET /jobs/submit -> 200 (OK) [PID=2199] [3.18 ms]
2025/11/02 09:30:00 [87d8893e-a1a5-4677-8b90-a489f6a2dabf] GET /jobs/submit -> 200 (OK) [PID=2199] [4.00 ms]
2025/11/02 09:30:00 [b49802fd-2279-4f02-9d8b-ec23cacdc359] GET /jobs/submit -> 200 (OK) [PID=2199] [3.50 ms]
2025/11/02 09:30:00 [94d05091-b02d-4b3f-a312-2154e1972135] GET /jobs/submit -> 200 (OK) [PID=2199] [3.23 ms]
2025/11/02 09:30:00 [e837dd72-9d29-445a-9239-538d385add29] GET /jobs/submit -> 200 (OK) [PID=2199] [7.15 ms]
2025/11/02 09:30:00 [8410c2b3-2de9-4e2a-81dd-d260d4eaef25] GET /jobs/submit -> 200 (OK) [PID=2199] [9.01 ms]
2025/11/02 09:30:00 [c98aa3df-6c4d-4a0a-b07c-08a0088090e7] GET /jobs/submit -> 200 (OK) [PID=2199] [14.04 ms]
2025/11/02 09:30:00 [d41e0c30-073e-4bf6-a372-6c8b76996e40] GET /jobs/submit -> 200 (OK) [PID=2199] [12.40 ms]
2025/11/02 09:30:00 [24b8735a-dfb7-4a5d-8ebf-d20c140813a8] GET /jobs/submit -> 200 (OK) [PID=2199] [10.34 ms]
2025/11/02 09:30:00 [d4eb37be-bc47-43a3-88ee-4177f808ea75] GET /jobs/submit -> 200 (OK) [PID=2199] [12.42 ms]
2025/11/02 09:30:00 [559b76fb-8302-4d81-a238-e42cf934ef94] GET /jobs/submit -> 200 (OK) [PID=2199] [11.91 ms]
2025/11/02 09:30:00 [fddcaedd-c086-4fed-a872-61e7434e0c35] GET /jobs/submit -> 200 (OK) [PID=2199] [11.31 ms]
2025/11/02 09:30:00 [c73b46ed-9817-4446-b707-6803bc86faec] GET /jobs/submit -> 200 (OK) [PID=2199] [10.93 ms]
2025/11/02 09:30:00 [ce8336c3-274e-403e-9d02-1f97f6d9a1bb] GET /jobs/submit -> 200 (OK) [PID=2199] [15.80 ms]
2025/11/02 09:30:00 [76ace207-3345-4fc7-8881-8c6240cfa56d] GET /jobs/submit -> 200 (OK) [PID=2199] [11.41 ms]
2025/11/02 09:30:00 [2e47ee66-204f-4c61-b515-7422d8d32369] GET /jobs/submit -> 200 (OK) [PID=2199] [12.29 ms]
2025/11/02 09:30:00 [fdbe1604-bd91-49c4-a77f-455334578401] GET /jobs/submit -> 200 (OK) [PID=2199] [12.20 ms]
2025/11/02 09:30:00 [c0debfb6-4060-4f67-a0d8-bec2f54671b3] GET /jobs/submit -> 200 (OK) [PID=2199] [12.92 ms]
2025/11/02 09:30:00 [f1a9766f-9e94-4db4-b154-3141c288594e] GET /jobs/submit -> 200 (OK) [PID=2199] [10.70 ms]
2025/11/02 09:30:00 [7e13ba4f-8529-4c0d-bdda-48f398e3e62f] GET /jobs/submit -> 200 (OK) [PID=2199] [10.69 ms]
2025/11/02 09:30:00 [5b3b5b96-d108-4584-87cc-a4cc5b261559] GET /jobs/submit -> 200 (OK) [PID=2199] [11.84 ms]
2025/11/02 09:30:00 [2258a2f6-579b-4621-b479-d13272439f76] GET /jobs/submit -> 200 (OK) [PID=2199] [12.54 ms]
2025/11/02 09:30:00 [e177be85-83a1-4c81-ac19-d7616e0ee276] GET /jobs/submit -> 200 (OK) [PID=2199] [10.29 ms]
    concurrency_test.go:142: [RESULT] Jobs OK: 30, Queue full (503): 0
    concurrency_test.go:149: [WARN] Cola no saturada — ajustar MAX_TOTAL o reducir límites para pruebas realistas
--- PASS: TestJobQueuePressure (0.29s)
=== RUN   TestMetricsAndStatusDuringLoad
    concurrency_test.go:170: 
        --- [RUNNING] Validando /metrics y /status durante ejecución de jobs ---
2025/11/02 09:30:00 [6aa9cf65-6829-47b8-970c-3a1aa97b4d38] GET /jobs/submit -> 200 (OK) [PID=2199] [10.57 ms]
2025/11/02 09:30:00 [0f998c9d-e275-40b5-b085-49a7048869be] GET /jobs/submit -> 200 (OK) [PID=2199] [13.11 ms]
2025/11/02 09:30:00 [68e2cf1b-7540-47f5-8d77-461ebe394a3a] GET /jobs/submit -> 200 (OK) [PID=2199] [12.31 ms]
2025/11/02 09:30:00 [c3e5e327-73df-42a1-b070-c63a724d4976] GET /jobs/submit -> 200 (OK) [PID=2199] [13.83 ms]
2025/11/02 09:30:00 [4a7a9950-4cf3-4db2-9bfb-8849faa85785] GET /jobs/submit -> 200 (OK) [PID=2199] [11.72 ms]
2025/11/02 09:30:00 [ed5304e7-18fa-436f-95a7-68280c5e834c] GET /metrics -> 200 (OK) [PID=2199] [0.86 ms]
2025/11/02 09:30:00 [6312b516-5822-42f7-9278-0c437fe0e182] GET /status -> 200 (OK) [PID=2199] [0.83 ms]
2025/11/02 09:30:01 [f91c8386-4983-4781-8287-8abafb0b3318] GET /metrics -> 200 (OK) [PID=2199] [0.64 ms]
2025/11/02 09:30:01 [08d82585-9633-4acd-bb79-7d2f7e0251d8] GET /status -> 200 (OK) [PID=2199] [0.46 ms]
2025/11/02 09:30:01 [e3c8d005-2a10-4eba-ade4-4bf8e4aea56a] GET /metrics -> 200 (OK) [PID=2199] [0.36 ms]
2025/11/02 09:30:01 [aa2d0b2f-a2ac-440f-9935-8816ef6874ac] GET /status -> 200 (OK) [PID=2199] [0.35 ms]
    concurrency_test.go:198: [OK] /metrics y /status permanecieron operativos bajo carga
--- PASS: TestMetricsAndStatusDuringLoad (1.58s)
=== RUN   TestCancelMultipleJobs
    concurrency_test.go:221: 
        --- [RUNNING] Cancelación simultánea de 5 jobs ---
2025/11/02 09:30:02 [46d4a572-b619-4dd7-a7bd-170e6bcf77ec] GET /jobs/submit -> 200 (OK) [PID=2199] [10.02 ms]
2025/11/02 09:30:02 [d495f3ec-41cc-41fc-92f2-a801e00fdd41] GET /jobs/submit -> 200 (OK) [PID=2199] [11.05 ms]
2025/11/02 09:30:02 [2edc12d2-ba18-4912-a7c1-2aae334d0bd9] GET /jobs/submit -> 200 (OK) [PID=2199] [8.99 ms]
2025/11/02 09:30:02 [c574f20c-8b83-4507-a621-306c34c367fe] GET /jobs/submit -> 200 (OK) [PID=2199] [8.01 ms]
2025/11/02 09:30:02 [9eb861f6-e004-4c53-901b-f3375f60834c] GET /jobs/submit -> 200 (OK) [PID=2199] [8.98 ms]
2025/11/02 09:30:02 [5413e45e-fd06-4996-8239-098421f6631b] GET /jobs/cancel -> 200 (OK) [PID=2199] [10.79 ms]
2025/11/02 09:30:02 [c63142bd-cfc9-4f4b-b0b5-e1e12a63bb11] GET /jobs/cancel -> 200 (OK) [PID=2199] [13.79 ms]
2025/11/02 09:30:02 [2de943d6-51f0-4e28-802f-b2fbe630c13f] GET /jobs/cancel -> 409 (Conflict) [PID=2199] [13.78 ms]
2025/11/02 09:30:02 [6c28ebe4-eacc-4bac-b447-ff886820e220] GET /jobs/cancel -> 200 (OK) [PID=2199] [18.30 ms]
2025/11/02 09:30:02 [78026a8b-fcc1-4a03-ad2d-6c6af840d5ed] GET /jobs/cancel -> 200 (OK) [PID=2199] [22.54 ms]
2025/11/02 09:30:02 [d1b5fbaa-8c16-4f67-9bc7-f659264f59da] GET /jobs/status -> 200 (OK) [PID=2199] [4.54 ms]
2025/11/02 09:30:02 [230f211f-a2d1-49c5-846f-da6572c01771] GET /jobs/status -> 200 (OK) [PID=2199] [5.92 ms]
2025/11/02 09:30:02 [60a9b338-db86-4eef-a13b-ca5f7b263870] GET /jobs/status -> 200 (OK) [PID=2199] [5.66 ms]
2025/11/02 09:30:02 [370348c5-74f8-4be3-8f40-cbefdb0e2150] GET /jobs/status -> 200 (OK) [PID=2199] [7.28 ms]
2025/11/02 09:30:02 [eb16ff4f-4979-4930-a0b2-dcab45eb1a8c] GET /jobs/status -> 200 (OK) [PID=2199] [6.03 ms]
    concurrency_test.go:261: [OK] Múltiples cancelaciones concurrentes manejadas sin errores
--- PASS: TestCancelMultipleJobs (0.41s)
=== RUN   TestServer_ReverseEndpoint
2025/11/02 09:30:02 [b4f08330-3cee-437d-a932-e40fc5257473] GET /reverse -> 200 (OK) [PID=2199] [0.19 ms]
--- PASS: TestServer_ReverseEndpoint (0.00s)
=== RUN   TestServer_ToUpper
2025/11/02 09:30:02 [d2b820f8-4487-406a-9f82-876120324459] GET /toupper -> 200 (OK) [PID=2199] [0.23 ms]
--- PASS: TestServer_ToUpper (0.00s)
=== RUN   TestServer_StatusMetricsHelp
2025/11/02 09:30:02 [9cf5baa0-34fb-40ea-b64a-1b4705b5f593] GET /status -> 200 (OK) [PID=2199] [0.24 ms]
2025/11/02 09:30:02 [3faf17a1-6c22-42e8-9680-89534aa2ea35] GET /metrics -> 200 (OK) [PID=2199] [1.40 ms]
2025/11/02 09:30:02 [12e78c0c-93e1-4d9c-98c5-2e9ed275ef3b] GET /help -> 200 (OK) [PID=2199] [0.32 ms]
--- PASS: TestServer_StatusMetricsHelp (0.00s)
=== RUN   TestServer_CreateDeleteFile
2025/11/02 09:30:02 [a4c65594-5115-4313-897a-2c1e469b3d45] GET /createfile -> 200 (OK) [PID=2199] [0.42 ms]
2025/11/02 09:30:02 [48c67ef8-3e29-4179-ac04-aa46ef180adc] GET /deletefile -> 200 (OK) [PID=2199] [0.35 ms]
--- PASS: TestServer_CreateDeleteFile (0.00s)
=== RUN   TestServer_FibonacciIntegration
2025/11/02 09:30:02 [7cbf339e-1c9b-4017-a7b0-de5ee6eaa752] GET /fibonacci -> 200 (OK) [PID=2199] [0.16 ms]
--- PASS: TestServer_FibonacciIntegration (0.00s)
=== RUN   TestServer_InvalidPath
2025/11/02 09:30:02 [f4f95f21-c538-44a4-862d-2347d867f868] GET /notfound -> 404 (0.16 ms)
--- PASS: TestServer_InvalidPath (0.00s)
=== RUN   TestJobs_SubmitStatusResultFlow
2025/11/02 09:30:02 [57bbacd2-a229-463a-857f-ed956f0e8ce6] GET /jobs/submit -> 200 (OK) [PID=2199] [7.32 ms]
2025/11/02 09:30:03 [8cd2ea1a-a025-4c97-9a2f-b903d09ffd2c] GET /jobs/status -> 200 (OK) [PID=2199] [7.30 ms]
2025/11/02 09:30:03 [8b307408-9c61-4f2f-8935-72b666e10b32] GET /jobs/result -> 200 (OK) [PID=2199] [7.63 ms]
--- PASS: TestJobs_SubmitStatusResultFlow (1.03s)
=== RUN   TestJobs_Cancel
2025/11/02 09:30:03 [9f6c7eab-842b-4ab9-ad47-68f0d2f68ffe] GET /jobs/submit -> 200 (OK) [PID=2199] [11.94 ms]
2025/11/02 09:30:03 [fc44c170-1ea9-47ee-abbf-fdb7ca289d89] GET /jobs/cancel -> 200 (OK) [PID=2199] [8.43 ms]
2025/11/02 09:30:04 [cd26b782-350e-4596-b021-fb4d4155ccc1] GET /jobs/status -> 200 (OK) [PID=2199] [0.45 ms]
--- PASS: TestJobs_Cancel (0.63s)
=== RUN   TestServer_ConcurrentRequests
    integration_test.go:378: 
        --- [RUNNING] 10 concurrent requests to /reverse ---
2025/11/02 09:30:04 [5e571e02-fdc6-47ed-89fa-90f6c3254d2d] GET /reverse -> 200 (OK) [PID=2199] [0.09 ms]
2025/11/02 09:30:04 [00e871bb-b3a0-49f1-ae9b-57a4a69e7363] GET /reverse -> 200 (OK) [PID=2199] [0.14 ms]
2025/11/02 09:30:04 [8946b88a-db6f-4e09-b2ea-0a5e1f8f8ac8] GET /reverse -> 200 (OK) [PID=2199] [0.20 ms]
2025/11/02 09:30:04 [d36c65e8-95a2-4a44-b583-ada92785952f] GET /reverse -> 200 (OK) [PID=2199] [0.10 ms]
2025/11/02 09:30:04 [48a095dd-0b98-4567-ae67-e68acb4abc2f] GET /reverse -> 200 (OK) [PID=2199] [0.12 ms]
2025/11/02 09:30:04 [6f782948-151c-47c1-bd7f-25a34ed54c35] GET /reverse -> 200 (OK) [PID=2199] [0.08 ms]
2025/11/02 09:30:04 [35d964ac-4ad4-4b05-b752-691065d667ab] GET /reverse -> 200 (OK) [PID=2199] [0.21 ms]
2025/11/02 09:30:04 [030b013f-4c38-44aa-bc9b-9b57df31cfa6] GET /reverse -> 200 (OK) [PID=2199] [0.15 ms]
2025/11/02 09:30:04 [e40f7810-52bf-4930-9a3a-0454ae054c55] GET /reverse -> 200 (OK) [PID=2199] [0.10 ms]
2025/11/02 09:30:04 [256d9f2c-9572-45a1-88bd-ad58a0306f7f] GET /reverse -> 200 (OK) [PID=2199] [0.11 ms]
    integration_test.go:419: [OK] All concurrent requests handled correctly (200/503)
--- PASS: TestServer_ConcurrentRequests (0.00s)
=== RUN   TestPool_GetPoolInfo_And_DefaultTimeout
--- PASS: TestPool_GetPoolInfo_And_DefaultTimeout (0.00s)
=== RUN   TestPool_SubmitAndWait_SuccessAndTimeout
--- PASS: TestPool_SubmitAndWait_SuccessAndTimeout (30.00s)
=== RUN   TestPool_InitPool_Duplicate
--- PASS: TestPool_InitPool_Duplicate (0.00s)
=== RUN   TestSetTimeout_ModifiesDefault
--- PASS: TestSetTimeout_ModifiesDefault (0.00s)
=== RUN   TestPerformance_PiHeavy
    performance_test.go:108: 
        --- [RUNNING] Pi(100000 dígitos) ---
2025/11/02 09:30:38 [edefa924-ea68-4eaf-a6c5-74518c81569c] GET /pi -> 200 (OK) [PID=2199] [4702.36 ms]
    performance_test.go:117: [DONE] Pi(100000 dígitos) completado en 4.704148665s
--- PASS: TestPerformance_PiHeavy (4.70s)
=== RUN   TestPerformance_MatrixMulHeavy
    performance_test.go:108: 
        --- [RUNNING] MatrixMul(size=1000, seed=7) ---
2025/11/02 09:30:46 [52248375-d8ce-4d75-94e2-3604b5bf6cc1] GET /matrixmul -> 200 (OK) [PID=2199] [8000.30 ms]
    performance_test.go:117: [DONE] MatrixMul(size=1000, seed=7) completado en 8.001086042s
--- PASS: TestPerformance_MatrixMulHeavy (8.00s)
=== RUN   TestPerformance_MandelbrotHeavy
    performance_test.go:108: 
        --- [RUNNING] Mandelbrot(3840x2160, max_iter=1200) ---
    performance_test.go:117: [DONE] Mandelbrot(3840x2160, max_iter=1200) completado en 14.427985168s
--- PASS: TestPerformance_MandelbrotHeavy (14.43s)
=== RUN   TestPerformance_SortFile
2025/11/02 09:31:01 [70f6c004-481c-400a-964a-b342a61d2aeb] GET /mandelbrot -> 200 (OK) [PID=2199] [14427.74 ms]
    performance_test.go:108: 
        --- [RUNNING] SortFile(≈10000000 líneas) ---
2025/11/02 09:31:22 [e55553d9-b9db-46b6-8e10-04379ed647ef] GET /sortfile -> 200 (OK) [PID=2199] [4404.51 ms]
    performance_test.go:117: [DONE] SortFile(≈10000000 líneas) completado en 4.405326713s
--- PASS: TestPerformance_SortFile (21.18s)
=== RUN   TestPerformance_Compress
    performance_test.go:108: 
        --- [RUNNING] Compress(gzip, ≈10000000 repeticiones) ---
2025/11/02 09:31:25 [4a661346-2344-42b1-9b0d-53dca8a35c92] GET /compress -> 200 (OK) [PID=2199] [1724.31 ms]
    performance_test.go:117: [DONE] Compress(gzip, ≈10000000 repeticiones) completado en 1.725263322s
    performance_test.go:122: [WARN] Compress(gzip, ≈10000000 repeticiones) se ejecutó demasiado rápido (posible entorno limitado)
--- PASS: TestPerformance_Compress (2.69s)
=== RUN   TestPerformance_HashFile
    performance_test.go:108: 
        --- [RUNNING] HashFile(sha256, ≈20000000 líneas) ---
2025/11/02 09:31:25 [60def0c0-ec59-43d8-803a-7988b69a2d07] GET /hashfile -> 200 (OK) [PID=2199] [209.88 ms]
    performance_test.go:117: [DONE] HashFile(sha256, ≈20000000 líneas) completado en 211.079589ms
    performance_test.go:122: [WARN] HashFile(sha256, ≈20000000 líneas) se ejecutó demasiado rápido (posible entorno limitado)
--- PASS: TestPerformance_HashFile (0.44s)
=== RUN   TestPerformance_MetricsAfterHeavyLoad
2025/11/02 09:31:25 [96a7bf58-3cc9-48b8-9a93-a2fbcaf23f40] GET /metrics -> 200 (OK) [PID=2199] [0.71 ms]
    performance_test.go:374: [OK] /metrics respondió correctamente tras pruebas de carga pesada (CPU & IO)
--- PASS: TestPerformance_MetricsAfterHeavyLoad (0.00s)
=== RUN   TestReverse_Success
    unit_test.go:78: 
        --- [TestReverse_Success] Reversa de texto válido ---
--- PASS: TestReverse_Success (0.00s)
=== RUN   TestReverse_EmptyText
    unit_test.go:78: 
        --- [TestReverse_EmptyText] Reversa con texto vacío debe devolver 400 ---
--- PASS: TestReverse_EmptyText (0.00s)
=== RUN   TestToUpper_Success
    unit_test.go:78: 
        --- [TestToUpper_Success] Conversión a mayúsculas con entrada válida ---
--- PASS: TestToUpper_Success (0.00s)
=== RUN   TestToUpper_MissingParam
    unit_test.go:78: 
        --- [TestToUpper_MissingParam] Falta parámetro 'text' debe devolver 400 ---
--- PASS: TestToUpper_MissingParam (0.00s)
=== RUN   TestHash_Success
    unit_test.go:78: 
        --- [TestHash_Success] Hash de texto válido produce sha256 ---
--- PASS: TestHash_Success (0.00s)
=== RUN   TestHash_Empty
    unit_test.go:78: 
        --- [TestHash_Empty] Hash con texto vacío debe devolver 400 ---
--- PASS: TestHash_Empty (0.00s)
=== RUN   TestRandom_Success
    unit_test.go:78: 
        --- [TestRandom_Success] Generación de números aleatorios válida ---
--- PASS: TestRandom_Success (0.00s)
=== RUN   TestRandom_InvalidRange
    unit_test.go:78: 
        --- [TestRandom_InvalidRange] Rango inválido (min>max) debe devolver 400 ---
--- PASS: TestRandom_InvalidRange (0.00s)
=== RUN   TestTimestamp_Success
    unit_test.go:78: 
        --- [TestTimestamp_Success] Timestamp debe incluir campo 'iso' ---
--- PASS: TestTimestamp_Success (0.00s)
=== RUN   TestTimestamp_Cancelled
    unit_test.go:78: 
        --- [TestTimestamp_Cancelled] Cancelación previa al inicio (sin invocación) ---
--- PASS: TestTimestamp_Cancelled (0.00s)
=== RUN   TestSimulate_Success
    unit_test.go:78: 
        --- [TestSimulate_Success] Simulación de trabajo exitosa ---
--- PASS: TestSimulate_Success (1.00s)
=== RUN   TestSimulate_Cancelled
    unit_test.go:78: 
        --- [TestSimulate_Cancelled] Simulación cancelada debe devolver 499 ---
--- PASS: TestSimulate_Cancelled (1.00s)
=== RUN   TestSleep_Success
    unit_test.go:78: 
        --- [TestSleep_Success] Sleep con segundos válidos ---
--- PASS: TestSleep_Success (1.00s)
=== RUN   TestSleep_Cancel
    unit_test.go:78: 
        --- [TestSleep_Cancel] Sleep cancelado debe devolver 499 ---
--- PASS: TestSleep_Cancel (1.00s)
=== RUN   TestSleep_InvalidSeconds
    unit_test.go:78: 
        --- [TestSleep_InvalidSeconds] Segundos inválidos (<=0) debe devolver 400 ---
--- PASS: TestSleep_InvalidSeconds (0.00s)
=== RUN   TestSleep_CancelledBeforeStart
    unit_test.go:78: 
        --- [TestSleep_CancelledBeforeStart] Cancelación previa al inicio debe devolver 499 ---
--- PASS: TestSleep_CancelledBeforeStart (0.00s)
=== RUN   TestLoadTest_Success
    unit_test.go:78: 
        --- [TestLoadTest_Success] Ejecución de tareas concurrentes válida ---
--- PASS: TestLoadTest_Success (1.00s)
=== RUN   TestLoadTest_InvalidParams
    unit_test.go:78: 
        --- [TestLoadTest_InvalidParams] Parámetros inválidos en LoadTest (tasks=0) ---
--- PASS: TestLoadTest_InvalidParams (0.00s)
=== RUN   TestLoadTest_NegativeSleep
    unit_test.go:78: 
        --- [TestLoadTest_NegativeSleep] Parámetro sleep negativo debe devolver 400 ---
--- PASS: TestLoadTest_NegativeSleep (0.00s)
=== RUN   TestLoadTest_CancelledBeforeStart
    unit_test.go:78: 
        --- [TestLoadTest_CancelledBeforeStart] Cancelación antes de iniciar debe devolver 499 ---
--- PASS: TestLoadTest_CancelledBeforeStart (0.00s)
=== RUN   TestLoadTest_CancelledDuringExecution
    unit_test.go:78: 
        --- [TestLoadTest_CancelledDuringExecution] Cancelación durante ejecución (sin llamada, conserva intención original) ---
--- PASS: TestLoadTest_CancelledDuringExecution (0.00s)
=== RUN   TestHashText_Consistency
    unit_test.go:78: 
        --- [TestHashText_Consistency] Mismo input debe producir mismo hash ---
--- PASS: TestHashText_Consistency (0.00s)
=== RUN   TestTimestamp_NotEmpty
    unit_test.go:78: 
        --- [TestTimestamp_NotEmpty] Body del timestamp no debe estar vacío ---
--- PASS: TestTimestamp_NotEmpty (0.00s)
=== RUN   TestReverse_Cancelled
    unit_test.go:78: 
        --- [TestReverse_Cancelled] Cancelación previa o bad request ---
--- PASS: TestReverse_Cancelled (0.00s)
=== RUN   TestCreateTempFileForFutureIO
    unit_test.go:78: 
        --- [TestCreateTempFileForFutureIO] Crear/eliminar archivo temporal ---
--- PASS: TestCreateTempFileForFutureIO (0.00s)
=== RUN   TestFibonacci_Valid
    unit_test.go:78: 
        --- [TestFibonacci_Valid] Serie de Fibonacci válida ---
--- PASS: TestFibonacci_Valid (0.00s)
=== RUN   TestFibonacci_Invalid
    unit_test.go:78: 
        --- [TestFibonacci_Invalid] n inválido debe devolver 400 ---
--- PASS: TestFibonacci_Invalid (0.00s)
=== RUN   TestIsPrime_TrialMethod
    unit_test.go:78: 
        --- [TestIsPrime_TrialMethod] Primalidad con método 'trial' ---
--- PASS: TestIsPrime_TrialMethod (0.00s)
=== RUN   TestIsPrime_TrialMethod2
    unit_test.go:78: 
        --- [TestIsPrime_TrialMethod2] Primalidad para 2 con método 'trial' ---
--- PASS: TestIsPrime_TrialMethod2 (0.00s)
=== RUN   TestIsPrime_TrialMethod3
    unit_test.go:78: 
        --- [TestIsPrime_TrialMethod3] No primalidad para 4 con 'trial' ---
--- PASS: TestIsPrime_TrialMethod3 (0.00s)
=== RUN   TestIsPrime_TrialMethod4
    unit_test.go:78: 
        --- [TestIsPrime_TrialMethod4] Valores <2 no son primos (trial) ---
--- PASS: TestIsPrime_TrialMethod4 (0.00s)
=== RUN   TestIsPrime_TrialMethod5
    unit_test.go:78: 
        --- [TestIsPrime_TrialMethod5] No primalidad para 35 (trial) ---
--- PASS: TestIsPrime_TrialMethod5 (0.00s)
=== RUN   TestIsPrime_MillerMethod
    unit_test.go:78: 
        --- [TestIsPrime_MillerMethod] Primalidad con método 'miller' ---
--- PASS: TestIsPrime_MillerMethod (0.00s)
=== RUN   TestIsPrime_InvalidMethod
    unit_test.go:78: 
        --- [TestIsPrime_InvalidMethod] Método inválido debe devolver 400 ---
--- PASS: TestIsPrime_InvalidMethod (0.00s)
=== RUN   TestIsPrime_DefaultMethod
    unit_test.go:78: 
        --- [TestIsPrime_DefaultMethod] Sin método => usa 'trial' ---
--- PASS: TestIsPrime_DefaultMethod (0.00s)
=== RUN   TestIsPrime_CancelledImmediately
    unit_test.go:78: 
        --- [TestIsPrime_CancelledImmediately] Cancelación inmediata debe devolver 499 ---
--- PASS: TestIsPrime_CancelledImmediately (0.00s)
=== RUN   TestIsPrime_CancelDuringTrial
    unit_test.go:78: 
        --- [TestIsPrime_CancelDuringTrial] Cancel durante trial en número grande ---
--- PASS: TestIsPrime_CancelDuringTrial (0.00s)
=== RUN   TestIsPrime_CancelDuringMiller
    unit_test.go:78: 
        --- [TestIsPrime_CancelDuringMiller] Cancel durante Miller-Rabin ---
--- PASS: TestIsPrime_CancelDuringMiller (0.00s)
=== RUN   TestIsPrime_MillerCompositeBranches
    unit_test.go:78: 
        --- [TestIsPrime_MillerCompositeBranches] Cobertura de ramas compuestas (Miller) ---
--- PASS: TestIsPrime_MillerCompositeBranches (0.00s)
=== RUN   TestFactorNumber_Valid
    unit_test.go:78: 
        --- [TestFactorNumber_Valid] Factorización de 84 ---
--- PASS: TestFactorNumber_Valid (0.00s)
=== RUN   TestFactorNumber_Valid2
    unit_test.go:78: 
        --- [TestFactorNumber_Valid2] Factorización de 15 ---
--- PASS: TestFactorNumber_Valid2 (0.00s)
=== RUN   TestFactorNumber_PrimeInput
    unit_test.go:78: 
        --- [TestFactorNumber_PrimeInput] Factorización de primo debe ser [n] ---
--- PASS: TestFactorNumber_PrimeInput (0.00s)
=== RUN   TestFactorNumber_Invalid
    unit_test.go:78: 
        --- [TestFactorNumber_Invalid] Entrada negativa debe devolver 400 ---
--- PASS: TestFactorNumber_Invalid (0.00s)
=== RUN   TestMatrixMultiply_Valid
    unit_test.go:78: 
        --- [TestMatrixMultiply_Valid] Multiplicación de matrices válida ---
--- PASS: TestMatrixMultiply_Valid (0.00s)
=== RUN   TestMatrixMultiply_InvalidSize
    unit_test.go:78: 
        --- [TestMatrixMultiply_InvalidSize] Size inválido debe devolver 400 ---
--- PASS: TestMatrixMultiply_InvalidSize (0.00s)
=== RUN   TestMandelbrot_Valid
    unit_test.go:78: 
        --- [TestMandelbrot_Valid] Generación válida de fractal ---
--- PASS: TestMandelbrot_Valid (0.00s)
=== RUN   TestMandelbrot_InvalidParams
    unit_test.go:78: 
        --- [TestMandelbrot_InvalidParams] Parámetros inválidos deben devolver 400 ---
--- PASS: TestMandelbrot_InvalidParams (0.00s)
=== RUN   TestPi_Valid
    unit_test.go:78: 
        --- [TestPi_Valid] Cálculo de PI con dígitos válidos ---
--- PASS: TestPi_Valid (0.00s)
=== RUN   TestPi_InvalidDigits
    unit_test.go:78: 
        --- [TestPi_InvalidDigits] Dígitos inválidos deben devolver 400 ---
--- PASS: TestPi_InvalidDigits (0.00s)
=== RUN   TestFibonacci_Cancelled
    unit_test.go:78: 
        --- [TestFibonacci_Cancelled] Cancelación durante Fibonacci ---
--- PASS: TestFibonacci_Cancelled (0.00s)
=== RUN   TestMatrixMultiply_Cancelled
    unit_test.go:78: 
        --- [TestMatrixMultiply_Cancelled] Cancelación durante multiplicación de matrices grande ---
--- PASS: TestMatrixMultiply_Cancelled (0.02s)
=== RUN   TestIsPrime_LargeNumber
    unit_test.go:78: 
        --- [TestIsPrime_LargeNumber] Prueba con número grande (bordes) ---
--- PASS: TestIsPrime_LargeNumber (0.00s)
=== RUN   TestMandelbrot_SaveFileTrue
    unit_test.go:78: 
        --- [TestMandelbrot_SaveFileTrue] Generación con volcado a archivo ---
--- PASS: TestMandelbrot_SaveFileTrue (0.00s)
=== RUN   TestPi_Consistency
    unit_test.go:78: 
        --- [TestPi_Consistency] Determinismo en cálculo de PI ---
--- PASS: TestPi_Consistency (0.00s)
=== RUN   TestMatrixMul_SeedEffect
    unit_test.go:78: 
        --- [TestMatrixMul_SeedEffect] Seeds diferentes deben producir hashes distintos ---
--- PASS: TestMatrixMul_SeedEffect (0.00s)
=== RUN   TestFactorNumber_StringConversion
    unit_test.go:78: 
        --- [TestFactorNumber_StringConversion] Sanity check de strconv.Itoa ---
--- PASS: TestFactorNumber_StringConversion (0.00s)
=== RUN   TestCreateFile_Success
    unit_test.go:78: 
        --- [TestCreateFile_Success] Creación de archivo válida ---
--- PASS: TestCreateFile_Success (0.00s)
=== RUN   TestCreateFile_Invalid
    unit_test.go:78: 
        --- [TestCreateFile_Invalid] Falta de nombre de archivo debe devolver 400 ---
--- PASS: TestCreateFile_Invalid (0.00s)
=== RUN   TestCreateFile_DefaultRepeat
    unit_test.go:78: 
        --- [TestCreateFile_DefaultRepeat] Repeat<=0 debe usar default=1 ---
--- PASS: TestCreateFile_DefaultRepeat (0.00s)
=== RUN   TestCreateFile_CancelledBeforeStart
    unit_test.go:78: 
        --- [TestCreateFile_CancelledBeforeStart] Cancelado antes de iniciar no debe crear archivo ---
--- PASS: TestCreateFile_CancelledBeforeStart (0.00s)
=== RUN   TestCreateFile_WriteError
    unit_test.go:78: 
        --- [TestCreateFile_WriteError] Error de escritura (ruta inválida) debe devolver 500 ---
--- PASS: TestCreateFile_WriteError (0.00s)
=== RUN   TestDeleteFile_Success
    unit_test.go:78: 
        --- [TestDeleteFile_Success] Borrado de archivo existente ---
--- PASS: TestDeleteFile_Success (0.00s)
=== RUN   TestDeleteFile_NotFound
    unit_test.go:78: 
        --- [TestDeleteFile_NotFound] Borrado de archivo inexistente debe devolver 500 ---
--- PASS: TestDeleteFile_NotFound (0.00s)
=== RUN   TestDeleteFile_MissingName
    unit_test.go:78: 
        --- [TestDeleteFile_MissingName] Falta de nombre debe devolver 400 ---
--- PASS: TestDeleteFile_MissingName (0.00s)
=== RUN   TestDeleteFile_CancelledBeforeStart
    unit_test.go:78: 
        --- [TestDeleteFile_CancelledBeforeStart] Cancelado antes de iniciar no debe eliminar archivo ---
--- PASS: TestDeleteFile_CancelledBeforeStart (0.00s)
=== RUN   TestSortFile_Success
    unit_test.go:78: 
        --- [TestSortFile_Success] Ordenamiento de números por merge sort ---
--- PASS: TestSortFile_Success (0.00s)
=== RUN   TestSortFile_InvalidAlgo
    unit_test.go:78: 
        --- [TestSortFile_InvalidAlgo] Algoritmo inválido debe devolver 400 ---
--- PASS: TestSortFile_InvalidAlgo (0.00s)
=== RUN   TestSortFile_MissingName
    unit_test.go:78: 
        --- [TestSortFile_MissingName] Falta de nombre de archivo debe devolver 400 ---
--- PASS: TestSortFile_MissingName (0.00s)
=== RUN   TestSortFile_FileNotFound
    unit_test.go:78: 
        --- [TestSortFile_FileNotFound] Archivo inexistente debe devolver 500 ---
--- PASS: TestSortFile_FileNotFound (0.00s)
=== RUN   TestSortFile_EmptyFile
    unit_test.go:78: 
        --- [TestSortFile_EmptyFile] Archivo vacío debe devolver 400 ---
--- PASS: TestSortFile_EmptyFile (0.00s)
=== RUN   TestSortFile_CancelledBeforeStart
    unit_test.go:78: 
        --- [TestSortFile_CancelledBeforeStart] Cancelación antes de iniciar debe devolver 499 ---
--- PASS: TestSortFile_CancelledBeforeStart (0.00s)
=== RUN   TestSortFile_CancelDuringRead
    unit_test.go:78: 
        --- [TestSortFile_CancelDuringRead] Cancelación durante lectura (sin llamada, se elimina el archivo) ---
--- PASS: TestSortFile_CancelDuringRead (0.00s)
=== RUN   TestSortFile_CancelDuringWrite
    unit_test.go:78: 
        --- [TestSortFile_CancelDuringWrite] Cancelación durante escritura del .sorted ---
--- PASS: TestSortFile_CancelDuringWrite (0.00s)
=== RUN   TestSortFile_DefaultQuickSort
    unit_test.go:78: 
        --- [TestSortFile_DefaultQuickSort] Algoritmo por defecto quicksort cuando 'algo' está vacío ---
--- PASS: TestSortFile_DefaultQuickSort (0.00s)
=== RUN   TestWordCount_Success
    unit_test.go:78: 
        --- [TestWordCount_Success] Recuento de un archivo simple ---
--- PASS: TestWordCount_Success (0.00s)
=== RUN   TestWordCount_MissingFile
    unit_test.go:78: 
        --- [TestWordCount_MissingFile] Archivo inexistente debe devolver 500 ---
--- PASS: TestWordCount_MissingFile (0.00s)
=== RUN   TestGrep_Success
    unit_test.go:78: 
        --- [TestGrep_Success] Búsqueda de patrón con coincidencias ---
--- PASS: TestGrep_Success (0.00s)
=== RUN   TestGrep_InvalidRegex
    unit_test.go:78: 
        --- [TestGrep_InvalidRegex] Regex inválido debe devolver 400 ---
--- PASS: TestGrep_InvalidRegex (0.00s)
=== RUN   TestGrep_MissingParams
    unit_test.go:78: 
        --- [TestGrep_MissingParams] Falta de parámetros debe devolver 400 ---
--- PASS: TestGrep_MissingParams (0.00s)
=== RUN   TestGrep_FileNotFound
    unit_test.go:78: 
        --- [TestGrep_FileNotFound] Archivo inexistente debe devolver 500 ---
--- PASS: TestGrep_FileNotFound (0.00s)
=== RUN   TestGrep_CancelledDuringRead
    unit_test.go:78: 
        --- [TestGrep_CancelledDuringRead] Cancelación durante lectura (sin llamada, conserva intención) ---
--- PASS: TestGrep_CancelledDuringRead (0.00s)
=== RUN   TestGrep_NoMatches
    unit_test.go:78: 
        --- [TestGrep_NoMatches] Búsqueda sin coincidencias debe retornar matches=0 ---
--- PASS: TestGrep_NoMatches (0.00s)
=== RUN   TestGrep_ScannerError
    unit_test.go:78: 
        --- [TestGrep_ScannerError] Pasar directorio debe provocar error (500) ---
--- PASS: TestGrep_ScannerError (0.00s)
=== RUN   TestHashFile_Success
    unit_test.go:78: 
        --- [TestHashFile_Success] Hash de archivo válido ---
--- PASS: TestHashFile_Success (0.00s)
=== RUN   TestHashFile_InvalidAlgo
    unit_test.go:78: 
        --- [TestHashFile_InvalidAlgo] Algoritmo inválido debe devolver 400 ---
--- PASS: TestHashFile_InvalidAlgo (0.00s)
=== RUN   TestCompressFile_GzipSuccess
    unit_test.go:78: 
        --- [TestCompressFile_GzipSuccess] Compresión gzip con archivo válido ---
--- PASS: TestCompressFile_GzipSuccess (0.00s)
=== RUN   TestCompressFile_InvalidCodec
    unit_test.go:78: 
        --- [TestCompressFile_InvalidCodec] Codec inválido debe devolver 400 ---
--- PASS: TestCompressFile_InvalidCodec (0.00s)
=== RUN   TestCompressFile_Cancelled
    unit_test.go:78: 
        --- [TestCompressFile_Cancelled] Cancelación durante compresión gzip ---
--- PASS: TestCompressFile_Cancelled (0.00s)
=== RUN   TestCompressFile_MissingName
    unit_test.go:78: 
        --- [TestCompressFile_MissingName] Falta de nombre debe devolver 400 ---
--- PASS: TestCompressFile_MissingName (0.00s)
=== RUN   TestCompressFile_FileNotFound
    unit_test.go:78: 
        --- [TestCompressFile_FileNotFound] Archivo inexistente debe devolver 500 ---
--- PASS: TestCompressFile_FileNotFound (0.00s)
=== RUN   TestCompressFile_DefaultCodec
    unit_test.go:78: 
        --- [TestCompressFile_DefaultCodec] Codec por defecto gzip cuando string vacío ---
--- PASS: TestCompressFile_DefaultCodec (0.00s)
=== RUN   TestCompressFile_XZCodec
    unit_test.go:78: 
        --- [TestCompressFile_XZCodec] Compresión con xz (resultado dependiente del entorno) ---
--- PASS: TestCompressFile_XZCodec (0.01s)
=== RUN   TestCompressFile_CreateOutputError
    unit_test.go:78: 
        --- [TestCompressFile_CreateOutputError] Error al crear archivo de salida debe devolver 500 ---
--- PASS: TestCompressFile_CreateOutputError (0.00s)
=== RUN   TestCompressFile_EmptyFile
    unit_test.go:78: 
        --- [TestCompressFile_EmptyFile] Compresión de archivo vacío (200) ---
--- PASS: TestCompressFile_EmptyFile (0.00s)
=== RUN   TestCompressFile_FileMissingAndXZ
    unit_test.go:78: 
        --- [TestCompressFile_FileMissingAndXZ] Missing file (500) y luego xz (200) ---
--- PASS: TestCompressFile_FileMissingAndXZ (0.00s)
=== RUN   TestSortFile_QuickSort
    unit_test.go:78: 
        --- [TestSortFile_QuickSort] Ordenamiento quicksort explícito ---
--- PASS: TestSortFile_QuickSort (0.00s)
=== RUN   TestIsPrime_EdgeCases
    unit_test.go:78: 
        --- [TestIsPrime_EdgeCases] Casos límite: 0,1,2,4 ---
--- PASS: TestIsPrime_EdgeCases (0.00s)
=== RUN   TestLoadTest_Cancelled
    unit_test.go:78: 
        --- [TestLoadTest_Cancelled] Cancelación durante LoadTest ---
--- PASS: TestLoadTest_Cancelled (1.00s)
=== RUN   TestOverall_Consistency
    unit_test.go:78: 
        --- [TestOverall_Consistency] Resumen de verificación general ---
    unit_test.go:1830: Todos los algoritmos básicos, CPU-bound e IO-bound verificados con éxito
--- PASS: TestOverall_Consistency (0.00s)
=== RUN   TestLoadTest_InvalidAndCancel
    unit_test.go:78: 
        --- [TestLoadTest_InvalidAndCancel] Params inválidos y cancelación en LoadTest ---
--- PASS: TestLoadTest_InvalidAndCancel (1.00s)
=== RUN   TestHashFile_FileMissingAndCancel
    unit_test.go:78: 
        --- [TestHashFile_FileMissingAndCancel] Missing file (500) y cancelación inmediata ---
--- PASS: TestHashFile_FileMissingAndCancel (0.00s)
=== RUN   TestSimulateWork_EdgeCases
    unit_test.go:78: 
        --- [TestSimulateWork_EdgeCases] Segundos=0 (400) y cancelación inmediata ---
--- PASS: TestSimulateWork_EdgeCases (0.00s)
=== RUN   TestSortFile_EmptyAndBadNumbers
    unit_test.go:78: 
        --- [TestSortFile_EmptyAndBadNumbers] Archivo vacío (400) y líneas no numéricas (400) ---
--- PASS: TestSortFile_EmptyAndBadNumbers (0.00s)
=== RUN   TestLoadTest_EdgeAndCancel
    unit_test.go:78: 
        --- [TestLoadTest_EdgeAndCancel] Tasks=0 (400), sleep<0 (400) y cancelación ---
--- PASS: TestLoadTest_EdgeAndCancel (1.00s)
=== RUN   TestGrep_NoMatchesAndCancel
    unit_test.go:78: 
        --- [TestGrep_NoMatchesAndCancel] Sin coincidencias y cancelación inmediata ---
--- PASS: TestGrep_NoMatchesAndCancel (0.00s)
=== RUN   TestRandom_EdgeCases
    unit_test.go:78: 
        --- [TestRandom_EdgeCases] count=0 (400), min=max (400) y cancelación ---
--- PASS: TestRandom_EdgeCases (0.00s)
PASS
coverage: 90.4% of statements in github.com/EngSteven/pso-http-server/internal/algorithms
ok      github.com/EngSteven/pso-http-server/tests      94.525s coverage: 90.4% of statements in github.com/EngSteven/pso-http-server/internal/algorithms
```