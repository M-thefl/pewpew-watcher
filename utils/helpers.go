package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type DiscordWebhook struct {
	Username  string         `json:"username"`
	AvatarURL string         `json:"avatar_url"`
	Embeds    []DiscordEmbed `json:"embeds"`
}

type DiscordEmbed struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Color       int            `json:"color"`
	Thumbnail   DiscordImage   `json:"thumbnail"`
	Image       DiscordImage   `json:"image"`
	Fields      []DiscordField `json:"fields"`
	Footer      DiscordFooter  `json:"footer"`
	Timestamp   string         `json:"timestamp"`
	URL         string         `json:"url"`
}

type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type DiscordFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url"`
}

type DiscordImage struct {
	URL string `json:"url"`
}

const (
	ColorGreen  = 0x00FF00
	ColorRed    = 0xFF0000
	ColorBlue   = 0x0099FF
	ColorOrange = 0xFFA500
	ColorPurple = 0x9B59B6
	ColorGold   = 0xFFD700
)

var PlatformColors = map[string]int{
	"hackerone": ColorGreen,
	"bugcrowd":  ColorOrange,
	"intigriti": ColorPurple,
	"yeswehack": ColorBlue,
}

func SendAlert(alert *Alert, config *Config) {
	color := PlatformColors[alert.Program.Platform]

	if config.DiscordWebhook != "" {
		sendDiscordAlert(alert, config.DiscordWebhook, color)
	}

	if config.Telegram.BotToken != "" && config.Telegram.ChatID != "" {
		sendTelegramAlert(alert, config.Telegram.BotToken, config.Telegram.ChatID)
	}
}

func sendDiscordAlert(alert *Alert, webhookURL string, color int) {
	webhook := &DiscordWebhook{
		Username:  "🔍 PewPew Watcher",
		AvatarURL: "https://github.com/M-thefl.png",
		Embeds:    []DiscordEmbed{createDiscordEmbed(alert, color)},
	}

	if err := sendWebhook(webhookURL, webhook); err != nil {
		log.Printf(" 🍀Discord alert failed: %v", err)
	} else {
		log.Printf(" 🍀Discord alert sent for %s", alert.Program.Name)
	}
}

func createDiscordEmbed(alert *Alert, color int) DiscordEmbed {
	programImage := alert.Program.Logo
	if programImage == "" {
		programImage = getPlatformLogo(alert.Program.Platform)
	}

	embed := DiscordEmbed{
		Color: color,
		Thumbnail: DiscordImage{
			URL: getPlatformLogo(alert.Program.Platform),
		},
		Image: DiscordImage{
			URL: programImage,
		},
		Footer: DiscordFooter{
			Text: " 🍀 • GitHub: https://github.com/M-thefl",
			// IconURL: "https://github.com/M-thefl.png",
		},
		Timestamp: time.Now().Format(time.RFC3339),
		URL:       alert.Program.URL,
	}

	if alert.IsRemoved {
		embed.Title = fmt.Sprintf("🗑️ %s Removed", alert.Program.Name)
		embed.Description = fmt.Sprintf("**%s** has been removed from **%s**\n\n💔 We'll miss this one!",
			alert.Program.Name, strings.Title(alert.Program.Platform))
		embed.Color = ColorRed
	} else if alert.IsNew {
		embed.Title = fmt.Sprintf("🎉 New Program: %s", alert.Program.Name)
		embed.Description = fmt.Sprintf("**%s** just launched on **%s**!\n\n🔗 [View Program](%s)\n📝 Type: `%s`\n\n🚀 Happy hunting!",
			alert.Program.Name, strings.Title(alert.Program.Platform), alert.Program.URL, alert.Program.Type)
		embed.Color = ColorGold
	} else {
		embed.Title = fmt.Sprintf("📝 %s Updated", alert.Program.Name)
		embed.Description = fmt.Sprintf("**%s** has been updated on **%s**\n\n🔗 [View Changes](%s)\n\n👀 Check what's new!",
			alert.Program.Name, strings.Title(alert.Program.Platform), alert.Program.URL)
	}

	if len(alert.NewScope) > 0 {
		scopeText := formatScope(alert.NewScope, 5)
		embed.Fields = append(embed.Fields, DiscordField{
			Name:   "🆕 New Scope",
			Value:  fmt.Sprintf("```%s```", scopeText),
			Inline: false,
		})
	}

	if len(alert.RemovedScope) > 0 {
		scopeText := formatScope(alert.RemovedScope, 5)
		embed.Fields = append(embed.Fields, DiscordField{
			Name:   " Removed Scope",
			Value:  fmt.Sprintf("```%s```", scopeText),
			Inline: false,
		})
	}

	if alert.NewType != "" {
		embed.Fields = append(embed.Fields, DiscordField{
			Name:   "🔄 Type Changed",
			Value:  fmt.Sprintf("`%s`", alert.NewType),
			Inline: true,
		})
	}

	if alert.Reward != nil {
		embed.Fields = append(embed.Fields, DiscordField{
			Name:   "💰 Bounty Update",
			Value:  fmt.Sprintf("`%s - %s`", alert.Reward.Min, alert.Reward.Max),
			Inline: true,
		})
	}

	embed.Fields = append(embed.Fields, DiscordField{
		Name:   "🔗 Quick Links",
		Value:  "[👤 Follow M-thefl](https://github.com/M-thefl)",
		Inline: false,
	})

	return embed
}

