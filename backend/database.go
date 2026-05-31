package main

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SensorReading struct {
	ID          int       `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Current     float64   `json:"current"`
	Voltage     float64   `json:"voltage"`
	Pressure    float64   `json:"pressure"`
	Temperature float64   `json:"temperature"`
	Vibration   float64   `json:"vibration"`
	IsProcessed int       `json:"is_processed"`
}

type Prediction struct {
	ID        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	RUL       int       `json:"rul"`
	Status    string    `json:"status"`
	Details   string    `json:"details"`
}

type Notification struct {
	ID        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	IsRead    int       `json:"is_read"`
}

var db *sql.DB

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "../sensor_data.db")
	if err != nil {
		return err
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS sensor_readings (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            timestamp TEXT NOT NULL,
            current REAL,
            voltage REAL,
            pressure REAL,
            temperature REAL,
            vibration REAL,
            is_processed INTEGER DEFAULT 0
        )`,
		`CREATE TABLE IF NOT EXISTS predictions (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            timestamp TEXT NOT NULL,
            rul INTEGER,
            status TEXT,
            details TEXT
        )`,
		`CREATE TABLE IF NOT EXISTS notifications (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            timestamp TEXT NOT NULL,
            level TEXT,
            message TEXT,
            is_read INTEGER DEFAULT 0
        )`,
	}

	for _, query := range queries {
		_, err = db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func getRecentReadings(minutes int) ([]SensorReading, error) {
	query := `
        SELECT id, timestamp, current, voltage, pressure, temperature, vibration, is_processed
        FROM sensor_readings 
        WHERE datetime(timestamp) >= datetime('now', '-' || ? || ' minutes')
        ORDER BY timestamp ASC
        LIMIT 500
    `
	rows, err := db.Query(query, minutes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []SensorReading
	for rows.Next() {
		var r SensorReading
		var ts string
		err := rows.Scan(&r.ID, &ts, &r.Current, &r.Voltage, &r.Pressure, &r.Temperature, &r.Vibration, &r.IsProcessed)
		if err != nil {
			return nil, err
		}
		r.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
		readings = append(readings, r)
	}
	return readings, nil
}

func getLatestPrediction() (*Prediction, error) {
	query := `
        SELECT id, timestamp, rul, status, details
        FROM predictions 
        ORDER BY timestamp DESC 
        LIMIT 1
    `
	row := db.QueryRow(query)
	var p Prediction
	var ts string
	err := row.Scan(&p.ID, &ts, &p.RUL, &p.Status, &p.Details)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	p.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
	return &p, nil
}

func getPredictionsHistory(limit int) ([]Prediction, error) {
	query := `
        SELECT id, timestamp, rul, status, details
        FROM predictions 
        ORDER BY timestamp DESC 
        LIMIT ?
    `
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var predictions []Prediction
	for rows.Next() {
		var p Prediction
		var ts string
		err := rows.Scan(&p.ID, &ts, &p.RUL, &p.Status, &p.Details)
		if err != nil {
			return nil, err
		}
		p.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
		predictions = append(predictions, p)
	}
	return predictions, nil
}

func getUnreadNotifications() ([]Notification, error) {
	query := `
        SELECT id, timestamp, level, message, is_read
        FROM notifications 
        WHERE is_read = 0
        ORDER BY timestamp DESC
    `
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		var ts string
		err := rows.Scan(&n.ID, &ts, &n.Level, &n.Message, &n.IsRead)
		if err != nil {
			return nil, err
		}
		n.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func markNotificationAsRead(id int) error {
	_, err := db.Exec("UPDATE notifications SET is_read = 1 WHERE id = ?", id)
	return err
}

func addNotification(level, message string) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := db.Exec(
		"INSERT INTO notifications (timestamp, level, message, is_read) VALUES (?, ?, ?, 0)",
		now, level, message,
	)
	return err
}

func savePrediction(rul int, status, details string) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := db.Exec(
		"INSERT INTO predictions (timestamp, rul, status, details) VALUES (?, ?, ?, ?)",
		now, rul, status, details,
	)
	return err
}
