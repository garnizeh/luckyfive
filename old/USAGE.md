Uso do programa LuckyFive (Quina)

Descrição

Esse CLI em Go carrega os dados históricos de concursos da Quina a partir de um arquivo Excel (opcional) e os armazena em um banco SQLite (arquivo ou in-memory). Ele oferece:
- Geração de jogos estatísticos baseados em frequências
- Gerador avançado híbrido (frequência, recência, co-ocorrência, posicionamento, cluster) com otimizações por hill-climbing e uma pequena GA
- Simulação (backtest) em intervalos de concursos para avaliar previsões

Como executar (exemplos)

1) Rodar com a planilha local (in-memory) — carregamento automático

```bash
cd /g/code/Go/copilot/luckyfive
# carrega data/Quina.xlsx em memória, gera 25 jogos, e mostra as previsões
go run main.go
```

2) Usar um banco SQLite em arquivo e pular o carregamento (útil para re-testes)

```bash
# usa o banco data/dados.db existente, não carrega a planilha novamente
go run main.go --db data/dados.db --load=false
```

3) Rodar uma simulação entre concursos

```bash
# simula previsões (10 por padrão) nos concursos de 6877 a 6881 usando o banco em arquivo
go run main.go --db data/dados.db --load=false --sim-start=6877 --sim-end=6881
```

4) Rodar com o gerador avançado e parâmetros para otimizar as previsões

```bash
# gera 20 previsões por concurso, aumenta o boost para números frios e sintoniza parâmetros de scoring e busca
go run main.go \
  --db data/dados.db --load=false --sim-start=6877 --sim-end=6881 \
  --sim-preds=20 \
  --hot-cold-boost=3.0 \
  --alpha=2.0 --beta=0.2 --gamma=1.0 --cluster-penalty=0.5 \
  --cands-mult=200 --hill-iter=100 --coocc-window=50 --hot-window=10
```

5) Rodar uma execução rápida para debug (menos candidatos, menos iterações)

```bash
go run main.go --db data/dados.db --load=false --sim-start=6877 --sim-end=6881 \
  --sim-preds=10 --cands-mult=20 --hill-iter=10
```

6) Melhor qualidade (rodar por mais tempo com mais candidatos e iterações)

```bash
go run main.go --db data/dados.db --load=false --sim-start=6500 --sim-end=6881 \
  --sim-preds=50 --cands-mult=1000 --hill-iter=400 \
  --alpha=3.0 --beta=0.2 --gamma=1.1 --cluster-penalty=0.8
```

Parâmetros (flags) aceitos

- --file (string) — Caminho para o arquivo .xlsx com histórico. Padrão: data/Quina.xlsx
- --sheet (string) — Nome da planilha a usar (opcional). Se vazio, será usada a primeira.
- --db (string) — Caminho para o arquivo SQLite. Se vazio, o programa usa um banco em memória (file:quinadb?mode=memory&cache=shared).
- --load (bool) — Se true, os dados serão carregados do .xlsx para o banco. Se usar banco em arquivo e já tiver dados, use false para pular. Default: true.

- --sim-start (int) — Concurso inicial para simulação. (Ex: 6877). Se não fornecido, não roda simulação.
- --sim-end (int) — Concurso final para simulação.
- --sim-prev-max (int) — Quantidade máxima de concursos anteriores (history window) a usar por cada concurso analisado. Padrão: 100.

- --sim-preds (int) — Quantidade de previsões a gerar por concurso na simulação (default: 10).

Gerador avançado (hiperparâmetros)

- --lambda (float) — fator de decaimento exponencial para recência; números recentes têm peso maior. Range típico: 0.01-0.5. Default: 0.05.
- --hot-cold-boost (float) — multiplicador para 'cold' numbers (número que NÃO apareceu nas últimas 'hot-window' extrações). Aumenta a chance de incluir "frios" no prediction. Default: 2.0.
- --alpha (float) — peso para co-ocorrência de pares (condicional P(b|a)). Default: 1.0.
- --beta (float) — peso para frequência marginal. Default: 0.3.
- --gamma (float) — peso para afinidade posicional (posição do número no jogo). Default: 1.0.
- --cluster-penalty (float) — penaliza clusters (muitos números na mesma dezena). Range: 0-2.0, default: 0.5.
- --coocc-window (int) — número de concursos recentes a usar para cálculo de co-ocorrência (ex.: 50). Default: 50.
- --hot-window (int) — quantos concursos contar para determinar se um número é 'recente' (hot). Default: 10.
- --cands-mult (int) — multiplicador de quantos candidatos gerar por previsão antes de refinar (ex: 200 -> 200 * numPredictions candidates). Higher = more exploration, more cost. Default: 200.
- --hill-iter (int) — iterações para refinamento local (hill-climb). Default: 50.
 - --sim-save (bool) — flag para salvar os resultados da simulação em arquivos JSON/CSV no diretório especificado por --sim-out. Default: true.
 - --sim-out (string) — diretório para salvar resultados da simulação. Default: data/simulations.

