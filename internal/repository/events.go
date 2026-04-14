package repository

import (
	"context"
	"time"
)

type Notification struct {
	ID                   int
	Title                string
	Description          string
	RecipientsDepartment string
	RecipientsGender     string
	Status               string
	CreatedBy            string
	CreatedAt            time.Time
	UpdatedAt            time.Time

	Type          string
	TimeToSend    *time.Time
	TriggerEvent  *string
	TriggerUserID *string
	Recurrence    *string
	NextSend      *time.Time
	LastSent      *time.Time
}

type Recipient struct {
	ChatID string
	Name   string
	Gender string
}

type NotificationStats struct {
	ID             int       `json:"id"`
	Title          string    `json:"title"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	ReadCount      int       `json:"read_count"`
	DeliveredCount int       `json:"delivered_count"`
	OpenRate       float64   `json:"open_rate"`
}

func (r *UserRepo) AddNotification(ctx context.Context, title, description, recipientsDepartment, recipientsGender, createdBy string, timeToSend time.Time) (int, error) {
	var notificationID int
	err := r.db.QueryRow(ctx, `
        INSERT INTO notifications (
            title, description,
            recipients_department, recipients_gender,
            status, created_by, created_at,
            type, time_to_send
        )
        VALUES ($1, $2, $3, $4, 'scheduled', $5, NOW(), 'one_time', $6)
        RETURNING id
    `, title, description, recipientsDepartment, recipientsGender, createdBy, timeToSend).Scan(&notificationID)

	return notificationID, err
}

func (r *UserRepo) AddRecurringNotification(ctx context.Context, title, description, recipientsDepartment, recipientsGender, recurrence, createdBy string, nextSend time.Time) (int, error) {
	var notificationID int
	err := r.db.QueryRow(ctx, `
        INSERT INTO notifications (
            title, description,
            recipients_department, recipients_gender,
            status, created_by, created_at,
            type, recurrence, next_send
        )
        VALUES ($1, $2, $3, $4, 'scheduled', $5, NOW(), 'recurring', $6, $7)
        RETURNING id
    `, title, description, recipientsDepartment, recipientsGender, createdBy, recurrence, nextSend).Scan(&notificationID)

	return notificationID, err
}

func (r *UserRepo) GetMyNotifications(ctx context.Context, createdBy string) ([]Notification, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, title, description,
               COALESCE(recipients_department, ''),
               COALESCE(recipients_gender, ''),
               status, created_by, created_at, updated_at, type, time_to_send,
               trigger_event, trigger_user_id, recurrence, next_send, last_sent
        FROM notifications
        WHERE created_by = $1
		ORDER BY created_at DESC
        
    `, createdBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		err := rows.Scan(
			&n.ID, &n.Title, &n.Description, &n.RecipientsDepartment, &n.RecipientsGender,
			&n.Status, &n.CreatedBy, &n.CreatedAt, &n.UpdatedAt, &n.Type, &n.TimeToSend,
			&n.TriggerEvent, &n.TriggerUserID, &n.Recurrence, &n.NextSend, &n.LastSent,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}

	return notifications, rows.Err()
}

func (r *UserRepo) GetRecipientsWithInfo(ctx context.Context, department, gender string) ([]Recipient, error) {
	rows, err := r.db.Query(ctx, `
        SELECT chat_id, name, COALESCE(gender, '')
        FROM users
        WHERE role = 'employee'
          AND ($1 = 'all' OR department = $1)
          AND ($2 = 'all' OR gender = $2)
    `, department, gender)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipients []Recipient
	for rows.Next() {
		var rec Recipient
		if err := rows.Scan(&rec.ChatID, &rec.Name, &rec.Gender); err != nil {
			return nil, err
		}
		recipients = append(recipients, rec)
	}

	return recipients, rows.Err()
}

