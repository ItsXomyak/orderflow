package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"orderflow/internal/domain/inventory"
)

type InventoryPG struct {
	pool *pgxpool.Pool
}

func NewInventoryPG(pool *pgxpool.Pool) *InventoryPG {
	return &InventoryPG{pool: pool}
}

func (r *InventoryPG) CreateProduct(ctx context.Context, product *inventory.Product) error {
	const q = `
		INSERT INTO products (id, name, sku, price, available, reserved, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, q,
		product.ID, product.Name, product.SKU, product.Price,
		product.Available, product.Reserved, product.CreatedAt, product.UpdatedAt,
	)
	return err
}

func (r *InventoryPG) GetProduct(ctx context.Context, productID string) (*inventory.Product, error) {
	const q = `
		SELECT id, name, sku, price, available, reserved, created_at, updated_at
		FROM products WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, q, productID)

	var product inventory.Product
	err := row.Scan(
		&product.ID, &product.Name, &product.SKU, &product.Price,
		&product.Available, &product.Reserved, &product.CreatedAt, &product.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, inventory.NewProductNotFoundError(productID)
	}
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *InventoryPG) UpdateProduct(ctx context.Context, product *inventory.Product) error {
	const q = `
		UPDATE products
		SET name = $2, sku = $3, price = $4, available = $5, reserved = $6, updated_at = $7
		WHERE id = $1
	`
	ct, err := r.pool.Exec(ctx, q,
		product.ID, product.Name, product.SKU, product.Price,
		product.Available, product.Reserved, product.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return inventory.NewProductNotFoundError(product.ID)
	}
	return nil
}

func (r *InventoryPG) GetProducts(ctx context.Context) ([]*inventory.Product, error) {
	const q = `
		SELECT id, name, sku, price, available, reserved, created_at, updated_at
		FROM products ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*inventory.Product
	for rows.Next() {
		var product inventory.Product
		err := rows.Scan(
			&product.ID, &product.Name, &product.SKU, &product.Price,
			&product.Available, &product.Reserved, &product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	return products, rows.Err()
}

func (r *InventoryPG) CreateReservation(ctx context.Context, reservation *inventory.Reservation) error {
	const q = `
		INSERT INTO reservations (id, order_id, product_id, quantity, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, q,
		reservation.ID, reservation.OrderID, reservation.ProductID,
		reservation.Quantity, reservation.ExpiresAt, reservation.CreatedAt,
	)
	return err
}

func (r *InventoryPG) GetReservationByOrderID(ctx context.Context, orderID string) (*inventory.Reservation, error) {
	const q = `
		SELECT id, order_id, product_id, quantity, expires_at, created_at
		FROM reservations WHERE order_id = $1
	`
	row := r.pool.QueryRow(ctx, q, orderID)

	var reservation inventory.Reservation
	err := row.Scan(
		&reservation.ID, &reservation.OrderID, &reservation.ProductID,
		&reservation.Quantity, &reservation.ExpiresAt, &reservation.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, inventory.NewReservationNotFoundError(orderID)
	}
	if err != nil {
		return nil, err
	}

	return &reservation, nil
}

func (r *InventoryPG) DeleteReservation(ctx context.Context, orderID string) error {
	const q = `DELETE FROM reservations WHERE order_id = $1`
	ct, err := r.pool.Exec(ctx, q, orderID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return inventory.NewReservationNotFoundError(orderID)
	}
	return nil
}

func (r *InventoryPG) GetExpiredReservations(ctx context.Context) ([]*inventory.Reservation, error) {
	const q = `
		SELECT id, order_id, product_id, quantity, expires_at, created_at
		FROM reservations WHERE expires_at < $1
	`
	rows, err := r.pool.Query(ctx, q, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservations []*inventory.Reservation
	for rows.Next() {
		var reservation inventory.Reservation
		err := rows.Scan(
			&reservation.ID, &reservation.OrderID, &reservation.ProductID,
			&reservation.Quantity, &reservation.ExpiresAt, &reservation.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, &reservation)
	}

	return reservations, rows.Err()
}
