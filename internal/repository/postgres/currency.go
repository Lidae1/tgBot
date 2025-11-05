package postgres

import (
	"context"
	"database/sql"
	"log/slog"
	"tgBotFinal/internal/domain/service"
	"time"

	"tgBotFinal/internal/entity"
)

type CurrencyRepo struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewCurrencyRepo(db *sql.DB, logger *slog.Logger) service.CurrencyRepository {
	return &CurrencyRepo{db: db, logger: logger.With(slog.String("component", "CurrencyRepo"))}

}

func (cr *CurrencyRepo) SaveOrUpdate(ctx context.Context, currency *entity.Price) error {
	cr.logger.Debug("Saving currency", "symbol", currency.Symbol)

	query := `
		INSERT INTO currencies (symbol, price, updated)
		VALUES ($1, $2, $3)
		ON CONFLICT (symbol)
		DO UPDATE SET price=$2, updated = $3
		`

	_, err := cr.db.ExecContext(ctx, query, currency.Symbol, currency.Price, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		cr.logger.Error("failed to save currency", "symbol", currency.Symbol, "err", err)
	} else {
		cr.logger.Debug("saved currency successfully", "symbol", currency.Symbol)
	}

	return err
}

func (cr *CurrencyRepo) GetBySymbol(ctx context.Context, symbol entity.CurrencyName) (*entity.Price, error) {
	cr.logger.Debug("Getting currency by symbol", "symbol", symbol)

	query := `
		SELECT symbol, price, updated FROM currencies
		WHERE symbol=$1;
`

	var currency entity.Price
	err := cr.db.QueryRowContext(ctx, query, symbol).Scan(
		&currency.Symbol, &currency.Price, &currency.Updated)

	if err == sql.ErrNoRows {
		cr.logger.Debug("no currency by symbol", "symbol", symbol)
		return nil, nil
	}

	if err != nil {
		cr.logger.Error("failed to get currency by symbol", "symbol", symbol, "err", err)
	} else {
		cr.logger.Debug("got currency by symbol", "symbol", symbol)
	}

	return &currency, nil

}

func (cr *CurrencyRepo) GetAll(ctx context.Context) ([]*entity.Price, error) {
	cr.logger.Debug("Getting all currencies")

	query := `
		SELECT symbol, price, updated FROM currencies
		`

	rows, err := cr.db.QueryContext(ctx, query)
	if err != nil {
		cr.logger.Error("failed to get all currencies", "err", err)
		return nil, err
	}
	defer rows.Close()

	var currencies []*entity.Price
	for rows.Next() {
		var currency entity.Price
		if err := rows.Scan(&currency.Symbol, &currency.Price, &currency.Updated); err != nil {
			cr.logger.Error("failed to get all currencies", "err", err)
			return nil, err
		}

		currencies = append(currencies, &currency)
	}

	cr.logger.Debug("got all currencies")
	return currencies, nil
}
