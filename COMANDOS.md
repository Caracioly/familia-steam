# Comandos Úteis - Família Steam

## 🔧 Desenvolvimento Local

### Instalar dependências
```bash
go mod download
```

### Compilar
```bash
go build -o app cmd/app/main.go
```

### Executar
```bash
go run cmd/app/main.go
```

### Testar compilação sem executar
```bash
go build ./...
```

## 🗄️ Banco de Dados

### Aplicar migrations (PostgreSQL)
```bash
# Via psql
psql "$DATABASE_URL" < migrations/001_init_schema.sql

# Ou usando script
./run-migrations.sh           # Linux/Mac
.\run-migrations.ps1          # Windows PowerShell
```

### Conectar ao banco local
```bash
psql "$DATABASE_URL"
```

### Verificar tabelas
```sql
\dt
```

### Ver schema de uma tabela
```sql
\d users
\d wallets
\d transactions
```

### Consultas úteis
```sql
-- Todos os usuários
SELECT * FROM users;

-- Saldo de um usuário
SELECT u.username, COALESCE(SUM(t.amount), 0) as saldo
FROM users u
LEFT JOIN wallets w ON w.user_id = u.id
LEFT JOIN transactions t ON t.wallet_id = w.id AND t.status = 'CONFIRMED'
WHERE u.discord_id = '123456789'
GROUP BY u.id, u.username;

-- Ranking completo
SELECT u.username, COALESCE(SUM(t.amount), 0) as saldo
FROM users u
INNER JOIN wallets w ON w.user_id = u.id
LEFT JOIN transactions t ON t.wallet_id = w.id AND t.status = 'CONFIRMED'
GROUP BY u.id, u.username
HAVING COALESCE(SUM(t.amount), 0) > 0
ORDER BY saldo DESC;

-- Transações pendentes
SELECT * FROM transactions WHERE status = 'PENDING';
```

## ☁️ Heroku

### Criar app
```bash
heroku create familia-steam
```

### Adicionar PostgreSQL
```bash
heroku addons:create heroku-postgresql:mini
```

### Ver variáveis de ambiente
```bash
heroku config
```

### Configurar variáveis
```bash
heroku config:set DISCORD_TOKEN="..."
heroku config:set MERCADOPAGO_ACCESS_TOKEN="..."
```

### Aplicar migrations
```bash
heroku pg:psql < migrations/001_init_schema.sql
```

### Conectar ao banco Heroku
```bash
heroku pg:psql
```

### Ver logs
```bash
heroku logs --tail
```

### Deploy
```bash
git push heroku main
```

### Restart
```bash
heroku restart
```

## 🧪 Testes de API

### Criar pagamento (cURL)
```bash
curl -X POST http://localhost:8080/api/payments/create \
  -H "Content-Type: application/json" \
  -d '{
    "discord_id": "123456789",
    "username": "TestUser",
    "amount": 10.50
  }'
```

### Consultar saldo
```bash
curl "http://localhost:8080/api/wallet/balance?discord_id=123456789"
```

### Ver ranking
```bash
curl "http://localhost:8080/api/wallet/ranking?limit=10"
```

### Simular webhook (teste local)
```bash
curl -X POST http://localhost:8080/api/payments/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "action": "payment.updated",
    "data": { "id": "12345" }
  }'
```

## 🤖 Comandos do Bot

```
!ping              # Testa se bot está online
!pix 10.50         # Gera pagamento de R$ 10,50
!saldo             # Consulta seu saldo
!saldo geral       # Consulta saldo total
!ranking           # Top 10 contribuidores
```

## 🔍 Debug

### Ver logs do app
```bash
# Local (imprime no terminal)
go run cmd/app/main.go

# Heroku
heroku logs --tail
```

### Verificar health
```bash
curl http://localhost:8080/health
```

### Verificar se porta está em uso (Windows)
```powershell
netstat -ano | findstr :8080
```

### Matar processo na porta 8080 (Windows)
```powershell
$port = 8080
$process = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
if ($process) {
    Stop-Process -Id $process.OwningProcess -Force
}
```

## 📦 Dependências

### Adicionar nova dependência
```bash
go get github.com/algum/pacote
go mod tidy
```

### Atualizar dependências
```bash
go get -u ./...
go mod tidy
```

### Limpar cache de módulos
```bash
go clean -modcache
```

## 🎯 Git

### Commit
```bash
git add .
git commit -m "feat: implementa sistema de carteira"
git push origin main
```

### Deploy Heroku via Git
```bash
git push heroku main
```

## 🔐 Segurança

### Nunca commitar .env
```bash
# Verificar se .env está no .gitignore
cat .gitignore | grep .env

# Se commitou por acidente
git rm --cached .env
git commit -m "Remove .env do versionamento"
```

### Rotacionar tokens
```bash
# Atualizar no Heroku
heroku config:set DISCORD_TOKEN="novo-token"
heroku config:set MERCADOPAGO_ACCESS_TOKEN="novo-token"

# Atualizar local
# Edite .env manualmente
```