func sendTelegramAlert(alert *Alert, botToken, chatID string) {
	var message string

	if alert.IsRemoved {
		message = fmt.Sprintf("🗑️ *Program Removed*\n\n*Platform:* %s\n*Program:* %s\n*Type:* %s\n\n💔 We'll miss this one!",
			strings.Title(alert.Program.Platform), alert.Program.Name, alert.Program.Type)
	} else if alert.IsNew {
		message = fmt.Sprintf("🎉 *New Program Alert!*\n\n*Platform:* %s\n*Program:* [%s](%s)\n*Type:* `%s`\n\n🚀 Happy hunting!",
			strings.Title(alert.Program.Platform), alert.Program.Name, alert.Program.URL, alert.Program.Type)
	} else {
		message = fmt.Sprintf("📝 *Program Updated*\n\n*Platform:* %s\n*Program:* [%s](%s)\n*Type:* `%s`\n\n👀 Check what's new!",
			strings.Title(alert.Program.Platform), alert.Program.Name, alert.Program.URL, alert.Program.Type)
	}

	if len(alert.NewScope) > 0 {
		message += "\n\n🆕 *New Scope:*\n" + formatScope(alert.NewScope, 3)
	}

	if len(alert.RemovedScope) > 0 {
		message += "\n\n *Removed Scope:*\n" + formatScope(alert.RemovedScope, 3)
	}

	if alert.NewType != "" {
		message += fmt.Sprintf("\n\n🔄 *Type Changed:* `%s`", alert.NewType)
	}

	if alert.Reward != nil {
		message += fmt.Sprintf("\n\n💰 *Bounty Update:* `%s - %s`", alert.Reward.Min, alert.Reward.Max)
	}

	message += "\n\n 🍀*Powered by M-thefl*"

	if err := sendTelegramMessage(botToken, chatID, message); err != nil {
		log.Printf(" 🍀Telegram alert failed: %v", err)
	} else {
		log.Printf(" 🍀Telegram alert sent for %s", alert.Program.Name)
	}
}

func SendStartupMessage(config *Config, firstRun bool) {
	if !firstRun {
		return
	}

	log.Printf("🚀 Sending startup messages...")

	if config.DiscordWebhook != "" {
		sendDiscordStartup(config.DiscordWebhook)
	}

	if config.Telegram.BotToken != "" && config.Telegram.ChatID != "" {
		sendTelegramStartup(config.Telegram.BotToken, config.Telegram.ChatID)
	}
}

