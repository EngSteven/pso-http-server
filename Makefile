# Makefile para PSO HTTP Server (Go)
APP_NAME=pso-http-server
CMD_PATH=./cmd/server
BUILD_DIR=./bin

.PHONY: all build run test-all lint clean

all: build

# ------------------------------------------------------------
# ⚙️ Compilar el proyecto
# ------------------------------------------------------------
build:
	@echo "Compilando $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)
	@echo "✅ Build completado: $(BUILD_DIR)/$(APP_NAME)"

# ------------------------------------------------------------
# 🚀 Ejecutar el binario compilado
# ------------------------------------------------------------
run:
	@echo "Ejecutando servidor..."
	@$(BUILD_DIR)/$(APP_NAME)

# ------------------------------------------------------------
# 🧪 Ejecutar TODOS los tests con cobertura
# ------------------------------------------------------------
test-all:
	@echo "===================================================="
	@echo "🧪 Running unit tests (algorithms only)..."
	@echo "===================================================="
	go test ./tests -v -count=1 \
	    -coverpkg=github.com/EngSteven/pso-http-server/internal/algorithms \
	    -coverprofile=reports/coverage.out

# ------------------------------------------------------------
# 🔍 Linter (opcional)
# ------------------------------------------------------------
lint:
	@echo "Ejecutando linter..."
	@golangci-lint run || echo "⚠️ Linter detectó advertencias"

# ------------------------------------------------------------
# 🧹 Limpiar binarios y reportes
# ------------------------------------------------------------
clean:
	@echo "🧹 Limpiando archivos compilados y reportes..."
	@rm -rf $(BUILD_DIR) reports test_report.txt coverage.out
