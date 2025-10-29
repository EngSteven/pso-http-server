/*
Autores: Steven Sequeira Araya, Jefferson Salas Cordero
Nombre del archivo: router.go
Descripcion: Router HTTP simple que mapea rutas a handlers usando
mapa de strings para busqueda directa y rapida de endpoints.
*/

package router

import "github.com/EngSteven/pso-http-server/internal/types"

// Router mantiene mapa de rutas a funciones handler.
type Router struct {
	routes map[string]types.HandlerFunc
}

// NewRouter crea nueva instancia de router con mapa vacio.
// Entrada: ninguna
// Salida: *Router - router inicializado
// Descripcion: Constructor que crea router con mapa de rutas vacio.
//              Inicializa estructura para mapeo directo de paths a handlers
//              permitiendo registro y busqueda rapida de endpoints.
func NewRouter() *Router {
	return &Router{routes: make(map[string]types.HandlerFunc)}
}

// Handle registra handler para ruta especifica en el router.
// Entrada: path (string) - ruta del endpoint a registrar
//          handler (types.HandlerFunc) - funcion handler para la ruta
// Salida: ninguna
// Descripcion: Registra mapeo de ruta a funcion handler en mapa interno.
//              Permite asociar paths especificos con funciones de procesamiento
//              para resolucion posterior durante requests HTTP.
func (r *Router) Handle(path string, handler types.HandlerFunc) {
	r.routes[path] = handler
}

// Match busca handler para ruta dada, retorna nil si no existe.
// Entrada: path (string) - ruta a buscar en el router
// Salida: types.HandlerFunc - handler asociado a la ruta o nil
// Descripcion: Busca handler registrado para ruta especifica en mapa interno.
//              Retorna funcion handler si existe o nil si la ruta no esta
//              registrada. Permite resolucion rapida de endpoints durante requests.
func (r *Router) Match(path string) types.HandlerFunc {
	if h, ok := r.routes[path]; ok {
		return h
	}
	return nil
}