func (r *UserRepo) UpdateNotificationStatus(ctx context.Context, notificationID int, status string) error {
	_, err := r.db.Exec(ctx, `
        UPDATE notifications
        SET status = $1, updated_at = NOW()
        WHERE id = $2
    `, status, notificationID)
	return err
}

func (r *UserRepo) GetReadyNotifications(ctx context.Context) ([]Notification, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, title, description,
               COALESCE(recipients_department, ''),
               COALESCE(recipients_gender, ''),
               status, created_by, created_at, updated_at, type, time_to_send,
               trigger_event, trigger_user_id, recurrence, next_send, last_sent
        FROM notifications
        WHERE status = 'scheduled'
          AND (
            (type = 'one_time' AND time_to_send <= NOW()) OR
            (type = 'recurring' AND next_send <= NOW())
          )
        ORDER BY
            CASE
                WHEN type = 'one_time' THEN time_to_send
                WHEN type = 'recurring' THEN next_send
            END ASC
        LIMIT 10
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		err := rows.Scan(
			&n.ID, &n.Title, &n.Description, &n.RecipientsDepartment, &n.RecipientsGender,
			&n.Status, &n.CreatedBy, &n.CreatedAt, &n.UpdatedAt, &n.Type, &n.TimeToSend,
			&n.TriggerEvent, &n.TriggerUserID, &n.Recurrence, &n.NextSend, &n.LastSent,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}

	return notifications, rows.Err()
}

func (r *UserRepo) GetNotificationByID(ctx context.Context, notificationID int) (Notification, error) {
	var n Notification
	err := r.db.QueryRow(ctx, `
        SELECT id, title, description,
               COALESCE(recipients_department, ''),
               COALESCE(recipients_gender, ''),
               status, created_by, created_at, updated_at, type, time_to_send,
               trigger_event, trigger_user_id, recurrence, next_send, last_sent
        FROM notifications
        WHERE id = $1
    `, notificationID).Scan(
		&n.ID, &n.Title, &n.Description, &n.RecipientsDepartment, &n.RecipientsGender,
		&n.Status, &n.CreatedBy, &n.CreatedAt, &n.UpdatedAt, &n.Type, &n.TimeToSend,
		&n.TriggerEvent, &n.TriggerUserID, &n.Recurrence, &n.NextSend, &n.LastSent,
	)
	if err != nil {
		return Notification{}, err
	}
	return n, nil
}

func (r *UserRepo) MarkRecurringNotificationSent(ctx context.Context, notification Notification) error {
	if notification.Recurrence == nil || notification.NextSend == nil {
		return r.UpdateNotificationStatus(ctx, notification.ID, "failed")
	}

	currentNextSend := *notification.NextSend
	var newNextSend time.Time

	switch *notification.Recurrence {
	case "weekly":
		newNextSend = currentNextSend.AddDate(0, 0, 7)

	case "monthly":
		newNextSend = currentNextSend.AddDate(0, 1, 0)

	case "workdays":
		newNextSend = currentNextSend.AddDate(0, 0, 1)
		for newNextSend.Weekday() == time.Saturday || newNextSend.Weekday() == time.Sunday {
			newNextSend = newNextSend.AddDate(0, 0, 1)
		}

	default:
		return r.UpdateNotificationStatus(ctx, notification.ID, "failed")
	}

	_, err := r.db.Exec(ctx, `
        UPDATE notifications
        SET last_sent = NOW(),
            next_send = $1,
            status = 'scheduled',
            updated_at = NOW()
        WHERE id = $2
    `, newNextSend, notification.ID)

	return err
}

