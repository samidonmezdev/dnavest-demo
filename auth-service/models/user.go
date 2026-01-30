package models

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Roles        []Role    `json:"roles,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}


type UserModel struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewUserModel(db *pgxpool.Pool, redis *redis.Client) *UserModel {
	return &UserModel{
		db:    db,
		redis: redis,
	}
}

// CreateUser creates a new user in the database and assigns default 'user' role
func (m *UserModel) CreateUser(ctx context.Context, email, passwordHash, name string) (*User, error) {
	var user User
	query := `
		INSERT INTO users (email, password_hash, name, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, name, created_at
	`

	err := m.db.QueryRow(ctx, query, email, passwordHash, name, time.Now()).
		Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	// Assign default 'user' role to new user
	roleQuery := `
		INSERT INTO user_roles (user_id, role_id)
		SELECT $1, id FROM roles WHERE name = 'user'
		ON CONFLICT DO NOTHING
	`
	_, err = m.db.Exec(ctx, roleQuery, user.ID)
	if err != nil {
		// Don't fail user creation if role assignment fails, just log it
		fmt.Printf("Warning: failed to assign default role to user %d: %v\n", user.ID, err)
	}

	return &user, nil
}


// GetUserByEmail retrieves a user by email
func (m *UserModel) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	query := `
		SELECT id, email, password_hash, name, created_at
		FROM users
		WHERE email = $1
	`

	err := m.db.QueryRow(ctx, query, email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (m *UserModel) GetUserByID(ctx context.Context, id int64) (*User, error) {
	var user User
	query := `
		SELECT id, email, password_hash, name, created_at
		FROM users
		WHERE id = $1
	`

	err := m.db.QueryRow(ctx, query, id).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// StoreRefreshToken stores a refresh token in Redis with 7 days TTL
func (m *UserModel) StoreRefreshToken(ctx context.Context, userID int64, token string) error {
	key := fmt.Sprintf("refresh_token:%d", userID)
	return m.redis.Set(ctx, key, token, 7*24*time.Hour).Err()
}

// ValidateRefreshToken checks if a refresh token is valid
func (m *UserModel) ValidateRefreshToken(ctx context.Context, userID int64, token string) bool {
	key := fmt.Sprintf("refresh_token:%d", userID)
	storedToken, err := m.redis.Get(ctx, key).Result()
	if err != nil {
		return false
	}
	return storedToken == token
}

// RevokeRefreshToken removes a refresh token from Redis
func (m *UserModel) RevokeRefreshToken(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("refresh_token:%d", userID)
	return m.redis.Del(ctx, key).Err()
}
