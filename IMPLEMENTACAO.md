# Implementação da Carteira Comunitária - Resumo

## ✅ Implementado

### 1. Estrutura de Banco de Dados
- ✅ Tabela `users` (Discord ID, username)
- ✅ Tabela `wallets` (1 por usuário)
- ✅ Tabela `transactions` (PENDING → CONFIRMED)
- ✅ Índices para performance
- ✅ Migration SQL em `migrations/001_init_schema.sql`

### 2. Repositories (Camada de Dados)
- ✅ `UserRepository` - CRUD de usuários
- ✅ `WalletRepository` - Carteiras + cálculo de saldo + ranking
- ✅ `TransactionRepository` - Transações com idempotência

### 3. Client Mercado Pago
- ✅ `CreatePixPayment()` - Gera QR Code PIX
- ✅ `GetPayment()` - Consulta status de pagamento
- ✅ Estruturas de request/response completas
- ✅ Tratamento de erros

### 4. Services (Lógica de Negócio)
- ✅ `PaymentService`:
  - `CreatePixPayment()` - Cria pagamento + transação PENDING
  - `ConfirmPayment()` - Confirma pagamento (webhook)
- ✅ `WalletService`:
  - `GetUserBalance()` - Saldo por Discord ID
  - `GetTotalBalance()` - Saldo total da vaquinha
  - `GetRanking()` - Top N contribuidores

### 5. API HTTP (Endpoints)
- ✅ `POST /api/payments/create` - Cria pagamento PIX
- ✅ `POST /api/payments/webhook` - Webhook Mercado Pago
- ✅ `GET /api/wallet/balance?discord_id=X` - Consulta saldo
- ✅ `GET /api/wallet/ranking?limit=10` - Ranking
- ✅ Validações de input
- ✅ Tratamento de erros

### 6. Bot Discord (Interface)
- ✅ Comando `!pix <valor>`:
  - Valida valor numérico > 0
  - Chama API para criar pagamento
  - Retorna QR Code copia-e-cola formatado
- ✅ Comando `!saldo` - Saldo pessoal
- ✅ Comando `!saldo geral` - Saldo total
- ✅ Comando `!ranking` - Top 10 com medalhas
- ✅ Bot não contém lógica de negócio (apenas chama API)

### 7. Configuração
- ✅ Variável `MERCADOPAGO_ACCESS_TOKEN` adicionada
- ✅ `.env` e `.env.example` atualizados
- ✅ README completo com instruções

### 8. Integração
- ✅ `main.go` atualizado com toda a injeção de dependências
- ✅ Fluxo completo funcional
- ✅ Shutdown gracioso mantido

## 🎯 Arquitetura Limpa

```
┌─────────────┐
│   Discord   │  (!pix, !saldo, !ranking)
└──────┬──────┘
       │ HTTP
┌──────▼──────┐
│   Bot       │  (Sem lógica de negócio)
└──────┬──────┘
       │ HTTP
┌──────▼──────┐
│   API       │  (Endpoints REST)
└──────┬──────┘
       │
┌──────▼──────┐
│  Services   │  (Regras de negócio)
└──────┬──────┘
       │
┌──────▼──────┐
│Repositories │  (Acesso aos dados)
└──────┬──────┘
       │
┌──────▼──────┐
│  PostgreSQL │
└─────────────┘
```

## 🔄 Fluxo de Pagamento

1. **Usuário no Discord:**
   ```
   !pix 50.00
   ```

2. **Bot:**
   - Valida valor
   - POST http://localhost:8080/api/payments/create
   ```json
   {
     "discord_id": "123456789",
     "username": "João",
     "amount": 50.00
   }
   ```

3. **API → Service → Repository:**
   - Cria/busca usuário
   - Cria/busca carteira
   - Chama Mercado Pago
   - Cria transação PENDING

4. **Mercado Pago:**
   - Retorna QR Code PIX

5. **API → Bot → Discord:**
   ```
   💰 Pagamento PIX criado!
   Valor: R$ 50.00
   PIX Copia e Cola:
   00020126580014br.gov.bcb.pix...
   ```

6. **Usuário paga via app bancário**

7. **Mercado Pago → Webhook:**
   ```
   POST /api/payments/webhook
   {
     "action": "payment.updated",
     "data": { "id": "12345" }
   }
   ```

8. **API:**
   - Busca transação por external_reference
   - Atualiza status → CONFIRMED (idempotente)
   - Saldo disponível automaticamente

## 📊 Decisões de Design

### Idempotência
- Webhook pode ser chamado múltiplas vezes
- `UpdateStatus()` verifica se já está CONFIRMED antes de atualizar

### Separação de Responsabilidades
- Bot: apenas interface Discord → HTTP
- API: validações + orquestração
- Services: lógica de negócio
- Repositories: SQL puro

### Cálculo de Saldo
- Saldo = SUM(transactions.amount WHERE status = 'CONFIRMED')
- Sempre calculado on-demand (fonte única de verdade)
- Nenhuma coluna `balance` denormalizada

### Security
- Webhook não valida assinatura (pode ser adicionado)
- DATABASE_URL com SSL obrigatório
- Tokens via variáveis de ambiente

## 🚀 Próximos Passos

1. **Aplicar migrations:**
   ```bash
   psql "$DATABASE_URL" < migrations/001_init_schema.sql
   ```

2. **Configurar Mercado Pago:**
   - Obter Access Token
   - Adicionar em `.env`

3. **Testar localmente:**
   ```bash
   go run cmd/app/main.go
   ```

4. **Testar comandos:**
   - `!pix 10.00`
   - `!saldo`
   - `!ranking`

5. **Deploy Heroku:**
   ```bash
   heroku config:set MERCADOPAGO_ACCESS_TOKEN="..."
   heroku pg:psql < migrations/001_init_schema.sql
   git push heroku main
   ```

## ⚠️ Importante

- **Discord Bot Intents:** Ativar MESSAGE CONTENT INTENT no painel
- **Webhook URL:** Configurar no Mercado Pago apontando para `/api/payments/webhook`
- **PostgreSQL:** Migrations devem ser aplicadas antes do primeiro uso
