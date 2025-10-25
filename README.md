# PSO HTTP Server (Proyecto 1 - Sistemas Operativos)

Servidor HTTP/1.0 concurrente escrito en Go para el curso de Principios de Sistemas Operativos.

## Autores

Steven Sequeira Araya 

Jefferson Salas Cordero

## Documentación de la API:

https://documenter.getpostman.com/view/37666062/2sB3QQK8Eh

## Compilación y Ejecución

Importante, para compilar y ejecutar el proyecto debe hacerlo de una terminal que soporte Linux

1. Desde el root del proyecto ejecute para compilar el servidor:

```bash 
make build
```

2. Desde el root del proyecto ejecute para correr el servidor:

```bash 
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
