package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
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
	ErrNotFound   = errors.New("user not found")
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

const selectUser = `
	SELECT u.id, u.first_name,u.last_name, u.email,u.password_hash,u.is_active, u.created_at, u.updated_at,
	COALESCE(array_agg(DISTINCT ro.name) FILTER(WHERE ro.name IS NOT NULL),'{}'),
	COALESCE(array_agg(DISTINCT p.name) FILTER(WHERE p.name IS NOT NULL),'{}')
	FROM users u
	LEFT JOIN user_roles ur ON ur.user_id=u.id
	LEFT JOIN roles ro ON ro.id=ur.role_id
	LEFT JOIN role_permissions rp ON rp.role_id=ro.id
	LEFT JOIN permissions p ON p.id=rp.permission_id
`

func (r *Repository) scanOne(ctx context.Context, where string, arg any) (*User, error) {
	var u User

	err := r.db.QueryRow(ctx, selectUser+where+" GROUP BY u.id", arg).Scan(
		&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.PasswordHash, &u.IsActive, &u.CreatedAt, &u.UpdatedAt, &u.Roles, &u.Permissions,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {

	return r.scanOne(ctx, " WHERE u.email=$1", email)
}
