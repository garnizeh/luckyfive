package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

// Estruturas copiadas do main.go
type GameConfig struct {
	Lambda          float64 `json:"Lambda"`
	HotColdBoost    float64 `json:"HotColdBoost"`
	Alpha           float64 `json:"Alpha"`
	Beta            float64 `json:"Beta"`
	Gamma           float64 `json:"Gamma"`
	ClusterPenalty  float64 `json:"ClusterPenalty"`
	CandsMult       int     `json:"CandsMult"`
	HillIter        int     `json:"HillIter"`
	CooccWindow     int     `json:"CooccWindow"`
	HotWindow       int     `json:"HotWindow"`
	UseSmartFilters bool    `json:"UseSmartFilters"`
}

type SimulationResult struct {
	Params   GameConfig         `json:"params"`
	Start    int                `json:"start_contest"`
	End      int                `json:"end_contest"`
	Contests []any              `json:"contests"`
	Summary  map[string]float64 `json:"summary"`
}

// Estrutura para facilitar a ordenação e exibição
type ReportRow struct {
	File      string
	Timestamp time.Time
	Result    SimulationResult
}

func main() {
	dir := flag.String("dir", "data/simulations", "Diretório com arquivos JSON")
	topN := flag.Int("n", 20, "Quantos arquivos analisar")
	sortBy := flag.String("sort", "quina", "Critério de ordenação: 'avg', 'quina', 'quadra', 'terno', 'duque', 'recency'")
	flag.Parse()

	// Se o usuário não especificar -n e usar ordenação de prêmio, ajustamos para N=5.
	if (*sortBy == "quina" || *sortBy == "quadra" || *sortBy == "terno" || *sortBy == "duque") && *topN == 20 {
		*topN = 5
	}

	files, err := os.ReadDir(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler diretório: %v\n", err)
		os.Exit(1)
	}

	var reports []ReportRow

	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".json" && strings.HasPrefix(f.Name(), "sim_") {
			path := filepath.Join(*dir, f.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			var simRes SimulationResult
			if err := json.Unmarshal(data, &simRes); err != nil {
				continue
			}

			info, _ := f.Info()
			reports = append(reports, ReportRow{
				File:      f.Name(),
				Timestamp: info.ModTime(),
				Result:    simRes,
			})
		}
	}

	if len(reports) == 0 {
		fmt.Println("Nenhuma simulação válida encontrada em", *dir)
		return
	}

	// Ordenação inicial por data (para pegar os N mais recentes primeiro)
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Timestamp.After(reports[j].Timestamp)
	})

	// Corta para os N mais recentes
	if len(reports) > *topN {
		reports = reports[:*topN]
	}

	// Reordena baseada na flag do usuário para exibição
	sort.Slice(reports, func(i, j int) bool {
		r1 := reports[i].Result.Summary
		r2 := reports[j].Result.Summary

		switch *sortBy {
		case "avg":
			return r1["avg_best_hits"] > r2["avg_best_hits"]
		case "quina":
			return r1["cnt_5_quina"] > r2["cnt_5_quina"]
		case "quadra":
			return r1["cnt_4_plus"] > r2["cnt_4_plus"]
		case "terno":
			return r1["cnt_3_plus"] > r2["cnt_3_plus"]
		case "duque":
			return r1["cnt_2_plus"] > r2["cnt_2_plus"]
		default: // "recency"
			return reports[i].Timestamp.After(reports[j].Timestamp)
		}
	})

	// Exibição Tabular
	fmt.Printf("\n🔍 Analisando Top %d simulações (Ordenado por: %s)\n", len(reports), strings.ToUpper(*sortBy))
	fmt.Println(strings.Repeat("-", 120))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Cabeçalho ATUALIZADO
	fmt.Fprintln(w, "ARQUIVO\t| MÉDIA\t| DUQUE (2)\t| TERNO (3)\t| QUADRA (4)\t| QUINA (5)\t| PARÂMETROS (A/B/L/Sm)")
	fmt.Fprintln(w, "-------\t| -----\t| ---------\t| ---------\t| ----------\t| ---------\t| ---------------------")

	for _, row := range reports {
		sum := row.Result.Summary
		p := row.Result.Params

		// Acertos Cumulativos
		cnt5 := int(sum["cnt_5_quina"])
		cnt4plus := int(sum["cnt_4_plus"])
		cnt3plus := int(sum["cnt_3_plus"])
		cnt2plus := int(sum["cnt_2_plus"])

		// Acertos Exatos
		cnt4Exact := cnt4plus - cnt5
		cnt3Exact := cnt3plus - cnt4plus
		cnt2Exact := cnt2plus - cnt3plus

		// Sanity Check
		if cnt4Exact < 0 {
			cnt4Exact = 0
		}
		if cnt3Exact < 0 {
			cnt3Exact = 0
		}
		if cnt2Exact < 0 {
			cnt2Exact = 0
		}

		// Formata nome do arquivo curto
		shortName := row.File
		if len(shortName) > 25 {
			shortName = shortName[:22] + "..."
		}

		// Marcador visual
		marker := ""
		if cnt5 > 0 {
			marker = "🏆"
		} else if cnt4Exact > 0 {
			marker = "🌟"
		} else if cnt3Exact > 0 {
			marker = "🥉"
		} else if sum["avg_best_hits"] >= 3.0 {
			marker = "🔥"
		}

		fmt.Fprintf(w, "%s %s\t| %.3f\t| %d\t| %d\t| %d\t| %d\t| A:%.2f B:%.2f L:%.3f Sm:%v\n",
			shortName, marker,
			sum["avg_best_hits"],
			cnt2Exact, // DUQUE EXATO
			cnt3Exact, // TERNO EXATO
			cnt4Exact, // QUADRA EXATA
			cnt5,      // QUINA
			p.Alpha, p.Beta, p.Lambda, p.UseSmartFilters,
		)
	}
	w.Flush()
	fmt.Println(strings.Repeat("-", 120))

	// --- CORREÇÃO DE SCOPE AQUI ---
	if len(reports) > 0 {
		best := reports[0]
		sum := best.Result.Summary

		// Recalculamos as contagens exatas para o resumo final
		cnt5 := int(sum["cnt_5_quina"])
		cnt4plus := int(sum["cnt_4_plus"])
		cnt3plus := int(sum["cnt_3_plus"])
		cnt2plus := int(sum["cnt_2_plus"])

		cnt4Exact := cnt4plus - cnt5
		cnt3Exact := cnt3plus - cnt4plus
		cnt2Exact := cnt2plus - cnt3plus

		if cnt4Exact < 0 {
			cnt4Exact = 0
		}
		if cnt3Exact < 0 {
			cnt3Exact = 0
		}
		if cnt2Exact < 0 {
			cnt2Exact = 0
		}

		fmt.Printf("\n💡 MELHOR ESTRATÉGIA NA LISTA:\n")
		fmt.Printf("   Arquivo: %s\n", best.File)
		fmt.Printf("   Média Acertos: %.4f\n", sum["avg_best_hits"])
		fmt.Printf("   Prêmios: Quinas: %d, Quadras (Exatas): %d, Ternos (Exatos): %d, Duques (Exatos): %d\n",
			cnt5, cnt4Exact, cnt3Exact, cnt2Exact)
		fmt.Printf("   Configuração Vencedora: Alpha=%.2f, Beta=%.2f, Gamma=%.2f, Lambda=%.3f, SmartFilters=%v\n",
			best.Result.Params.Alpha, best.Result.Params.Beta, best.Result.Params.Gamma, best.Result.Params.Lambda, best.Result.Params.UseSmartFilters)
	}
}
