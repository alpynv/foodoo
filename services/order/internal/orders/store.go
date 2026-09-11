package orders

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("order not found")

type Order struct {
	ID               uuid.UUID `json:"id"`
	CustomerID       uuid.UUID `json:"customer_id"`
	Status           string    `json:"status"`
	TotalAmountCents int64     `json:"total_amount_cents"`
	Currency         string    `json:"currency"`
	CreatedAt        time.Time `json:"created_at"`
}

type CreateInput struct {
	CustomerID       uuid.UUID
	TotalAmountCents int64
	Currency         string
	IdempotencyKey   string
	CorrelationID    string
}

type Store interface {
	Create(context.Context, CreateInput) (Order, error)
	Get(context.Context, uuid.UUID) (Order, error)
	Ping(context.Context) error
}

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *PostgresStore) Create(ctx context.Context, input CreateInput) (Order, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Order{}, err
	}
	defer tx.Rollback(ctx) // No-op after a successful commit.

	order := Order{ID: uuid.New(), CustomerID: input.CustomerID, Status: "pending", TotalAmountCents: input.TotalAmountCents, Currency: input.Currency}
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (id, customer_id, status, total_amount_cents, currency, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (customer_id, idempotency_key) DO NOTHING
		RETURNING created_at`,
		order.ID, order.CustomerID, order.Status, order.TotalAmountCents, order.Currency, input.IdempotencyKey,
	).Scan(&order.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return getByIdempotencyKey(ctx, tx, input.CustomerID, input.IdempotencyKey)
	}
	if err != nil {
		return Order{}, err
	}

	eventID := uuid.New()
	payload, err := json.Marshal(map[string]any{
		"event_id":       eventID,
		"schema_version": 1,
		"order_id":       order.ID,
		"correlation_id": input.CorrelationID,
		"occurred_at":    order.CreatedAt,
		"data": map[string]any{
			"customer_id":        order.CustomerID,
			"status":             order.Status,
			"total_amount_cents": order.TotalAmountCents,
			"currency":           order.Currency,
		},
	})
	if err != nil {
		return Order{}, err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO outbox_events (id, aggregate_id, event_type, schema_version, payload)
		VALUES ($1, $2, 'OrderCreated', 1, $3)`, eventID, order.ID, payload)
	if err != nil {
		return Order{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Order{}, err
	}
	return order, nil
}

func (s *PostgresStore) Get(ctx context.Context, id uuid.UUID) (Order, error) {
	return get(ctx, s.pool, `SELECT id, customer_id, status, total_amount_cents, currency, created_at FROM orders WHERE id = $1`, id)
}

type queryRower interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func getByIdempotencyKey(ctx context.Context, db queryRower, customerID uuid.UUID, key string) (Order, error) {
	return get(ctx, db, `SELECT id, customer_id, status, total_amount_cents, currency, created_at FROM orders WHERE customer_id = $1 AND idempotency_key = $2`, customerID, key)
}

func get(ctx context.Context, db queryRower, query string, args ...any) (Order, error) {
	var order Order
	err := db.QueryRow(ctx, query, args...).Scan(
		&order.ID, &order.CustomerID, &order.Status, &order.TotalAmountCents, &order.Currency, &order.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrNotFound
	}
	return order, err
}
