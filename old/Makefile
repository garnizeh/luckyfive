# --- Configurações Gerais ---
GO=go
BIN_MAIN=tools/loader/loader.go
BIN_COMPARE=tools/simulations/compare_simulations.go
BIN_CONCURRENCY=tools/sweep/concurrent_sweep.go
BIN_THINKER=tools/thinker/thinker.go
# Binário pré-compilado para rodar o sweep mais rápido
# On Windows the executable will have a .exe suffix; set EXE accordingly
EXE :=
ifeq ($(OS),Windows_NT)
EXE := .exe
endif
MAIN_BIN=./main_bin$(EXE)
THINKER_BIN=./thinker_bin$(EXE)
DATA_DIR=data
SIM_OUT=$(DATA_DIR)/simulations
DB=$(DATA_DIR)/dados.db

# --- Parâmetros da Simulação (Backtesting) ---
SIM_START=6877
SIM_END=6881
SIM_PREV_MAX=200
SIM_PREDS=25
# Quantas vezes repetir a mesma simulação para validar consistência
REPEAT=5 

# --- Parâmetros do Modelo (Ajuste Fino) ---
LAMBDA=0.08
HOT_COLD_BOOST=1.5
ALPHA=1.5
BETA=0.5
GAMMA=1.0
CLUSTER_PENALTY=2.0
CANDS_MULT=100
HILL_ITER=100
COOCC_WINDOW=100
HOT_WINDOW=15
SMART_FILTERS=true
SEED=0

# --- Parâmetros de Geração (Produção) ---
GEN_GAMES=10

# Constrói a string de argumentos para a simulação
SIM_ARGS=--db $(DB) --load=false \
	--sim-start=$(SIM_START) --sim-end=$(SIM_END) \
	--sim-prev-max=$(SIM_PREV_MAX) --sim-preds=$(SIM_PREDS) \
	--lambda=$(LAMBDA) --hot-cold-boost=$(HOT_COLD_BOOST) \
	--alpha=$(ALPHA) --beta=$(BETA) --gamma=$(GAMMA) \
	--cluster-penalty=$(CLUSTER_PENALTY) \
	--cands-mult=$(CANDS_MULT) --hill-iter=$(HILL_ITER) \
	--coocc-window=$(COOCC_WINDOW) --hot-window=$(HOT_WINDOW) \
	--smart-filters=$(SMART_FILTERS) \
	--sim-save=true --sim-out=$(SIM_OUT) --seed=$(SEED)

# --- Parâmetros de Sweep (Otimização de Portfólio) ---
PORTFOLIO_RUNS=100
MIN_PREDS=10   # Mínimo de jogos a testar (por configuração)
MAX_PREDS=25   # Máximo de jogos a testar (por configuração)
PREDS_STEP=5   # Incremento no número de jogos (10, 15, 20, 25)

# Range de Parâmetros para Randomização (Mais fino para o teste)
ALPHA_MIN=0.8
ALPHA_MAX=2.5
BETA_MIN=0.1
BETA_MAX=1.0
LAMBDA_MIN=0.05
LAMBDA_MAX=0.15

# CMD base para o portfolio sweep (sem SIM_PREDS, Alpha, Beta, Lambda, SmartFilters)
PORTFOLIO_BASE_CMD=--db $(DB) --load=false \
	--sim-start=$(SIM_START) --sim-end=$(SIM_END) \
	--sim-prev-max=$(SIM_PREV_MAX) \
	--hot-cold-boost=$(HOT_COLD_BOOST) --gamma=$(GAMMA) \
	--cluster-penalty=$(CLUSTER_PENALTY) \
	--cands-mult=$(CANDS_MULT) --hill-iter=$(HILL_ITER) \
	--coocc-window=$(COOCC_WINDOW) --hot-window=$(HOT_WINDOW) \
	--sim-save=true --sim-out=$(SIM_OUT) --sim-clean=false

# CMD base para o sweep (exclui Alpha, Beta, Lambda, SmartFilters que serão randomizados)
MEGA_SWEEP_BASE_CMD=--db $(DB) --load=false \
	--sim-start=$(SIM_START) --sim-end=$(SIM_END) \
	--sim-prev-max=$(SIM_PREV_MAX) --sim-preds=$(SIM_PREDS) \
	--hot-cold-boost=$(HOT_COLD_BOOST) --gamma=$(GAMMA) \
	--cluster-penalty=$(CLUSTER_PENALTY) \
	--cands-mult=$(CANDS_MULT) --hill-iter=$(HILL_ITER) \
	--coocc-window=$(COOCC_WINDOW) --hot-window=$(HOT_WINDOW) \
	--sim-save=true --sim-out=$(SIM_OUT) --sim-clean=false

# --- Concorrência ---
# Defina o número exato de threads
NUM_THREADS=27
# Se você quiser que o padrão seja NumCPU - 1, deixe a linha acima com apenas NUM_THREADS?=

# --- Regras de Build e Setup ---

.PHONY: build setup clean

# Regra Build: Compila o main.go para um binário executável
build: $(MAIN_BIN)

