# LuckyFive Playbook: Guia Completo de Uso

## Introdução

LuckyFive é um sistema de análise de loteria desenvolvido em Go que permite processar resultados históricos, executar simulações de apostas e analisar padrões estatísticos. Este playbook guia você passo a passo pelo uso completo do sistema, desde a configuração inicial até o monitoramento de simulações assíncronas.

## Pré-requisitos

- **Go 1.21+**: Linguagem de programação principal.
- **SQLite**: Banco de dados para armazenamento local.
- **curl**: Para testar as APIs (ou qualquer cliente HTTP).
- **jq** (opcional): Para formatar respostas JSON nos exemplos.

## Instalação e Configuração

### 1. Clonagem e Build

```bash
# Clone o repositório
git clone https://github.com/garnizeh/luckyfive.git
cd luckyfive

# Build dos binários
make build
```

Isso gera os executáveis em `bin/`: `api`, `migrate`, `worker`, `import`.

### 2. Configuração

Copie e ajuste o arquivo de configuração:

```bash
cp configs/dev.env .env
# Edite .env se necessário (padrão: localhost:8080)
```

## Inicialização do Sistema

### 1. Migrações do Banco

Execute as migrações para criar as tabelas:

```bash
./bin/migrate up --env-file=configs/dev.env
```

**Exemplo de saída esperada:**
```
INFO migrating up db=data/db/results.db
INFO applying migration version=1 file=001_create_results.sql
...
INFO migrating up db=data/db/simulations.db
...
```

### 2. Iniciar Worker (Opcional, para simulações assíncronas)

```bash
make run-worker
```

**Exemplo de log:**
```
{"time":"2025-11-22T00:12:21.882","level":"INFO","msg":"Starting worker","worker_id":"worker-xxx"}
```

### 3. Iniciar API

```bash
make run-api
```

A API estará disponível em `http://localhost:8080`.

## Fluxos Principais

### 1. Upload e Importação de Dados

#### Upload do Arquivo XLSX

```bash
curl -X POST \
  -F "file=@data/results/Quina.xlsx" \
  http://localhost:8080/api/v1/results/upload
```

**Resposta esperada:**
```json
{
  "artifact_id": "abc123",
  "filename": "Quina.xlsx",
  "size": 499507,
  "message": "File uploaded successfully"
}
```

#### Importação dos Dados

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"artifact_id": "abc123"}' \
  http://localhost:8080/api/v1/results/import
```

**Resposta esperada:**
```json
{
  "artifact_id": "abc123",
  "filename": "Quina.xlsx",
  "imported_at": "2025-11-22T00:12:32Z",
  "rows_inserted": 6882,
  "rows_skipped": 0,
  "rows_errors": 0,
  "duration": "9.437s",
  "message": "Import completed successfully"
}
```

### 2. Consulta de Resultados

#### Listar Resultados (com paginação)

```bash
curl "http://localhost:8080/api/v1/results?limit=5&offset=0"
```

**Resposta esperada:**
```json
{
  "draws": [
    {
      "contest": 6882,
      "draw_date": "2025-11-19T00:00:00Z",
      "bola1": 16,
      "bola2": 18,
      "bola3": 36,
      "bola4": 60,
      "bola5": 80,
      "source": "xlsx:QUINA"
    }
  ],
  "limit": 5,
  "offset": 0,
  "count": 5
}
```

#### Obter Resultado Específico

```bash
curl http://localhost:8080/api/v1/results/6882
```

### 3. Simulações Simples (com Presets)

#### Executar Simulação Síncrona

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "preset": "conservative",
    "start_contest": 6800,
    "end_contest": 6880,
    "async": false
  }' \
  http://localhost:8080/api/v1/simulations/simple
```

**Resposta esperada:**
```json
{
  "id": 1,
  "status": "completed",
  "recipe_name": "conservative",
  "summary_json": "{\"TotalContests\":81,\"QuinaHits\":0,\"QuadraHits\":0,\"TernoHits\":1,\"AverageHits\":0.037}"
}
```

