package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/sinhaparth5/whatstyle-mcp/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func InitDB(dbPath string) (*DB, error) {
	log.Printf("Initializing database at: %s", dbPath)
	
	conn, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	log.Printf("Database initialized successfully")
	return db, nil
}

func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

func (db *DB) createTables() error {
	createMessagesTable := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		content TEXT NOT NULL,
		role TEXT NOT NULL CHECK(role IN ('user', 'assistant')),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_messages_user_id ON messages(user_id);
	CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
	CREATE INDEX IF NOT EXISTS idx_messages_user_time ON messages(user_id, created_at);
	`

	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT UNIQUE NOT NULL,
		phone_number TEXT,
		name TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_seen DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_users_user_id ON users(user_id);
	CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone_number);
	`

	createSessionsTable := `
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		session_data TEXT,
		expires_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(user_id)
	);
	
	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);
	`

	tables := []string{createMessagesTable, createUsersTable, createSessionsTable}
	
	for _, table := range tables {
		if _, err := db.conn.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

func (db *DB) SaveMessage(userID, content, role string) error {
	if userID == "" || content == "" || role == "" {
		return fmt.Errorf("userID, content, and role are required")
	}

	if role != "user" && role != "assistant" {
		return fmt.Errorf("role must be 'user' or 'assistant'")
	}

	query := `INSERT INTO messages (user_id, content, role) VALUES (?, ?, ?)`
	_, err := db.conn.Exec(query, userID, content, role)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	log.Printf("Saved %s message for user %s", role, userID)
	return nil
}

func (db *DB) GetChatHistory(userID string, limit int) ([]models.Message, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT id, user_id, content, role, created_at 
		FROM messages 
		WHERE user_id = ? 
		ORDER BY created_at DESC 
		LIMIT ?
	`

	rows, err := db.conn.Query(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(&msg.ID, &msg.UserID, &msg.Content, &msg.Role, &msg.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Reverse to get chronological order (oldest first)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (db *DB) GetUserMessageCount(userID string) (int, error) {
	if userID == "" {
		return 0, fmt.Errorf("userID is required")
	}

	query := `SELECT COUNT(*) FROM messages WHERE user_id = ?`
	var count int
	err := db.conn.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}

	return count, nil
}

func (db *DB) CreateOrUpdateUser(userID, phoneNumber, name string) error {
	if userID == "" {
		return fmt.Errorf("userID is required")
	}

	query := `
		INSERT INTO users (user_id, phone_number, name, last_seen) 
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id) DO UPDATE SET 
			phone_number = COALESCE(?, phone_number),
			name = COALESCE(?, name),
			last_seen = CURRENT_TIMESTAMP
	`
	
	_, err := db.conn.Exec(query, userID, phoneNumber, name, phoneNumber, name)
	if err != nil {
		return fmt.Errorf("failed to create/update user: %w", err)
	}

	return nil
}

func (db *DB) GetUser(userID string) (*models.User, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	query := `SELECT user_id, phone_number, name, created_at, last_seen FROM users WHERE user_id = ?`
	
	var user models.User
	var phoneNumber, name sql.NullString
	
	err := db.conn.QueryRow(query, userID).Scan(
		&user.UserID, &phoneNumber, &name, &user.CreatedAt, &user.LastSeen,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if phoneNumber.Valid {
		user.PhoneNumber = phoneNumber.String
	}
	if name.Valid {
		user.Name = name.String
	}

	return &user, nil
}

func (db *DB) GetRecentUsers(limit int) ([]models.User, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT user_id, phone_number, name, created_at, last_seen 
		FROM users 
		ORDER BY last_seen DESC 
		LIMIT ?
	`

	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var phoneNumber, name sql.NullString
		
		err := rows.Scan(&user.UserID, &phoneNumber, &name, &user.CreatedAt, &user.LastSeen)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		if phoneNumber.Valid {
			user.PhoneNumber = phoneNumber.String
		}
		if name.Valid {
			user.Name = name.String
		}

		users = append(users, user)
	}

	return users, nil
}

func (db *DB) SaveSession(userID, sessionData string, expiresAt time.Time) error {
	if userID == "" {
		return fmt.Errorf("userID is required")
	}

	query := `
		INSERT INTO sessions (user_id, session_data, expires_at) 
		VALUES (?, ?, ?)
	`
	
	_, err := db.conn.Exec(query, userID, sessionData, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

func (db *DB) GetSession(userID string) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("userID is required")
	}

	query := `
		SELECT session_data 
		FROM sessions 
		WHERE user_id = ? AND expires_at > CURRENT_TIMESTAMP 
		ORDER BY created_at DESC 
		LIMIT 1
	`
	
	var sessionData sql.NullString
	err := db.conn.QueryRow(query, userID).Scan(&sessionData)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // No active session
		}
		return "", fmt.Errorf("failed to get session: %w", err)
	}

	if sessionData.Valid {
		return sessionData.String, nil
	}
	
	return "", nil
}

func (db *DB) CleanupExpiredSessions() error {
	query := `DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP`
	
	result, err := db.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to cleanup sessions: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected > 0 {
		log.Printf("Cleaned up %d expired sessions", affected)
	}

	return nil
}

func (db *DB) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total messages
	var totalMessages int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM messages").Scan(&totalMessages)
	if err != nil {
		return nil, fmt.Errorf("failed to count messages: %w", err)
	}
	stats["total_messages"] = totalMessages

	// Total users
	var totalUsers int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}
	stats["total_users"] = totalUsers

	// Active sessions
	var activeSessions int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM sessions WHERE expires_at > CURRENT_TIMESTAMP").Scan(&activeSessions)
	if err != nil {
		return nil, fmt.Errorf("failed to count active sessions: %w", err)
	}
	stats["active_sessions"] = activeSessions

	// Messages today
	var messagesToday int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM messages WHERE DATE(created_at) = DATE('now')").Scan(&messagesToday)
	if err != nil {
		return nil, fmt.Errorf("failed to count today's messages: %w", err)
	}
	stats["messages_today"] = messagesToday

	return stats, nil
}