$(MAIN_BIN): $(BIN_MAIN)
	@echo "🛠️ Compilando $(BIN_MAIN) para binário..."
	$(GO) build -o $(MAIN_BIN) $(BIN_MAIN)

# Compila o utilitário thinker (tools/thinker/thinker.go) para um binário
thinker: $(THINKER_BIN)

$(THINKER_BIN): $(BIN_THINKER)
	@echo "🛠️ Compilando $(BIN_THINKER) para binário..."
	$(GO) build -o $(THINKER_BIN) $(BIN_THINKER)

# --- Targets ---

.PHONY: help setup init-db sim run clean compare sweep sweep-alpha thinker

help: ## Mostra esta ajuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Cria diretórios necessários
	@mkdir -p $(DATA_DIR)
	@mkdir -p $(SIM_OUT)

# Inicializa o banco `data/dados.db` a partir do Excel `data/Quina.xlsx`
init-db: setup
	@echo "📥 Inicializando banco $(DB) a partir de data/Quina.xlsx..."
	# Pass --gen-games=0 to avoid the default generation output during init
	$(GO) run $(BIN_MAIN) --db $(DB) --load=true --gen-games=0

sim: setup ## Roda uma única simulação com os parâmetros atuais
	@echo "🚀 Iniciando Simulação..."
	$(GO) run $(BIN_MAIN) $(SIM_ARGS) --sim-clean=false
	@$(MAKE) compare

run: setup ## Gera jogos REAIS para o próximo concurso (Modo Produção)
	@echo "🎲 Gerando $(GEN_GAMES) jogos otimizados para aposta..."
	$(GO) run $(BIN_MAIN) --db $(DB) --load=false --gen-games=$(GEN_GAMES) \
		--lambda=$(LAMBDA) --alpha=$(ALPHA) --beta=$(BETA) --smart-filters=$(SMART_FILTERS)

compare: ## Analisa os resultados das simulações (ordena por Quina)
	@echo "📊 Comparando resultados..."
	$(GO) run $(BIN_COMPARE) -dir $(SIM_OUT) -n 20 -sort quina

