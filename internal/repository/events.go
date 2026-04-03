package repository

import (
	"context"
	"time"
)

type Event struct {
	ID          int
	Title       string
	Description string
	Department  string
	Gender      string
	Status      string
	CreatedBy   string

	TimeToSend time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (r *UserRepo) AddEvent(
	ctx context.Context,
	title, description, department, gender, createdBy string,
	timeToSend time.Time,
) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO events (title, description, department, gender, created_by, time_to_send)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, title, description, department, gender, createdBy, timeToSend)

	return err
}

func (r *UserRepo) GetEvents(ctx context.Context, createdBy string) ([]Event, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, title, description, department, gender, status, created_by, time_to_send, created_at
        FROM events
        WHERE created_by = $1
          AND time_to_send >= NOW()
        ORDER BY time_to_send ASC
    `, createdBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event

	for rows.Next() {
		var e Event
		err := rows.Scan(
			&e.ID,
			&e.Title,
			&e.Description,
			&e.Department,
			&e.Gender,
			&e.Status,
			&e.CreatedBy,
			&e.TimeToSend,
			&e.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *UserRepo) GetRecipients(ctx context.Context, department, gender string) ([]string, error) {
	rows, err := r.db.Query(ctx, `
        SELECT chat_id
        FROM users
        WHERE role = 'employee'
          AND department = $1
          AND gender = $2
    `, department, gender)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipients []string
	for rows.Next() {
		var chatID string
		if err := rows.Scan(&chatID); err != nil {
			return nil, err
		}
		recipients = append(recipients, chatID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return recipients, nil
}
