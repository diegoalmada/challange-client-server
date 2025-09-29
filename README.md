# Challenge: Cliente e Servidor (Go)

Este projeto implementa um pequeno serviço HTTP em Go que expõe a cotação USD->BRL e um cliente para consumir este serviço.

- Servidor: expõe o endpoint GET `/cotacao`, busca a cotação em um provedor externo, salva o resultado em um banco SQLite local (`app.db`) e retorna a cotação em JSON.
- Cliente: faz uma requisição ao servidor e imprime a resposta no stdout.

## Requisitos
- Go 1.20+ (recomendado 1.21 ou superior)
- Acesso à internet (o servidor consulta a API pública de câmbio)

## Como iniciar o servidor (server.go)

1. Instale dependências (o Go fará o download automaticamente via `go mod`):
   ```bash
   go mod download
   ```
2. Execute o servidor:
   ```bash
   go run server.go
   ```
3. O servidor subirá em `http://localhost:8080`.

Observações:
- O banco de dados SQLite é criado/atualizado automaticamente no arquivo `app.db` na raiz do projeto.
- O servidor utiliza a biblioteca GORM com o driver SQLite.

## Como iniciar o cliente (client.go)

Em um terminal separado, com o servidor já rodando:
```bash
go run client.go
```
O cliente fará uma requisição GET para `http://localhost:8080/cotacao` com timeout de ~300ms e imprimirá a resposta JSON no console.

## Endpoints da API

### GET /cotacao
Retorna a última cotação USD/BRL obtida no momento da requisição e a persiste no banco local.

- URL: `http://localhost:8080/cotacao`
- Método: `GET`
- Autenticação: Não requer
- Content-Type de resposta: `application/json`

#### Exemplo de requisição com cURL
```bash
curl -s http://localhost:8080/cotacao | jq .
```

#### Exemplo de resposta (200 OK)
```json
{
  "code": "USD",
  "name": "Dólar Americano/Real Brasileiro",
  "bid": 5.4321,
  "quotated_at": "2025-09-29T11:22:33Z"
}
```

#### Esquema de resposta
- `code` (string): Código da moeda base (ex.: "USD").
- `name` (string): Descrição da paridade (ex.: "Dólar Americano/Real Brasileiro").
- `bid` (number): Cotação de compra (convertida para `float64`).
- `quotated_at` (string, ISO-8601): Data/hora da cotação conforme fornecida pela API externa e parseada pelo servidor.

#### Códigos de status
- `200 OK`: Cotação obtida, persistida e retornada com sucesso.
- `408 Request Timeout`: Alguma operação (consulta externa ou persistência) excedeu o tempo limite (contexto cancelado/expirado).
- `500 Internal Server Error`: Erro interno ao processar a cotação (ex.: falhas de parse, falha de rede não mapeada como timeout, etc.).