Comportamento esperado da simulação

- O programa carrega os dados (se --load=true), imprime preview do banco, gera 25 jogos estatísticos (função GerarJogosEstatisticos, que usa apenas frequências) e, se solicitado, executa simulações no intervalo de concurso especificado.
- Para cada concurso simulado: carrega a combinação real do concurso e usa os dados anteriores (limitado por --sim-prev-max) para gerar previsões. O algoritmo avançado gera e refina muitas combinações, pontua-as e seleciona um conjunto final de previsões otimizadas.
- Na execução atual, a simulação imprime previsões que obtiveram >= 2 acertos para facilitar a visualização. O melhor resultado por concurso também é impresso.

Boas práticas e dicas de performance

- Para desenvolvimento e debug: reduza --cands-mult (e.g. 10-50) e --hill-iter (10) para testar rapidamente.
- Para melhor qualidade: aumente --cands-mult (100 - 1000) e --hill-iter (50 - 400). Isso aumenta custo computacional, mas melhora chances.
- Use um banco em arquivo (--db data/dados.db) e carregue os dados uma vez (--load=true) para acelerar repetidos testes (--load=false).
- Ajuste --lambda para enfatizar recência (makes recent numbers much more likely). Valores pequenos (0.01) dão pouca preferência à recência, valores maiores (0.1-0.5) aumentam recency effect.
- Ajuste --alpha / --beta para balancear co-ocorrência x frequência. Aumente alpha para reforçar pares reais que co-ocorrem frequentemente.

Notas finais

- A natureza do jogo é naturalmente aleatória; mesmo algoritmos bem projetados podem acertar pouco em amostras curtas. O objetivo deste gerador é aumentar a probabilidade estatística de acertos, otimizar por co-ocorrência e recency, e oferecer a ferramenta para backtesting e exploração de estratégias.
- Se quiser, eu posso:
  - Implementar uma saída CSV com métricas de performance (avg hits, hit distribution)
  - Adicionar grid-search/hyperparameter tuning automático (executar testes rápidos e escolher melhores parâmetros)
  - Exportar previsões para arquivo CSV para análise posterior

---

Formato e arquivos gerados pela simulação

- Arquivos gerados: Para cada simulação (intervalo de concursos e parâmetros), o programa gravará na pasta `--sim-out`:
  - JSON detalhado: `sim_<start>-<end>_params..._<timestamp>.json` — contém os parâmetros da simulação, os resultados por concurso (cada previsão, hits) e um resumo estatístico.
  - CSV (detalhes): `..._details.csv` — linhas: concurso, índice da previsão, lista de números previstos, hits, resultado real (coluna `actual`). Esse arquivo é útil para analisar previsões por linha e por concurso.
  - CSV (summary): `..._summary.csv` — linhas: concurso, melhor acerto (best_hits), índice da previsão vencedora e a combinação real. Esse arquivo é útil para agregar e comparar estratégias rapidamente.

Exemplo de nome de arquivo:
`sim_6877-6881_preds20_alpha2p00_beta0p20_lambda0p050_hotcold3p00_gamma1p00_cluster0p50_cands200_hill100_coocc50_hotw10_20251119T200119.json`

Esses arquivos são projetados para serem fáceis de comparar entre diferentes execuções (parâmetros no nome do arquivo) e para permitir uma análise estatística posterior (métricas no JSON + CSV para tabelas/BI).

Se quiser, eu adiciono agora: sample CSV export e/ou um comando para gravar resultados da simulação em arquivo Y (CSV) e a opção de usar multiplas estratégias em ensemble.