#### Executar Simulação Assíncrona

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "preset": "balanced",
    "start_contest": 6800,
    "end_contest": 6880,
    "async": true
  }' \
  http://localhost:8080/api/v1/simulations/simple
```

**Resposta esperada:**
```json
{
  "simulation_id": 2,
  "status": "pending",
  "message": "Simulation queued for processing"
}
```

### 4. Simulações Avançadas

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "recipe": {
      "version": "1.0",
      "name": "custom",
      "parameters": {
        "alpha": 0.4,
        "beta": 0.3,
        "gamma": 0.2,
        "delta": 0.1,
        "sim_prev_max": 400,
        "sim_preds": 25
      }
    },
    "start_contest": 6800,
    "end_contest": 6880,
    "async": false
  }' \
  http://localhost:8080/api/v1/simulations/advanced
```

### 5. Monitoramento de Simulações

#### Listar Simulações

```bash
curl "http://localhost:8080/api/v1/simulations?limit=10&offset=0"
```

#### Ver Detalhes de uma Simulação

```bash
curl http://localhost:8080/api/v1/simulations/1
```

#### Ver Resultados Detalhados de uma Simulação

```bash
curl "http://localhost:8080/api/v1/simulations/1/results?limit=50&offset=0"
```

#### Cancelar Simulação (se pendente)

```bash
curl -X POST http://localhost:8080/api/v1/simulations/1/cancel
```

### 6. Monitoramento de Simulações Assíncronas

Para simulações assíncronas, monitore o status periodicamente:

```bash
# Verificar status
curl http://localhost:8080/api/v1/simulations/2 | jq '.status'

# Aguardar conclusão (exemplo em script)
while [ "$(curl -s http://localhost:8080/api/v1/simulations/2 | jq -r '.status')" != "completed" ]; do
  echo "Aguardando..."
  sleep 5
done
echo "Simulação concluída!"
```

**Logs do Worker:**
- Inicie o worker em um terminal separado.
- Monitore os logs para ver processamento de jobs:
  ```
  {"time":"...","level":"INFO","msg":"processing job","job_id":2}
  {"time":"...","level":"INFO","msg":"job completed","job_id":2}
  ```

## Fluxos Adicionais

### Health Check

```bash
curl http://localhost:8080/health
```

### Configurações

#### Listar Configs

```bash
curl http://localhost:8080/api/v1/configs
```

#### Criar Config Personalizada

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my_config",
    "description": "Minha configuração",
    "recipe": {
      "version": "1.0",
      "name": "my_config",
      "parameters": {
        "alpha": 0.35,
        "beta": 0.3,
        "gamma": 0.25,
        "delta": 0.1,
        "sim_prev_max": 500,
        "sim_preds": 20
      }
    }
  }' \
  http://localhost:8080/api/v1/configs
```

## Troubleshooting

### Problemas Comuns

1. **Erro de migração**: Verifique se o diretório `data/db/` existe e tem permissões.
2. **API não responde**: Confirme se a porta 8080 está livre e o binário foi buildado.
3. **Simulação falha**: Verifique se os dados foram importados corretamente.
4. **Worker não processa**: Certifique-se de que o worker está rodando e conectado ao mesmo banco.

### Logs

- **API**: Logs aparecem no terminal onde foi iniciado.
- **Worker**: Logs no terminal do worker.
- **Debug**: Use `LOG_LEVEL=DEBUG` no `.env`.

### Reset do Sistema

Para limpar tudo e recomeçar:

```bash
make reset-db
make build
```

## Conclusão

Este playbook cobre todos os fluxos principais do LuckyFive. Comece com migrações e upload de dados, depois explore simulações simples e avançadas. Para produção, considere configurações de ambiente e monitoramento contínuo.