func (r *UserRepo) GetWelcomeTrigger(ctx context.Context) (Notification, error) {
	var n Notification
	err := r.db.QueryRow(ctx, `
        SELECT id, title, description,
               COALESCE(recipients_department, ''),
               COALESCE(recipients_gender, ''),
               status, created_by, created_at, updated_at, type, time_to_send,
               trigger_event, trigger_user_id, recurrence, next_send, last_sent
        FROM notifications
        WHERE type = 'event_trigger' AND trigger_event = 'new_employee'
        LIMIT 1
    `).Scan(
		&n.ID, &n.Title, &n.Description, &n.RecipientsDepartment, &n.RecipientsGender,
		&n.Status, &n.CreatedBy, &n.CreatedAt, &n.UpdatedAt, &n.Type, &n.TimeToSend,
		&n.TriggerEvent, &n.TriggerUserID, &n.Recurrence, &n.NextSend, &n.LastSent,
	)
	if err != nil {
		return Notification{}, err
	}
	return n, nil
}

func (r *UserRepo) EnsureWelcomeTrigger(ctx context.Context) error {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM notifications
			WHERE type = 'event_trigger'
			  AND trigger_event = 'new_employee'
		)
	`).Scan(&exists)

	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// создаём триггерное уведомление
	_, err = r.db.Exec(ctx, `
		INSERT INTO notifications (
			title,
			description,
			recipients_department,
			recipients_gender,
			status,
			created_at,
			type,
			trigger_event
		)
		VALUES (
			$1, $2,
			'all',
			'all',
			'scheduled',
			NOW(),
			'event_trigger',
			'new_employee'
		)
	`,
		"Добро пожаловать, {name}!",
		"Профиль заполнен! 🎉\n\nТеперь вы будете получать уведомления о корпоративных событиях вашего отдела.\nКак только появится новое мероприятие — я сразу оповещу вас! 🔔\nЕсли появятся вопросы — введите /info",
	)

	return err
}

func (r *UserRepo) MarkTriggerNotificationSent(ctx context.Context, notificationID int) error {
	_, err := r.db.Exec(ctx, `
        UPDATE notifications
        SET last_sent = NOW(),
            updated_at = NOW()
        WHERE id = $1
    `, notificationID)
	return err
}

func (r *UserRepo) GetNotificationStats(ctx context.Context) ([]NotificationStats, error) {
	rows, err := r.db.Query(ctx, `
		WITH recipient_counts AS (
			SELECT 
				n.id AS notification_id,
				COUNT(u.chat_id) AS delivered_count
			FROM notifications n
			CROSS JOIN users u
			WHERE u.role = 'employee'
				AND (COALESCE(n.recipients_department, 'all') = 'all' OR u.department = n.recipients_department)
				AND (COALESCE(n.recipients_gender, 'all') = 'all' OR u.gender = n.recipients_gender)
			GROUP BY n.id
		),
		read_counts AS (
			-- Считаем прочтения из message_reads
			SELECT notification_id, COUNT(*) AS read_count
			FROM message_reads
			GROUP BY notification_id
		)
		SELECT
			n.id,
			n.title,
			n.status,
			n.created_at,
			COALESCE(rc.read_count, 0) AS read_count,
			COALESCE(rec.delivered_count, 0) AS delivered_count,
			CASE
				WHEN COALESCE(rec.delivered_count, 0) = 0 THEN 0
				ELSE ROUND(COALESCE(rc.read_count, 0)::decimal / rec.delivered_count * 100, 1)
			END AS open_rate
		FROM notifications n
		LEFT JOIN read_counts rc ON rc.notification_id = n.id
		LEFT JOIN recipient_counts rec ON rec.notification_id = n.id
		ORDER BY n.created_at DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []NotificationStats
	for rows.Next() {
		var s NotificationStats
		if err := rows.Scan(
			&s.ID,
			&s.Title,
			&s.Status,
			&s.CreatedAt,
			&s.ReadCount,
			&s.DeliveredCount,
			&s.OpenRate,
		); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, rows.Err()
}

func (r *UserRepo) MarkAsRead(ctx context.Context, notificationID int, chatID string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO message_reads (notification_id, chat_id)
		VALUES ($1, $2)
		ON CONFLICT (notification_id, chat_id) DO NOTHING
	`, notificationID, chatID)

	return err
}
