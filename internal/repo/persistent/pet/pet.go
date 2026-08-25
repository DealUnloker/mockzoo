// Package pet implements the Postgres-backed Pet repository.
package pet

import (
	"context"
	"errors"
	"fmt"

	"github.com/DealUnloker/mockzoo/internal/entity"
	"github.com/DealUnloker/mockzoo/internal/repo"
	"github.com/DealUnloker/mockzoo/pkg/postgres"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

// Repo -.
type Repo struct {
	*postgres.Postgres
}

// New returns a Pet repository.
func New(pg *postgres.Postgres) repo.PetRepo {
	return &Repo{pg}
}

func petColumns() []string {
	return []string{"id", "name", "status", "species", "breed", "photo_url", "tags", "created_at"}
}

func scanPet(row pgx.Row) (entity.Pet, error) {
	var p entity.Pet

	err := row.Scan(&p.ID, &p.Name, &p.Status, &p.Species, &p.Breed, &p.PhotoURL, &p.Tags, &p.CreatedAt)
	if err != nil {
		return entity.Pet{}, err
	}

	if p.Tags == nil {
		p.Tags = []string{}
	}

	return p, nil
}

// List -.
func (r *Repo) List(ctx context.Context, filter repo.PetFilter) ([]entity.Pet, int, error) {
	countBuilder := r.Builder.Select("COUNT(*)").From("pets")
	countBuilder = applyFilter(countBuilder, filter)

	countSQL, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("PetRepo - List - countBuilder: %w", err)
	}

	var total int

	err = r.Pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("PetRepo - List - count query: %w", err)
	}

	dataBuilder := r.Builder.
		Select(petColumns()...).
		From("pets").
		OrderBy("id ASC").
		Limit(filter.Limit).
		Offset(filter.Offset)
	dataBuilder = applyFilter(dataBuilder, filter)

	dataSQL, dataArgs, err := dataBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("PetRepo - List - dataBuilder: %w", err)
	}

	rows, err := r.Pool.Query(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("PetRepo - List - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	pets := make([]entity.Pet, 0, filter.Limit)

	for rows.Next() {
		p, scanErr := scanPet(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("PetRepo - List - rows.Scan: %w", scanErr)
		}

		pets = append(pets, p)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("PetRepo - List - rows: %w", err)
	}

	return pets, total, nil
}

func applyFilter(builder sq.SelectBuilder, filter repo.PetFilter) sq.SelectBuilder {
	if filter.Status != nil {
		builder = builder.Where(sq.Eq{"status": *filter.Status})
	}

	if filter.Species != nil {
		builder = builder.Where(sq.Eq{"species": *filter.Species})
	}

	return builder
}

// GetByID -.
func (r *Repo) GetByID(ctx context.Context, id int64) (entity.Pet, error) {
	sql, args, err := r.Builder.
		Select(petColumns()...).
		From("pets").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return entity.Pet{}, fmt.Errorf("PetRepo - GetByID - r.Builder: %w", err)
	}

	p, err := scanPet(r.Pool.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Pet{}, entity.ErrPetNotFound
		}

		return entity.Pet{}, fmt.Errorf("PetRepo - GetByID - r.Pool.QueryRow: %w", err)
	}

	return p, nil
}

// Create -.
func (r *Repo) Create(ctx context.Context, p *entity.Pet) error {
	sql, args, err := r.Builder.
		Insert("pets").
		Columns("name", "status", "species", "breed", "photo_url", "tags").
		Values(p.Name, p.Status, p.Species, p.Breed, p.PhotoURL, p.Tags).
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("PetRepo - Create - r.Builder: %w", err)
	}

	err = r.Pool.QueryRow(ctx, sql, args...).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return fmt.Errorf("PetRepo - Create - r.Pool.QueryRow: %w", err)
	}

	return nil
}

// Update -.
func (r *Repo) Update(ctx context.Context, p *entity.Pet) error {
	sql, args, err := r.Builder.
		Update("pets").
		Set("name", p.Name).
		Set("status", p.Status).
		Set("species", p.Species).
		Set("breed", p.Breed).
		Set("photo_url", p.PhotoURL).
		Set("tags", p.Tags).
		Where(sq.Eq{"id": p.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("PetRepo - Update - r.Builder: %w", err)
	}

	result, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("PetRepo - Update - r.Pool.Exec: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrPetNotFound
	}

	return nil
}

// Delete -.
func (r *Repo) Delete(ctx context.Context, id int64) error {
	sql, args, err := r.Builder.
		Delete("pets").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("PetRepo - Delete - r.Builder: %w", err)
	}

	result, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("PetRepo - Delete - r.Pool.Exec: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrPetNotFound
	}

	return nil
}

// Reset restores the sandbox: truncate the table, reinsert the canonical seed
// pets and reset the id sequence so the next created pet gets id 1001.
func (r *Repo) Reset(ctx context.Context) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("PetRepo - Reset - r.Pool.Begin: %w", err)
	}
	//nolint:errcheck // best-effort rollback: no-op once tx.Commit has already succeeded
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, "TRUNCATE pets RESTART IDENTITY")
	if err != nil {
		return fmt.Errorf("PetRepo - Reset - truncate: %w", err)
	}

	insertBuilder := r.Builder.
		Insert("pets").
		Columns("id", "name", "status", "species", "breed", "photo_url", "tags")

	for _, p := range seedPets() {
		insertBuilder = insertBuilder.Values(p.ID, p.Name, p.Status, p.Species, p.Breed, p.PhotoURL, p.Tags)
	}

	insertSQL, insertArgs, err := insertBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("PetRepo - Reset - insertBuilder: %w", err)
	}

	_, err = tx.Exec(ctx, insertSQL, insertArgs...)
	if err != nil {
		return fmt.Errorf("PetRepo - Reset - insert seed: %w", err)
	}

	_, err = tx.Exec(ctx, "SELECT setval('pets_id_seq', $1, true)", _seedIDStart)
	if err != nil {
		return fmt.Errorf("PetRepo - Reset - setval: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("PetRepo - Reset - tx.Commit: %w", err)
	}

	return nil
}
