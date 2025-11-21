package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

// ensureDependencies instala o pacote necessário, se ausente.
func ensureDependencies() {
	fmt.Println("Verificando dependências...")
	cmd := exec.Command("go", "get", "golang.org/x/sync/errgroup")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao instalar dependência 'golang.org/x/sync/errgroup': %v\n", err)
		os.Exit(1)
	}
}

func main() {
	ensureDependencies()

	// Parâmetros de Concorrência
	defaultThreads := runtime.NumCPU() - 1
	if defaultThreads < 1 {
		defaultThreads = 1
	}
	concurrency := flag.Int("threads", defaultThreads, "Limite de goroutines (simulações paralelas). Padrão: num_cpus - 1")

	// Parâmetros de Sweep
	runs := flag.Int("runs", 100, "Número total de configurações a testar")
	minPreds := flag.Int("min-preds", 0, "Mínimo de jogos a testar (0 para ignorar a iteração de preds)")
	maxPreds := flag.Int("max-preds", 0, "Máximo de jogos a testar (0 para ignorar a iteração de preds)")
	stepPreds := flag.Int("preds-step", 1, "Incremento no número de jogos (SIM_PREDS)")

	// Ranges de Parâmetros
	alphaMin := flag.Float64("alpha-min", 0.5, "Mínimo para Alpha (A)")
	alphaMax := flag.Float64("alpha-max", 3.0, "Máximo para Alpha (A)")
	betaMin := flag.Float64("beta-min", 0.1, "Mínimo para Beta (B)")
	betaMax := flag.Float64("beta-max", 1.5, "Máximo para Beta (B)")
	lambdaMin := flag.Float64("lambda-min", 0.03, "Mínimo para Lambda (L)")
	lambdaMax := flag.Float64("lambda-max", 0.20, "Máximo para Lambda (L)")

	// Caminho para o binário principal
	mainBinPath := flag.String("main-bin", "./main", "Caminho para o binário principal (main)")

	// Base de Comandos (o resto dos parâmetros)
	baseCmdStr := flag.String("base-cmd", "", "Comando base com o resto dos flags de simulação")

	flag.Parse()

	if *concurrency == 0 {
		*concurrency = 1
	}

	rand.Seed(time.Now().UnixNano())

	fmt.Printf("\n🔥 Iniciando Sweep Paralelo: %d configurações (Threads: %d)\n", *runs, *concurrency)
	fmt.Printf("Parâmetros: A:[%.2f-%.2f] B:[%.2f-%.2f] L:[%.3f-%.3f] SimPreds:[%d-%d] (Passo: %d)\n",
		*alphaMin, *alphaMax, *betaMin, *betaMax, *lambdaMin, *lambdaMax, *minPreds, *maxPreds, *stepPreds)

	// Separamos o comando base
	baseArgs := strings.Fields(*baseCmdStr)

	// Lista de valores de SIM_PREDS para iterar
	var predsList []int
	if *minPreds == 0 {
		predsList = append(predsList, 0) // Executar uma vez, sem iteração de preds
	} else {
		for p := *minPreds; p <= *maxPreds; p += *stepPreds {
			predsList = append(predsList, p)
		}
	}

	// Calculamos quantas iterações por SIM_PREDS teremos
	runsPerPred := *runs / len(predsList)
	if runsPerPred < 1 {
		runsPerPred = 1
	}

	// Estrutura para limitar a concorrência
	var g errgroup.Group
	g.SetLimit(*concurrency)

	totalRuns := 0

	// Loop principal para gerar e executar as configurações
	for _, p := range predsList {
		for i := 1; i <= runsPerPred; i++ {
			totalRuns++

			// Variável de controle do loop para uso na goroutine
			p := p

			// 1. Gera parâmetros aleatórios
			rAlpha := *alphaMin + rand.Float64()*(*alphaMax-*alphaMin)
			rBeta := *betaMin + rand.Float64()*(*betaMax-*betaMin)
			rLambda := *lambdaMin + rand.Float64()*(*lambdaMax-*lambdaMin)
			rSmart := rand.Intn(2) == 1

			// 2. Monta o comando de simulação
			args := make([]string, 0, 30) // Pré-aloca espaço
			args = append(args, baseArgs...)

			args = append(args, fmt.Sprintf("--alpha=%.2f", rAlpha))
			args = append(args, fmt.Sprintf("--beta=%.2f", rBeta))
			args = append(args, fmt.Sprintf("--lambda=%.3f", rLambda))
			args = append(args, fmt.Sprintf("--smart-filters=%t", rSmart))

			if p > 0 {
				args = append(args, fmt.Sprintf("--sim-preds=%d", p))
			}

			// 3. Executa a simulação em uma goroutine
			g.Go(func() error {
				fmt.Printf(" [RUN %d/%d] A:%.2f B:%.2f L:%.3f P:%d Sm:%t\n", totalRuns, *runs, rAlpha, rBeta, rLambda, p, rSmart)

				cmd := exec.Command(*mainBinPath, args...)
				// Silenciar a saída padrão de cada run (para não poluir o terminal)
				cmd.Stdout = nil
				cmd.Stderr = nil

				if err := cmd.Run(); err != nil {
					// O sweep deve continuar mesmo se uma simulação falhar
					// logs o erro, mas retorna nil para o errgroup
					fmt.Fprintf(os.Stderr, "  [ERRO] Simulação %d falhou: %v\n", totalRuns, err)
					return nil
				}
				return nil
			})

			// Parar o loop se já atingimos o número total de runs (ajuste de arredondamento)
			if totalRuns >= *runs {
				break
			}
		}
		if totalRuns >= *runs {
			break
		}
	}

	// Espera todas as goroutines terminarem
	if err := g.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Um erro fatal ocorreu durante a simulação: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ Todas as simulações concluídas.")
}
