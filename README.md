# Família Steam

Bot do Discord com API HTTP e banco de dados PostgreSQL, usando arquitetura limpa e preparado para deploy no Heroku.

## 🚀 Tecnologias

- Go 1.21
- PostgreSQL (Heroku)
- Discord.js (discordgo)
- net/http (servidor HTTP nativo)

## 📁 Estrutura do Projeto

```
familia-steam/
├── cmd/app/main.go          # Ponto de entrada da aplicação
├── internal/
│   ├── config/config.go     # Gerenciamento de configurações
│   ├── db/postgres.go       # Conexão com PostgreSQL
│   ├── bot/bot.go           # Bot do Discord
│   └── api/server.go        # Servidor HTTP
├── .env                     # Variáveis de ambiente (local)
├── .env.example             # Template de variáveis
├── go.mod                   # Dependências Go
└── Procfile                 # Configuração Heroku
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
   ```

3. **Instale as dependências:**
   ```bash
   go mod download
   ```

4. **Execute localmente:**
   ```bash
   # PowerShell (Windows)
   Get-Content .env | ForEach-Object {
       if ($_ -match '^([^=]+)=(.*)$') {
           [Environment]::SetEnvironmentVariable($matches[1], $matches[2])
       }
   }
   go run cmd/app/main.go
   
   # Ou compile e execute
   go build -o app cmd/app/main.go
   ./app
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

3. **Configure o token do Discord:**
   ```bash
   heroku config:set DISCORD_TOKEN="seu-token-aqui"
   ```

4. **Deploy:**
   ```bash
   git push heroku main
   ```

5. **Verifique os logs:**
   ```bash
   heroku logs --tail
   ```

## 🔍 Endpoints da API

- `GET /` - Rota raiz
- `GET /health` - Health check (verifica banco de dados)

## 🤖 Comandos do Bot

- `!ping` - Responde "Pong!" (exemplo básico)

Para adicionar mais comandos, edite `internal/bot/bot.go` no método `onMessageCreate`.

## 📝 Notas

- O projeto usa apenas bibliotecas padrão do Go, exceto `discordgo` e `lib/pq`
- Shutdown gracioso implementado (SIGINT/SIGTERM)
- Pool de conexões PostgreSQL configurado automaticamente
- SSL obrigatório para conexões de banco (Heroku)