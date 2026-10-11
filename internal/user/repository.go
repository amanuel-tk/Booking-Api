package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

var (
	ErrEmailTaken = errors.New("email already registered")
)

func (r *Repository) Create(ctx context.Context, u *User, roleName string) error {

	tx, err := r.db.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `INSERT INTO users (first_name, last_name, email, password_hash) VALUES ($1, $2, $3, $4) RETURNING id,is_active,created_at,updated_at`, u.FirstName, u.LastName, u.Email, u.PasswordHash).Scan(&u.ID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrEmailTaken
		}
		return err

	}

	tag, err := tx.Exec(ctx, `INSERT INTO user_roles (user_id,role_id) SELECT $1, id FROM roles WHERE name=$2`, u.ID, roleName)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return errors.New("role does not exist: " + roleName)
	}
	return tx.Commit(ctx)

}
