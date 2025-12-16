# Família Steam

Bot do Discord com carteira comunitária (vaquinha) integrada ao Mercado Pago, usando arquitetura limpa e pronto para deploy no Heroku.

## 🚀 Tecnologias

- Go 1.21
- PostgreSQL (Heroku)
- Discord Bot (discordgo)
- Mercado Pago API (pagamentos PIX)
- net/http (servidor HTTP nativo)

## 📁 Estrutura do Projeto

```
familia-steam/
├── cmd/app/main.go              # Ponto de entrada
├── internal/
│   ├── config/                  # Configurações
│   ├── db/                      # Conexão PostgreSQL
│   ├── repository/              # Camada de dados (users, wallets, transactions)
│   ├── service/                 # Lógica de negócio (payment, wallet)
│   ├── mercadopago/             # Client Mercado Pago
│   ├── bot/                     # Bot Discord (sem lógica de negócio)
│   └── api/                     # Endpoints HTTP
├── migrations/                  # SQL migrations
├── .env                         # Variáveis de ambiente (local)
└── Procfile                     # Configuração Heroku
```

## ⚙️ Configuração Local

1. **Copie o arquivo de exemplo:**
   ```bash
   cp .env.example .env
   ```

2. **Edite `.env` com suas credenciais:**
   ```bash
   PORT=8080
   DATABASE_URL=postgres://user:password@host:5432/database?sslmode=require
   DISCORD_TOKEN=seu-token-do-discord
   MERCADOPAGO_ACCESS_TOKEN=seu-access-token-do-mercadopago
   ```

3. **Instale as dependências:**
   ```bash
   go mod download
   ```

4. **Aplique as migrations no banco:**
   ```bash
   # Linux/Mac
   chmod +x run-migrations.sh
   ./run-migrations.sh
   
   # Ou manualmente via psql
   psql "$DATABASE_URL" < migrations/001_init_schema.sql
   ```

5. **Execute localmente:**
   ```bash
   go run cmd/app/main.go
   ```

## 🐳 Deploy no Heroku

1. **Crie o app no Heroku:**
   ```bash
   heroku create familia-steam
   ```

2. **Adicione o PostgreSQL:**
   ```bash
   heroku addons:create heroku-postgresql:mini
   ```

3. **Configure as variáveis de ambiente:**
   ```bash
   heroku config:set DISCORD_TOKEN="seu-token-aqui"
   heroku config:set MERCADOPAGO_ACCESS_TOKEN="seu-token-mercadopago"
   ```

4. **Aplique as migrations:**
   ```bash
   # Via heroku CLI
   heroku pg:psql < migrations/001_init_schema.sql
   ```

5. **Deploy:**
   ```bash
   git push heroku main
   ```

6. **Verifique os logs:**
   ```bash
   heroku logs --tail
   ```

## 🔍 Endpoints da API

### Pagamentos
- `POST /api/payments/create` - Cria pagamento PIX
- `POST /api/payments/webhook` - Webhook Mercado Pago

### Carteira
- `GET /api/wallet/balance?discord_id=<id>` - Consulta saldo
- `GET /api/wallet/ranking?limit=10` - Ranking de contribuidores

### Sistema
- `GET /health` - Health check
- `GET /` - Informações da API

## 🤖 Comandos do Bot

### Pagamentos
- `!pix <valor>` - Gera QR Code PIX para contribuir
  - Exemplo: `!pix 10.50`
  - Retorna QR Code copia-e-cola

### Consultas
- `!saldo` - Consulta seu saldo pessoal
- `!saldo geral` - Consulta saldo total da vaquinha
- `!ranking` - Top 10 contribuidores

### Teste
- `!ping` - Verifica se o bot está online

## 🔐 Configuração do Discord Bot

1. Acesse https://discord.com/developers/applications
2. Selecione seu bot
3. Vá em **Bot** → **Privileged Gateway Intents**
4. Ative **MESSAGE CONTENT INTENT**
5. Salve as alterações

## 💳 Configuração do Mercado Pago

1. Crie uma conta em https://mercadopago.com.br
2. Acesse https://www.mercadopago.com.br/developers/panel/app
3. Crie uma aplicação
4. Copie o **Access Token** (Production ou Test)
5. Configure em `MERCADOPAGO_ACCESS_TOKEN`

## 🗄️ Banco de Dados

### Tabelas
- `users` - Usuários do Discord
- `wallets` - Carteiras (1 por usuário)
- `transactions` - Transações (status: PENDING → CONFIRMED)

### Fluxo de Pagamento
1. Usuário executa `!pix 10.50`
2. API cria transação PENDING no banco
3. API chama Mercado Pago e gera QR Code
4. Bot retorna QR Code para o usuário
5. Usuário paga via PIX
6. Mercado Pago envia webhook
7. API atualiza transação para CONFIRMED
8. Saldo é creditado automaticamente

## 📝 Notas

- O projeto usa apenas bibliotecas padrão do Go, exceto `discordgo` e `lib/pq`
- Shutdown gracioso implementado (SIGINT/SIGTERM)
- Pool de conexões PostgreSQL configurado automaticamente
- SSL obrigatório para conexões de banco (Heroku)