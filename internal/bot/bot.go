package bot

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session *discordgo.Session
	db      *sql.DB
}

func New(token string, db *sql.DB) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar sessão do Discord: %w", err)
	}

	bot := &Bot{
		session: session,
		db:      db,
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

	if m.Content == "!ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

}
