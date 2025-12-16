package bot

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session *discordgo.Session
	apiURL  string
}

func New(token, apiURL string) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar sessão do Discord: %w", err)
	}

	// Configura as intents necessárias para ler mensagens
	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	bot := &Bot{
		session: session,
		apiURL:  apiURL,
	}

	bot.registerHandlers()

	return bot, nil
}

func (b *Bot) Start() error {
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("erro ao conectar ao Discord: %w", err)
	}

	log.Println("Bot do Discord conectado")
	return nil
}

func (b *Bot) Stop() error {
	log.Println("Encerrando bot do Discord...")
	return b.session.Close()
}

func (b *Bot) registerHandlers() {
	b.session.AddHandler(b.onReady)
	b.session.AddHandler(b.onMessageCreate)
}

func (b *Bot) onReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Printf("Bot logado como %s#%s", event.User.Username, event.User.Discriminator)
}

func (b *Bot) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	content := strings.TrimSpace(m.Content)

	if content == "!ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
		return
	}

	if strings.HasPrefix(content, "!pix ") {
		b.handlePixCommand(s, m)
		return
	}

	if content == "!saldo" {
		b.handleBalanceCommand(s, m)
		return
	}

	if content == "!saldo geral" {
		b.handleTotalBalanceCommand(s, m)
		return
	}

	if content == "!ranking" {
		b.handleRankingCommand(s, m)
		return
	}
}

func (b *Bot) handlePixCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	parts := strings.Fields(m.Content)
	if len(parts) != 2 {
		s.ChannelMessageSend(m.ChannelID, "❌ Uso correto: `!pix <valor>`\nExemplo: `!pix 10.50`")
		return
	}

	amount, err := strconv.ParseFloat(parts[1], 64)
	if err != nil || amount <= 0 {
		s.ChannelMessageSend(m.ChannelID, "❌ Valor inválido. Use um número maior que zero.\nExemplo: `!pix 10.50`")
		return
	}

	reqBody := map[string]interface{}{
		"discord_id": m.Author.ID,
		"username":   m.Author.Username,
		"amount":     amount,
	}
	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post(b.apiURL+"/api/payments/create", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Erro ao criar pagamento: %v", err)
		s.ChannelMessageSend(m.ChannelID, "❌ Erro ao criar pagamento. Tente novamente.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.ChannelMessageSend(m.ChannelID, "❌ Erro ao criar pagamento. Tente novamente.")
		return
	}

	var payment struct {
		TransactionID int64   `json:"transaction_id"`
		Amount        float64 `json:"amount"`
		QRCodeBase64  string  `json:"qr_code_base64"`
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &payment)

	log.Printf("Payment response: TransactionID=%d, Amount=%.2f, QRCodeBase64 length=%d",
		payment.TransactionID, payment.Amount, len(payment.QRCodeBase64))

	if payment.QRCodeBase64 == "" {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
			"❌ QR Code PIX não disponível.\n"+
				"Transação ID: `%d`\n"+
				"Valor: R$ %.2f\n\n"+
				"Possível causa: Token de teste não gera QR codes reais.",
			payment.TransactionID, payment.Amount))
		return
	}

	qrCodeBytes, err := base64.StdEncoding.DecodeString(payment.QRCodeBase64)
	if err != nil {
		log.Printf("Erro ao decodificar QR code base64: %v", err)
		s.ChannelMessageSend(m.ChannelID, "❌ Erro ao processar imagem do QR Code.")
		return
	}

	message := fmt.Sprintf("💰 **Pagamento PIX criado!**\n\n"+
		"Valor: **R$ %.2f**\n"+
		"ID da transação: `%d`\n\n"+
		"📱 Escaneie o QR Code abaixo com seu app de pagamento:",
		payment.Amount, payment.TransactionID)

	s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content: message,
		Files: []*discordgo.File{
			{
				Name:        "qrcode.png",
				ContentType: "image/png",
				Reader:      bytes.NewReader(qrCodeBytes),
			},
		},
	})
}

func (b *Bot) handleBalanceCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	resp, err := http.Get(fmt.Sprintf("%s/api/wallet/balance?discord_id=%s", b.apiURL, m.Author.ID))
	if err != nil {
		log.Printf("Erro ao buscar saldo: %v", err)
		s.ChannelMessageSend(m.ChannelID, "❌ Erro ao buscar saldo.")
		return
	}
	defer resp.Body.Close()

	var result struct {
		Balance float64 `json:"balance"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	message := fmt.Sprintf("💰 **Seu saldo:** R$ %.2f", result.Balance)
	s.ChannelMessageSend(m.ChannelID, message)
}

func (b *Bot) handleTotalBalanceCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	resp, err := http.Get(fmt.Sprintf("%s/api/wallet/ranking?limit=1000", b.apiURL))
	if err != nil {
		log.Printf("Erro ao buscar saldo total: %v", err)
		s.ChannelMessageSend(m.ChannelID, "❌ Erro ao buscar saldo total.")
		return
	}
	defer resp.Body.Close()

	var ranking []struct {
		Username string  `json:"username"`
		Balance  float64 `json:"balance"`
	}
	json.NewDecoder(resp.Body).Decode(&ranking)

	total := 0.0
	for _, entry := range ranking {
		total += entry.Balance
	}

	message := fmt.Sprintf("💰 **Saldo total da vaquinha:** R$ %.2f", total)
	s.ChannelMessageSend(m.ChannelID, message)
}

func (b *Bot) handleRankingCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	resp, err := http.Get(fmt.Sprintf("%s/api/wallet/ranking?limit=10", b.apiURL))
	if err != nil {
		log.Printf("Erro ao buscar ranking: %v", err)
		s.ChannelMessageSend(m.ChannelID, "❌ Erro ao buscar ranking.")
		return
	}
	defer resp.Body.Close()

	var ranking []struct {
		Username string  `json:"username"`
		Balance  float64 `json:"balance"`
	}
	json.NewDecoder(resp.Body).Decode(&ranking)

	if len(ranking) == 0 {
		s.ChannelMessageSend(m.ChannelID, "📊 **Ranking vazio!** Ninguém contribuiu ainda.")
		return
	}

	message := "📊 **Top 10 Contribuidores:**\n\n"
	for i, entry := range ranking {
		medal := ""
		switch i {
		case 0:
			medal = "🥇"
		case 1:
			medal = "🥈"
		case 2:
			medal = "🥉"
		default:
			medal = fmt.Sprintf("%d.", i+1)
		}
		message += fmt.Sprintf("%s **%s** - R$ %.2f\n", medal, entry.Username, entry.Balance)
	}

	s.ChannelMessageSend(m.ChannelID, message)
}