func sendDiscordStartup(webhookURL string) {
	webhook := &DiscordWebhook{
		Username:  "🔍 PewPew Watcher 🚀",
		AvatarURL: "https://github.com/M-thefl.png",
		Embeds: []DiscordEmbed{{
			Title:       "🎯 PewPew Watcher Started Successfully!",
			Description: "**Hello Hunter!** 👋\n\nI'm now monitoring your favorite bug bounty platforms for new programs, scope changes, and updates.\n\nStay tuned for real-time alerts! 🔥",
			Color:       ColorPurple,
			Thumbnail: DiscordImage{
				URL: "https://github.com/M-thefl.png",
			},
			Image: DiscordImage{
				URL: "https://i.pinimg.com/originals/af/fb/ad/affbadfe492f696f184f9cb10eb148cf.gif",
			},
			Footer: DiscordFooter{
				Text: "Crafted with 🌙 by M-thefl",
				// IconURL: "https://github.com/M-thefl.png",
			},
			Timestamp: time.Now().Format(time.RFC3339),
			Fields: []DiscordField{
				{
					Name:   "📊 Platforms",
					Value:  "• HackerOne\n• Bugcrowd\n• Intigriti\n• YesWeHack",
					Inline: true,
				},
				{
					Name:   "🔔 Alerts",
					Value:  "• New Programs\n• Scope Changes\n• Program Removals\n• Type Updates",
					Inline: true,
				},
				{
					Name:   "🔗 Quick Links",
					Value:  "[⭐ Star on GitHub](https://github.com/M-thefl/pewpew-watcher)\n[👤 Follow M-thefl](https://github.com/M-thefl)",
					Inline: false,
				},
			},
		}},
	}

	if err := sendWebhook(webhookURL, webhook); err != nil {
		log.Printf(" 🍀Discord startup failed: %v", err)
	} else {
		log.Printf(" 🍀Discord startup sent")
	}
}

func sendTelegramStartup(botToken, chatID string) {
	message := `🚀 *PewPew Watcher Started!*

Hello Hunter! I'm now monitoring bug bounty platforms for you.

*Platforms:* HackerOne, Bugcrowd, Intigriti, YesWeHack
*Alerts:* New programs, scope changes, removals, type updates

Stay tuned for real-time updates!

_Crafted by M-thefl_`

	if err := sendTelegramMessage(botToken, chatID, message); err != nil {
		log.Printf(" 🍀Telegram startup failed: %v", err)
	} else {
		log.Printf(" 🍀Telegram startup sent")
	}
}

func sendWebhook(url string, webhook *DiscordWebhook) error {
	jsonData, err := json.Marshal(webhook)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status: %d", resp.StatusCode)
	}

	return nil
}

func sendTelegramMessage(botToken, chatID, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	payload := map[string]string{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status: %d", resp.StatusCode)
	}

	return nil
}

func formatScope(scope []string, maxItems int) string {
	if len(scope) == 0 {
		return ""
	}

	items := scope
	if len(scope) > maxItems {
		items = scope[:maxItems]
	}

	result := strings.Join(items, "\n")
	if len(scope) > maxItems {
		result += fmt.Sprintf("\n... and %d more", len(scope)-maxItems)
	}

	return result
}

func shorten(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
}

func getPlatformLogo(platform string) string {
	logos := map[string]string{
		"hackerone": "https://asset.brandfetch.io/idhUp0l1vN/id7Vk4WqZc.png",
		"bugcrowd":  "https://asset.brandfetch.io/idZPL+3f8a/idw6hFgY3p.png",
		"intigriti": "https://api.intigriti.com/file/api/file/public_bucket_d23a1f29-c2fe-4d03-8daf-df24d1e076ea-c2449aa2-3a08-4bf5-a430-441a11020851",
		"yeswehack": "https://pbs.twimg.com/profile_images/1154580610072817664/5YjR6tTI_400x400.png",
	}
	return logos[platform]
}