clean:
	@echo "🧹 Limpando data/simulations..."
	rm -rf $(SIM_OUT)/*
	@echo "🧹 Limpando binário..."
	# Remove o binário para forçar a recompilação se necessário
	rm -f $(MAIN_BIN)	

# --- Testes de Estratégia (Grid Search) ---

sweep: sweep-alpha compare ## Roda teste comparativo variando Alpha (1.0, 1.5, 2.0)

sweep-alpha: setup
	@echo "🧪 Testando Alpha = 1.0 (Conservador)..."
	$(GO) run $(BIN_MAIN) $(SIM_ARGS) --alpha=1.0 --sim-clean=false
	@echo "🧪 Testando Alpha = 1.5 (Balanceado)..."
	$(GO) run $(BIN_MAIN) $(SIM_ARGS) --alpha=1.5 --sim-clean=false
	@echo "🧪 Testando Alpha = 2.5 (Focado em Pares)..."
	$(GO) run $(BIN_MAIN) $(SIM_ARGS) --alpha=2.5 --sim-clean=false

# Roda simulação rápida para debug
quick: 
	$(GO) run $(BIN_MAIN) $(SIM_ARGS) --cands-mult=20 --hill-iter=10 --sim-clean=false

compare-top5: ## Analisa os TOP 5 resultados das simulações (ordena por Quina)
	@echo "📊 Comparando TOP 5 resultados..."
	$(GO) run $(BIN_COMPARE) -dir $(SIM_OUT) -n 5 -sort quina

test-preds: setup build ## Testa taxa de obter ao menos 1 terno variando jogos entre 10 e 25 (concorrente)
	@echo "🔬 Teste preds=10..25 (concorrente, threads: 27, runs: $(PORTFOLIO_RUNS))"
	@$(GO) run $(BIN_CONCURRENCY) \
		-main-bin=$(MAIN_BIN) \
		-threads=27 \
		-runs=$(PORTFOLIO_RUNS) \
		-min-preds=10 -max-preds=25 -preds-step=5 \
		-alpha-min=$(ALPHA_MIN) -alpha-max=$(ALPHA_MAX) \
		-beta-min=$(BETA_MIN) -beta-max=$(BETA_MAX) \
		-lambda-min=$(LAMBDA_MIN) -lambda-max=$(LAMBDA_MAX) \
		-base-cmd='$(PORTFOLIO_BASE_CMD)'

# Variáveis configuráveis (pode sobrescrever via env/make):
# Ex: make gen-contest TARGET_CONTEST=7000 TARGET_GAMES=30
TARGET_CONTEST ?= 6882
TARGET_GAMES ?= 25
TARGET_LIMIT ?= $(shell expr $(TARGET_CONTEST) - 1)

# Parâmetros do gerador (padrões herdados do topo do Makefile)
GEN_LAMBDA ?= $(LAMBDA)
GEN_HOT_COLD_BOOST ?= $(HOT_COLD_BOOST)
GEN_ALPHA ?= $(ALPHA)
GEN_BETA ?= $(BETA)
GEN_GAMMA ?= $(GAMMA)
GEN_CLUSTER_PENALTY ?= $(CLUSTER_PENALTY)
GEN_CANDS_MULT ?= $(CANDS_MULT)
GEN_HILL_ITER ?= $(HILL_ITER)
GEN_COOCC_WINDOW ?= $(COOCC_WINDOW)
GEN_HOT_WINDOW ?= $(HOT_WINDOW)
GEN_SMART_FILTERS ?= $(SMART_FILTERS)
GEN_SEED ?= $(SEED)

# Arquivos de saída (podem ser sobrescritos)
GEN_OUT ?= $(DATA_DIR)/generated_$(TARGET_CONTEST).txt
COMPARE_OUT ?= $(DATA_DIR)/compare_$(TARGET_CONTEST).txt

.PHONY: gen-contest
gen-contest: init-db ## Gera $(TARGET_GAMES) cartões para o concurso $(TARGET_CONTEST) usando histórico até $(TARGET_LIMIT)
	@echo "🎯 Gerando $(TARGET_GAMES) cartoes para o concurso $(TARGET_CONTEST) (histórico até $(TARGET_LIMIT))"
	$(GO) run $(BIN_MAIN) --db $(DB) --load=false \
		--gen-games=$(TARGET_GAMES) --limit-contest=$(TARGET_LIMIT) \
		--lambda=$(GEN_LAMBDA) --hot-cold-boost=$(GEN_HOT_COLD_BOOST) \
		--alpha=$(GEN_ALPHA) --beta=$(GEN_BETA) --gamma=$(GEN_GAMMA) \
		--cluster-penalty=$(GEN_CLUSTER_PENALTY) \
		--cands-mult=$(GEN_CANDS_MULT) --hill-iter=$(GEN_HILL_ITER) \
		--coocc-window=$(GEN_COOCC_WINDOW) --hot-window=$(GEN_HOT_WINDOW) \
		--smart-filters=$(GEN_SMART_FILTERS) --seed=$(GEN_SEED) | tee $(GEN_OUT)

.PHONY: compare-contest
compare-contest: gen-contest
	@echo "🔎 Comparando os $(TARGET_GAMES) cartoes com o resultado do concurso $(TARGET_CONTEST)"
	$(GO) run tools/cards/compare_cards.go --db $(DB) --cards-file $(GEN_OUT) --contest=$(TARGET_CONTEST) | tee $(COMPARE_OUT)

# Certifique-se de ter esta variável definida no topo do Makefile:
# NUM_THREADS?= (vazio para default ou um número como 27)
# MAIN_BIN=./main_bin
# BIN_CONCURRENCY=tools/concurrent_sweep.go

mega-sweep: setup clean build ## Roda $(SWEEP_RUNS) simulações aleatórias em paralelo e mostra o Top 5
	@echo "🚀 Iniciando Mega-Sweep Concorrente: $(SWEEP_RUNS) configurações (Threads: $(NUM_THREADS))"
	$(GO) run $(BIN_CONCURRENCY) \
		-main-bin=$(MAIN_BIN) \
		-threads=$(NUM_THREADS) \
		-runs=$(SWEEP_RUNS) \
		-alpha-min=$(ALPHA_MIN) -alpha-max=$(ALPHA_MAX) \
		-beta-min=$(BETA_MIN) -beta-max=$(BETA_MAX) \
		-lambda-min=$(LAMBDA_MIN) -lambda-max=$(LAMBDA_MAX) \
		-base-cmd='$(BASE_SWEEP_CMD)'
	@echo "Simulações concluídas. Analisando o TOP 5..."
	@$(MAKE) compare-top5

portfolio-sweep: setup clean build ## Otimiza a quantidade de jogos (10-25) com 100 configs para maximizar Terno
	@echo "🧪 Iniciando Portfolio-Sweep Concorrente (Otimização de Portfólio) (Threads: $(NUM_THREADS))"
	$(GO) run $(BIN_CONCURRENCY) \
		-main-bin=$(MAIN_BIN) \
		-threads=$(NUM_THREADS) \
		-runs=$(PORTFOLIO_RUNS) \
		-min-preds=$(MIN_PREDS) -max-preds=$(MAX_PREDS) -preds-step=$(PREDS_STEP) \
		-alpha-min=$(ALPHA_MIN) -alpha-max=$(ALPHA_MAX) \
		-beta-min=$(BETA_MIN) -beta-max=$(BETA_MAX) \
		-lambda-min=$(LAMBDA_MIN) -lambda-max=$(LAMBDA_MAX) \
		-base-cmd='$(BASE_SWEEP_CMD)'
	@echo "Simulações concluídas. Analisando o TOP 5 por Terno..."
	@$(MAKE) compare-portfolio

compare-portfolio: ## Analisa os TOP 5 resultados por Terno (Exato)
	@echo "📊 Comparando TOP 5 resultados ordenados por TERNO..."
	$(GO) run $(BIN_COMPARE) -dir $(SIM_OUT) -n 5 -sort terno