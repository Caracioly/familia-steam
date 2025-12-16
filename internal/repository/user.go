package repository

import (
	"database/sql"
	"fmt"
	"time"
)

type User struct {
	ID        int64
	DiscordID string
	Username  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(userID int64) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(`
		SELECT id, discord_id, username, created_at, updated_at
		FROM users
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.DiscordID, &user.Username, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}

	return user, nil
}

func (r *UserRepository) FindByDiscordID(discordID string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(`
		SELECT id, discord_id, username, created_at, updated_at
		FROM users
		WHERE discord_id = $1
	`, discordID).Scan(&user.ID, &user.DiscordID, &user.Username, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}

	return user, nil
}

func (r *UserRepository) Create(discordID, username string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(`
		INSERT INTO users (discord_id, username)
		VALUES ($1, $2)
		RETURNING id, discord_id, username, created_at, updated_at
	`, discordID, username).Scan(&user.ID, &user.DiscordID, &user.Username, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("erro ao criar usuário: %w", err)
	}

	return user, nil
}

func (r *UserRepository) FindOrCreate(discordID, username string) (*User, error) {
	user, err := r.FindByDiscordID(discordID)
	if err != nil {
		return nil, err
	}

	if user != nil {
		if user.Username != username {
			_, err = r.db.Exec(`UPDATE users SET username = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, username, user.ID)
			if err != nil {
				return nil, fmt.Errorf("erro ao atualizar username: %w", err)
			}
			user.Username = username
		}
		return user, nil
	}

	return r.Create(discordID, username)
}
