package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"pewpew-watcher/platforms"
	"pewpew-watcher/utils"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Config struct {
	DiscordWebhook string                    `json:"DiscordWebhook"`
	Telegram       utils.TelegramConfig      `json:"telegram"`
	Database       utils.DatabaseConfig      `json:"database"`
	Platforms      map[string]utils.Platform `json:"platforms"`
}

func loadConfig() *Config {
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal("  Config file error:", err)
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		log.Fatal("  Config parse error:", err)
	}

	log.Printf("🔧 Webhook loaded: %v", config.DiscordWebhook != "")
	if config.DiscordWebhook != "" {
		log.Printf("🔧 Webhook preview: %s...", config.DiscordWebhook[:30])
	}

	return &config
}

func main() {
	fmt.Println(`
	██████╗ ███████╗ ██╗    ██╗██████╗ ███████╗ ██╗    ██╗
	██╔══██╗██╔════╝ ██║    ██║██╔══██╗██╔════╝ ██║    ██║
	██████╔╝█████╗   ██║ █╗ ██║██████╔╝█████╗   ██║ █╗ ██║
	██╔═══╝ ██╔══╝   ██║███╗██║██╔══██╗██╔══╝   ██║███╗██║
	██║     ███████╗ ╚███╔███╔╝██║  ██║███████╗ ╚███╔███╔╝
	╚═╝     ╚══════╝  ╚══╝╚══╝ ╚═╝  ╚═╝╚══════╝  ╚══╝╚══╝ 
														 
	🚀 PewPew Watcher v1.0 by M-thefl
	📧 GitHub: https://github.com/M-thefl
	`)

	config := loadConfig()

	db, err := sql.Open("sqlite", config.Database.Path)
	if err != nil {
		log.Fatal("  Database connection error:", err)
	}
	defer db.Close()

	utils.InitDatabase(db)

	firstRun := utils.IsFirstRun(db)
	log.Printf("🎯 First run: %v", firstRun)

	utilsConfig := &utils.Config{
		DiscordWebhook: config.DiscordWebhook,
		Telegram:       config.Telegram,
		Database:       config.Database,
		Platforms:      config.Platforms,
	}

	startTime := time.Now()

	log.Println("📨 Sending startup message...")
	utils.SendStartupMessage(utilsConfig, firstRun)
	log.Println(" 🍀 Startup message process completed")
	time.Sleep(2 * time.Second)
	log.Println("🕵️ Starting platform monitoring...")
	testAllAPIs(config.Platforms)
	log.Printf("🔧 First run mode: %v - %s", firstRun, getFirstRunMessage(firstRun))
	monitorPlatforms(utilsConfig, db, firstRun)

	if firstRun {
		log.Println("🧪Sending test notification after first run...")
		sendTestNotification(utilsConfig)
	}

	duration := time.Since(startTime)
	log.Printf(" 🍀 Monitoring completed in %v", duration)
	log.Println("🎉 PewPew Watcher finished successfully!")
}

func getFirstRunMessage(firstRun bool) string {
	if firstRun {
		return "Storing programs in database, no notifications will be sent"
	}
	return "Monitoring for changes and sending notifications"
}

func testAllAPIs(platforms map[string]utils.Platform) {
	log.Println("🧪 Testing API connections...")

	for name, platform := range platforms {
		if !platform.Monitor {
			log.Printf("⏭️ %s: Monitoring disabled", strings.Title(name))
			continue
		}

		log.Printf("🔗 Testing %s API...", strings.Title(name))
		body, err := utils.FetchData(platform.URL)
		if err != nil {
			log.Printf(" %s: %v", strings.Title(name), err)
		} else {
			log.Printf("🍀 %s: Success (%d bytes)", strings.Title(name), len(body))
		}
		time.Sleep(1 * time.Second)
	}
}

func monitorPlatforms(config *utils.Config, db *sql.DB, firstRun bool) {
	platformsList := []struct {
		name    string
		monitor func(*utils.Config, *sql.DB, bool)
	}{
		{"hackerone", platforms.MonitorHackerOne},
		{"bugcrowd", platforms.MonitorBugcrowd},
		{"intigriti", platforms.MonitorIntigriti},
		// {"yeswehack", platforms.MonitorYesWeHack},
	}

	successCount := 0
	for _, platform := range platformsList {
		if platformConfig, exists := config.Platforms[platform.name]; exists && platformConfig.Monitor {
			log.Printf("🔍 Monitoring %s...", strings.Title(platform.name))
			platform.monitor(config, db, firstRun)
			successCount++
		} else {
			log.Printf("⏭️ Skipping %s (disabled)", strings.Title(platform.name))
		}
	}

	if successCount == 0 {
		log.Println(" 🍀 No platforms were monitored. Check your config.json!")
	} else {
		log.Printf("🎊 Successfully monitored %d platform(s)", successCount)
	}
}

func sendTestNotification(config *utils.Config) {
	testAlert := &utils.Alert{
		Program: &utils.Program{
			Name:     "PewPew Watcher Test",
			URL:      "https://github.com/M-thefl/pewpew-watcher",
			Type:     "TEST",
			Platform: "hackerone",
			Logo:     "https://github.com/M-thefl.png",
		},
		IsNew:     true,
		IsRemoved: false,
		NewScope:  []string{"*.example.com", "api.example.com"},
	}

	log.Println("🧪 Sending test notification...")
	utils.SendAlert(testAlert, config)
	log.Println(" 🍀Test notification sent!")
}
