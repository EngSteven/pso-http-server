# Makefile para PSO HTTP Server (Go)
APP_NAME=pso-http-server
CMD_PATH=./cmd/server
BUILD_DIR=./bin

.PHONY: all build run test clean

all: build

# Compilar el proyecto
build:
	@echo "Compilando $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)
	@echo "Build completado: $(BUILD_DIR)/$(APP_NAME)"

# Ejecutar el binario
run:
	@echo "Ejecutando servidor..."
	@$(BUILD_DIR)/$(APP_NAME)

# Corrrer todos los tests
test:
	@echo "Ejecutando pruebas unitarias..."
	@go test ./... -v

lint:
	@echo "Ejecutando linter..."
	@golangci-lint run || echo "Linter detectó advertencias"

# Borrar los binarios
clean:
	@echo "Limpiando archivos compilados..."
	@rm -rf $(BUILD_DIR)
