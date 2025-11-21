package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

func main() {
    dbPath := flag.String("db", "data/dados.db", "Caminho para o sqlite DB")
    cardsFile := flag.String("cards-file", "data/generated_6882.txt", "Arquivo com os jogos gerados")
    contest := flag.Int("contest", 6882, "Número do concurso a comparar")
    flag.Parse()

    db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?cache=shared&mode=rwc", *dbPath))
    if err != nil {
        fmt.Fprintf(os.Stderr, "erro abrir db: %v\n", err)
        os.Exit(1)
    }
    defer db.Close()

    tableName := "quina"
    contestCol, numCols, err := detectSchema(db, tableName)
    if err != nil {
        fmt.Fprintf(os.Stderr, "erro detect schema: %v\n", err)
        os.Exit(1)
    }

    actual, err := loadDraw(db, tableName, contestCol, numCols, *contest)
    if err != nil {
        fmt.Fprintf(os.Stderr, "erro carregar concurso %d: %v\n", *contest, err)
        os.Exit(1)
    }
    sort.Ints(actual)
    fmt.Printf("Concurso %d atual: %v\n", *contest, actual)

    // open cards file and parse lines like: Jogo #01: [18 26 44 57 61]
    f, err := os.Open(*cardsFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "erro abrir arquivo de cartoes: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    re := regexp.MustCompile(`Jogo #\d+: \[(.*?)\]`)
    cards := [][]int{}
    for scanner.Scan() {
        line := scanner.Text()
        m := re.FindStringSubmatch(line)
        if len(m) == 2 {
            parts := strings.Fields(m[1])
            nums := []int{}
            for _, p := range parts {
                if p == "" {
                    continue
                }
                n, err := strconv.Atoi(strings.TrimSpace(p))
                if err != nil {
                    continue
                }
                nums = append(nums, n)
            }
            if len(nums) > 0 {
                cards = append(cards, nums)
            }
        }
    }

    if err := scanner.Err(); err != nil {
        fmt.Fprintf(os.Stderr, "erro ler arquivo: %v\n", err)
        os.Exit(1)
    }

    if len(cards) == 0 {
        fmt.Fprintf(os.Stderr, "nenhum cartao encontrado em %s\n", *cardsFile)
        os.Exit(2)
    }

    counts := map[int]int{} // hits -> count
    for i, c := range cards {
        sort.Ints(c)
        hits, hitNums := compareGame(actual, c)
        counts[hits]++
        fmt.Printf("Card #%02d: %v -> hits=%d (matched: %v)\n", i+1, c, hits, hitNums)
    }

    fmt.Println("\nResumo:")
    fmt.Printf("Total cartoes: %d\n", len(cards))
    fmt.Printf("Duque (2 acertos): %d\n", counts[2])
    fmt.Printf("Terno (3 acertos): %d\n", counts[3])
    fmt.Printf("Quadra (4 acertos): %d\n", counts[4])
    fmt.Printf("Quina (5 acertos): %d\n", counts[5])

}

// --- Local helper functions (copied/adapted from main.go) ---
func detectSchema(db *sql.DB, tableName string) (contestCol string, numCols []string, err error) {
    rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s);", tableName))
    if err != nil {
        return "", nil, err
    }
    defer rows.Close()
    var cid int
    var name, ctype string
    var notnull, dflt_value, pk interface{}
    cols := []string{}
    for rows.Next() {
        if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
            return "", nil, err
        }
        cols = append(cols, name)
    }
    // find contest column
    for _, c := range cols {
        lc := strings.ToLower(c)
        if strings.Contains(lc, "conc") || strings.Contains(lc, "concurso") || strings.Contains(lc, "id") {
            contestCol = c
            break
        }
    }
    if contestCol == "" && len(cols) > 0 {
        contestCol = cols[0]
    }

    lower := map[string]string{}
    for _, c := range cols {
        lower[strings.ToLower(c)] = c
    }
    want := []string{"bola1", "bola2", "bola3", "bola4", "bola5"}
    ok := true
    for _, w := range want {
        if _, found := lower[w]; !found {
            ok = false
            break
        }
    }
    if ok {
        for _, w := range want {
            numCols = append(numCols, lower[w])
        }
        return contestCol, numCols, nil
    }

    for _, c := range cols {
        lc := strings.ToLower(c)
        if strings.Contains(lc, "bola") || strings.HasPrefix(lc, "b") || strings.Contains(lc, "num") || strings.Contains(lc, " n ") || strings.HasPrefix(lc, "n") {
            if c != contestCol {
                numCols = append(numCols, c)
            }
        }
    }
    if len(numCols) == 0 {
        for _, c := range cols {
            if c != contestCol {
                numCols = append(numCols, c)
            }
        }
    }
    return contestCol, numCols, nil
}

func loadDraw(db *sql.DB, tableName, contestCol string, numCols []string, contest interface{}) ([]int, error) {
    cols := strings.Join(numCols, ", ")
    query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ? LIMIT 1;", cols, tableName, contestCol)
    row := db.QueryRow(query, contest)
    vals := make([]interface{}, len(numCols))
    ptrs := make([]interface{}, len(numCols))
    for i := range vals {
        ptrs[i] = &vals[i]
    }
    if err := row.Scan(ptrs...); err != nil {
        return nil, err
    }
    res := []int{}
    for _, v := range vals {
        if v == nil {
            continue
        }
        switch t := v.(type) {
        case int64:
            res = append(res, int(t))
        case int:
            res = append(res, t)
        case []uint8:
            s := string(t)
            var n int
            fmt.Sscanf(s, "%d", &n)
            if n > 0 {
                res = append(res, n)
            }
        case string:
            var n int
            fmt.Sscanf(t, "%d", &n)
            if n > 0 {
                res = append(res, n)
            }
        default:
            var n int
            fmt.Sscanf(fmt.Sprintf("%v", v), "%d", &n)
            if n > 0 {
                res = append(res, n)
            }
        }
    }
    return res, nil
}

func compareGame(actual, pred []int) (int, []int) {
    set := map[int]bool{}
    for _, a := range actual {
        set[a] = true
    }
    hits := []int{}
    for _, p := range pred {
        if set[p] {
            hits = append(hits, p)
        }
    }
    return len(hits), hits
}
