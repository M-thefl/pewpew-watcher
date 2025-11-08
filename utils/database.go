package utils

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Config struct {
	DiscordWebhook string              `json:"DiscordWebhook"`
	Telegram       TelegramConfig      `json:"telegram"`
	Database       DatabaseConfig      `json:"database"`
	Platforms      map[string]Platform `json:"platforms"`
}

type TelegramConfig struct {
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
}

type DatabaseConfig struct {
	Path string `json:"path"`
}

type Platform struct {
	URL           string        `json:"url"`
	Monitor       bool          `json:"monitor"`
	Notifications Notifications `json:"notifications"`
}

type Notifications struct {
	NewProgram     bool `json:"new_program"`
	RemovedProgram bool `json:"removed_program"`
	NewScope       bool `json:"new_scope"`
	RemovedScope   bool `json:"removed_scope"`
	ChangedScope   bool `json:"changed_scope"`
	NewType        bool `json:"new_type"`
}

type Program struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	Type       string `json:"type"`
	Key        string `json:"key"`
	Platform   string `json:"platform"`
	Logo       string `json:"logo"`
	Scope      string `json:"scope"`
	InScope    string `json:"in_scope"`
	OutOfScope string `json:"out_of_scope"`
	Reward     string `json:"reward"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type Reward struct {
	Min string `json:"min"`
	Max string `json:"max"`
}

type Alert struct {
	Program      *Program
	IsNew        bool
	IsRemoved    bool
	NewScope     []string
	RemovedScope []string
	ChangedScope []ScopeChange
	NewType      string
	Reward       *Reward
}

type ScopeChange struct {
	Old string `json:"old"`
	New string `json:"new"`
}

func InitDatabase(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS programs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		type TEXT NOT NULL,
		key TEXT UNIQUE NOT NULL,
		platform TEXT NOT NULL,
		logo TEXT,
		scope TEXT DEFAULT '{}',
		in_scope TEXT DEFAULT '[]',
		out_of_scope TEXT DEFAULT '[]',
		reward TEXT DEFAULT '{}',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_programs_key ON programs(key);
	CREATE INDEX IF NOT EXISTS idx_programs_platform ON programs(platform);
	`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("  Database initialization error:", err)
	}
	log.Println("💾 Database initialized successfully")
}

func IsFirstRun(db *sql.DB) bool {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM programs").Scan(&count)
	if err != nil {
		return true
	}
	return count == 0
}

func SaveProgram(db *sql.DB, program *Program) error {
	query := `
	INSERT OR REPLACE INTO programs 
	(name, url, type, key, platform, logo, scope, in_scope, out_of_scope, reward, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err := db.Exec(query,
		program.Name, program.URL, program.Type, program.Key, program.Platform,
		program.Logo, program.Scope, program.InScope, program.OutOfScope, program.Reward,
	)
	return err
}

func GetProgram(db *sql.DB, key string) (*Program, error) {
	program := &Program{}
	err := db.QueryRow(`
		SELECT id, name, url, type, key, platform, logo, scope, in_scope, out_of_scope, reward, created_at, updated_at
		FROM programs WHERE key = ?
	`, key).Scan(
		&program.ID, &program.Name, &program.URL, &program.Type, &program.Key, &program.Platform,
		&program.Logo, &program.Scope, &program.InScope, &program.OutOfScope, &program.Reward,
		&program.CreatedAt, &program.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return program, nil
}

func DeleteProgram(db *sql.DB, key string) error {
	_, err := db.Exec("DELETE FROM programs WHERE key = ?", key)
	return err
}

func GetAllProgramKeys(db *sql.DB, platform string) ([]string, error) {
	var keys []string
	query := "SELECT key FROM programs"
	if platform != "" {
		query += " WHERE platform = ?"
	}

	rows, err := db.Query(query, platform)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	return keys, nil
}

func FetchData(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var lastErr error
	for i := 0; i < 3; i++ {
		resp, err := client.Get(url)
		if err != nil {
			lastErr = err
			log.Printf(" 🍀 Attempt %d failed for %s: %v", i+1, url, err)
			time.Sleep(2 * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("bad status: %s", resp.Status)
			log.Printf(" 🍀 Attempt %d: %s returned %s", i+1, url, resp.Status)
			time.Sleep(2 * time.Second)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		log.Printf(" 🍀 Successfully fetched %s (%d bytes)", url, len(body))
		return body, nil
	}

	return nil, fmt.Errorf("failed to fetch %s after 3 attempts: %v", url, lastErr)
}

func GenerateProgramKey(name, url string) string {
	hash := md5.Sum([]byte(fmt.Sprintf("%s|%s", name, url)))
	return hex.EncodeToString(hash[:])
}

func SerializeScope(scope map[string]string) (string, error) {
	if scope == nil {
		return "{}", nil
	}
	data, err := json.Marshal(scope)
	return string(data), err
}

func DeserializeScope(scopeStr string) (map[string]string, error) {
	scope := make(map[string]string)
	if scopeStr == "" {
		return scope, nil
	}
	err := json.Unmarshal([]byte(scopeStr), &scope)
	return scope, err
}

func SerializeStringArray(arr []string) (string, error) {
	if arr == nil {
		return "[]", nil
	}
	data, err := json.Marshal(arr)
	return string(data), err
}

func DeserializeStringArray(arrStr string) ([]string, error) {
	var arr []string
	if arrStr == "" {
		return arr, nil
	}
	err := json.Unmarshal([]byte(arrStr), &arr)
	return arr, err
}

func SerializeReward(reward Reward) (string, error) {
	data, err := json.Marshal(reward)
	return string(data), err
}

func DeserializeReward(rewardStr string) (Reward, error) {
	var reward Reward
	if rewardStr == "" {
		return reward, nil
	}
	err := json.Unmarshal([]byte(rewardStr), &reward)
	return reward, err
}

func CompareScopes(oldScope, newScope map[string]string) ([]string, []string, []ScopeChange) {
	var newScopes, removedScopes []string
	var changedScopes []ScopeChange

	for id, newValue := range newScope {
		if _, exists := oldScope[id]; !exists {
			newScopes = append(newScopes, newValue)
		}
	}

	for id, oldValue := range oldScope {
		if _, exists := newScope[id]; !exists {
			removedScopes = append(removedScopes, oldValue)
		}
	}

	for id, newValue := range newScope {
		if oldValue, exists := oldScope[id]; exists && oldValue != newValue {
			changedScopes = append(changedScopes, ScopeChange{
				Old: oldValue,
				New: newValue,
			})
		}
	}

	return newScopes, removedScopes, changedScopes
}

func FindNewItems(newList, oldList []string) []string {
	var newItems []string
	for _, item := range newList {
		if !contains(oldList, item) {
			newItems = append(newItems, item)
		}
	}
	return newItems
}

func FindRemovedItems(oldList, newList []string) []string {
	var removedItems []string
	for _, item := range oldList {
		if !contains(newList, item) {
			removedItems = append(removedItems, item)
		}
	}
	return removedItems
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func ShouldSendNotification(isNew bool, alert *Alert, platformConfig Platform, firstRun bool) bool {
	if firstRun {
		return isNew && platformConfig.Notifications.NewProgram
	}

	if alert.IsRemoved {
		return platformConfig.Notifications.RemovedProgram
	}

	if isNew {
		return platformConfig.Notifications.NewProgram
	}

	hasChanges := len(alert.NewScope) > 0 || len(alert.RemovedScope) > 0 ||
		len(alert.ChangedScope) > 0 || alert.NewType != "" || alert.Reward != nil

	return hasChanges
}
