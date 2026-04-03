package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

//type User struct {
//	ID         int
//	ChatID     string
//	Name       string
//	Role       string
//	Gender     string
//	Department string
//	LastSeen   string
//	CreatedAt  time.Time
//}

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) AddUser(ctx context.Context, chatID, name string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (chat_id, name, last_seen)
		VALUES ($1, $2, NOW())
		ON CONFLICT (chat_id) DO UPDATE
		SET name      = EXCLUDED.name,    
		last_seen = NOW()
                     
	`, chatID, name)

	return err
}

func (r *UserRepo) UpdateRole(ctx context.Context, role, chatID string) error {
	_, err := r.db.Exec(ctx, `
     	UPDATE users SET role=$1 WHERE chat_id=$2
	`, role, chatID)

	return err
}

func (r *UserRepo) AddDept(ctx context.Context, dept, chatID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET department=$1 WHERE chat_id=$2
	`, dept, chatID)

	return err
}

func (r *UserRepo) AddGender(ctx context.Context, gender, chatID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET gender=$1 WHERE chat_id=$2
	`, gender, chatID)
	return err
}

func (r *UserRepo) GetRole(ctx context.Context, chatID string) (string, error) {
	var role string
	err := r.db.QueryRow(ctx,
		`SELECT role FROM users WHERE chat_id = $1`, chatID).Scan(&role)

	return role, err
}
