package postgres

import (
	"context"
	"database/sql"
	"log/slog"
	"tgBotFinal/internal/domain/service"

	"tgBotFinal/internal/entity"
)

type UserRepo struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewUserRepo(db *sql.DB, logger *slog.Logger) service.UserRepository {
	return &UserRepo{
		db:     db,
		logger: logger.With(slog.String("components", "UserRepo")),
	}
}

func (ur *UserRepo) SaveOrUpdate(ctx context.Context, user *entity.User) error {
	ur.logger.Debug("save user", "user", user.ChatID)

	query := `
		INSERT INTO users VALUES ($1, $2, $3)
		ON CONFLICT(ChatID)
		DO UPDATE SET username = $2, active = $3;
`
	_, err := ur.db.ExecContext(ctx, query, user.ChatID, user.Username, user.Active)
	if err != nil {
		ur.logger.Error("error save user", "err", err)
	} else {
		ur.logger.Info("save user", "user", user.ChatID)
	}

	return err
}

func (ur *UserRepo) GetByChatID(ctx context.Context, chatID int64) (*entity.User, error) {
	ur.logger.Debug("get user by chatID", "chatID", chatID)

	query := `SELECT * FROM users WHERE chat_id = $1;`

	var user entity.User
	err := ur.db.QueryRowContext(ctx, query, chatID).Scan(
		&user.ChatID, &user.Username, &user.Active)
	if err == sql.ErrNoRows {
		ur.logger.Debug("no user found", "err", err)
		return nil, nil
	}

	if err != nil {
		ur.logger.Error("error getting user", "err", err)
	} else {
		ur.logger.Debug("got user", "user", user.ChatID)
	}

	return &user, err
}

func (ur *UserRepo) GetAll(ctx context.Context) ([]*entity.User, error) {
	ur.logger.Debug("get all users")

	query := `SELECT * FROM users;`

	var users []*entity.User
	rows, err := ur.db.QueryContext(ctx, query)
	if err != nil {
		ur.logger.Error("error getting all users", "err", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.ChatID, &user.Username, &user.Active); err != nil {
			ur.logger.Error("error getting all users", "err", err)
			return nil, err
		}

		users = append(users, &user)
	}

	ur.logger.Debug("got all users")
	return users, err
}

func (ur *UserRepo) GetAllActive(ctx context.Context) ([]*entity.User, error) {
	ur.logger.Debug("get all active users")
	query := `SELECT * FROM users WHERE active=true;`

	var users []*entity.User
	rows, err := ur.db.QueryContext(ctx, query)
	if err != nil {
		ur.logger.Error("error getting all active users", "err", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.ChatID, &user.Username, &user.Active); err != nil {
			ur.logger.Error("error getting all active users", "err", err)
			return nil, err
		}

		users = append(users, &user)
	}

	ur.logger.Debug("got all active users")
	return users, err
}
