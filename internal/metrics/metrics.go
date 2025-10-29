/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: metrics.go
Descripcion: Sistema de metricas para pools de workers que registra latencias,
contadores de procesamiento y calcula percentiles para monitoreo.
*/

package metrics

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

var totalConnections int64

// PoolMetrics almacena estadisticas de rendimiento por pool de workers.
// Mantiene muestras de latencia en buffer circular para calculos estadisticos.
type PoolMetrics struct {
	mu             sync.Mutex
	TotalProcessed int64
	TotalLatencyMs int64
	Samples        []int64 // ring buffer-like (append up to cap)
	maxSamples     int
}

// NewPoolMetrics crea nueva instancia de metricas con limite de muestras.
// Entrada: maxSamples (int) - numero maximo de muestras a mantener
// Salida: *PoolMetrics - instancia inicializada de metricas
// Descripcion: Constructor que crea PoolMetrics con buffer de muestras limitado
//
//	para calculos estadisticos. Inicializa slice con capacidad
//	especificada para evitar realocaciones frecuentes.
func NewPoolMetrics(maxSamples int) *PoolMetrics {
	// === INICIALIZACION DE BUFFER CON CAPACIDAD FIJA ===
	// Crear slice con capacidad definida para evitar realocaciones
	return &PoolMetrics{Samples: make([]int64, 0, maxSamples), maxSamples: maxSamples}
}

// Record registra nueva latencia en las metricas del pool.
// Entrada: latency (time.Duration) - duracion de la operacion completada
// Salida: ninguna
// Descripcion: Actualiza contadores totales y agrega muestra al buffer circular.
//
//	Mantiene numero limitado de muestras reemplazando las mas antiguas.
//	Thread-safe usando mutex para acceso concurrente.
func (m *PoolMetrics) Record(latency time.Duration) {
	// === CONVERSION A MILISEGUNDOS ===
	// Convertir duration a entero para almacenamiento eficiente
	ms := latency.Milliseconds()
	m.mu.Lock()
	defer m.mu.Unlock()
	// === ACTUALIZACION DE CONTADORES ACUMULATIVOS ===
	// Incrementar totales para cálculo de promedio global
	m.TotalProcessed++
	m.TotalLatencyMs += ms
	// === GESTION DEL BUFFER CIRCULAR ===
	// Mantener número limitado de muestras para percentiles
	if len(m.Samples) < m.maxSamples {
		// === FASE DE CRECIMIENTO - AGREGAR MUESTRA ===
		m.Samples = append(m.Samples, ms)
	} else {
		// === FASE ESTABLE - REEMPLAZAR MUESTRA MAS ANTIGUA ===
		// Descartar primera muestra y agregar nueva al final
		m.Samples = append(m.Samples[1:], ms)
	}
}

// AvgLatencyMs calcula latencia promedio basada en totales acumulados.
// Entrada: ninguna (metodo de PoolMetrics)
// Salida: float64 - latencia promedio en milisegundos
// Descripcion: Calcula promedio dividiendo latencia total acumulada entre
//
//	numero total de operaciones procesadas. Retorna 0 si no
//	hay operaciones registradas. Thread-safe con mutex.
func (m *PoolMetrics) AvgLatencyMs() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	// === VALIDACION DE DIVISION POR CERO ===
	// Prevenir error si no hay operaciones registradas
	if m.TotalProcessed == 0 {
		return 0
	}
	// === CALCULO DE PROMEDIO ACUMULATIVO ===
	// Usar totales acumulados para promedio eficiente
	return float64(m.TotalLatencyMs) / float64(m.TotalProcessed)
}

// Percentile calcula percentil especificado de las muestras de latencia.
// Entrada: p (float64) - percentil a calcular (0.0-100.0)
// Salida: int64 - valor del percentil en milisegundos
// Descripcion: Ordena copia de muestras y retorna valor en posicion calculada
//
//	del percentil especificado. Maneja casos edge con indices
//	fuera de rango. Thread-safe copiando slice antes de ordenar.
func (m *PoolMetrics) Percentile(p float64) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	// === VALIDACION DE MUESTRAS DISPONIBLES ===
	// Retornar 0 si no hay datos para calcular percentil
	if len(m.Samples) == 0 {
		return 0
	}
	// === COPIA DEFENSIVA PARA ORDENAMIENTO ===
	// Crear copia para no modificar slice original
	cp := make([]int64, len(m.Samples))
	copy(cp, m.Samples)
	// === ORDENAMIENTO ASCENDENTE ===
	// Ordenar copia para cálculo de percentil
	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })
	// === CALCULO DE INDICE DEL PERCENTIL ===
	// Convertir porcentaje a índice con redondeo
	idx := int((p/100.0)*float64(len(cp)-1) + 0.5)
	// === CLAMPEO DE INDICES ===
	// Asegurar que índice esté dentro de límites válidos
	if idx < 0 {
		idx = 0
	}
	if idx >= len(cp) {
		idx = len(cp) - 1
	}
	// === RETORNO DEL VALOR EN POSICION CALCULADA ===
	return cp[idx]
}

// IncrementConnections incrementa contador global de conexiones de forma atomica.
// Entrada: ninguna
// Salida: ninguna
// Descripcion: Incrementa atomicamente el contador global de conexiones TCP
//
//	procesadas por el servidor. Thread-safe usando operaciones
//	atomicas para evitar race conditions en entorno concurrente.
func IncrementConnections() {
	// === INCREMENTO ATOMICO THREAD-SAFE ===
	// Usar operaciones atómicas para concurrencia sin mutex
	atomic.AddInt64(&totalConnections, 1)
}

// GetTotalConnections retorna total de conexiones procesadas atomicamente.
// Entrada: ninguna
// Salida: int64 - numero total de conexiones procesadas
// Descripcion: Lee atomicamente el valor actual del contador global de
//
//	conexiones. Thread-safe usando operaciones atomicas para
//	acceso consistente desde multiples goroutines.
func GetTotalConnections() int64 {
	// === LECTURA ATOMICA CONSISTENTE ===
	// Leer valor actual sin posibilidad de race conditions
	return atomic.LoadInt64(&totalConnections)
}
