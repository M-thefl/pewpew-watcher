# 🔍 PewPew Watcher 

<div align="center">

![PewPew Watcher](https://github.com/M-thefl.png?size=100)
  
**Real-time Bug Bounty Platform Monitor**  
*Never miss a new program or scope change again!*

[![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/M-thefl/pewpew-watcher?style=for-the-badge)](https://github.com/M-thefl/pewpew-watcher/stargazers)
[![GitHub Issues](https://img.shields.io/github/issues/M-thefl/pewpew-watcher?style=for-the-badge)](https://github.com/M-thefl/pewpew-watcher/issues)

</div>

## ✨ Features

- 🚨 **Real-time Monitoring** - Instant alerts for new programs and changes
- 📱 **Multi-Platform Support** - HackerOne, Bugcrowd, Intigriti, YesWeHack
- 🔔 **Smart Notifications** - Discord & Telegram integration
- 🎯 **Scope Tracking** - Monitor scope additions, removals, and changes
- 💰 **Bounty Alerts** - Get notified about reward changes
- 🛡️ **VDP/RDP Detection** - Automatic program type classification
- 📊 **Database Backed** - SQLite for persistent storage
- ⚡ **Lightweight** - Built with Go for high performance

## 📸 Screenshots

### Discord Notifications
![Discord Alert](https://via.placeholder.com/800x400/7289DA/FFFFFF?text=Discord+Notifications+Preview)

### Telegram Alerts
![Telegram Alert](https://via.placeholder.com/400x600/0088CC/FFFFFF?text=Telegram+Alerts+Preview)

## 🚀 Quick Start

### Prerequisites
- Go 1.19 or higher
- Discord Webhook URL (optional)
- Telegram Bot Token (optional)

### Installation

1. **Clone the repository**
  ```bash
git clone https://github.com/M-thefl/pewpew-watcher.git
cd pewpew-watcher
```

2. **Build the application**
```bash
go build -o pewpew-watcher main.go
```
3. **Configure your settings**
```json
{
  "DiscordWebhook": "https://discord.com/api/webhooks/...",
  "telegram": {
    "BotToken": "123456:ABC-DEF...",
    "ChatID": "-1001234567890"
  },
  "database": {
    "path": "programs.db"
  },
  "platforms": {
    "hackerone": {
      "url": "https://github.com/arkadiyt/bounty-targets-data/raw/main/data/hackerone_data.json",
      "monitor": true
    },
    "bugcrowd": {
      "url": "https://github.com/arkadiyt/bounty-targets-data/raw/main/data/bugcrowd_data.json",
      "monitor": true
    },
    "intigriti": {
      "url": "https://github.com/arkadiyt/bounty-targets-data/raw/main/data/intigriti_data.json",
      "monitor": true
    },
    "yeswehack": {
      "url": "https://github.com/arkadiyt/bounty-targets-data/raw/main/data/yeswehack_data.json",
      "monitor": true
    }
  }
}
```
4. **Run the watcher**
  ```bash
./pewpew-watcher
```

# ⚙️ Configuration
**Discord Setup**
- 1 Create a new webhook in your Discord server
- 2 Copy the webhook URL to config.json

**Telegram Setup**
- 1 Create a bot with @BotFather
- 2 Get your bot token and chat ID
- 3  Add them to config.json

**Platform Configuration**
- Each platform can be enabled/disabled individually in the config file.


# 🏗️ Architecture 

<img width="2828" height="2290" alt="flll" src="https://github.com/user-attachments/assets/4ee2ba5d-05b3-42df-8e61-c4b096982653" />

# 📋 Supported Platforms
Platform | Status | Features
------------ | ---------- | --------------------------
HackerOne | ✅ Active | Programs, Scope, Bounties
Bugcrowd | ✅ Active | Programs, Scope, Rewards
Intigriti | ✅ Active | Programs, Scope, Bounties
YesWeHack | 🔄 Planned | Coming Soon

# 🎯 Alert Types
- 🎉 New Programs - When a new bug bounty program launches

- 📝 Scope Changes - Additions or removals from program scope

- 🔄 Type Updates - VDP to RDP transitions

- 💰 Bounty Changes - Reward amount updates

- 🗑️ Program Removals - When programs are discontinued

# 🛠️ Development
**Building from Source**
```bash
# Clone the repository
git clone https://github.com/M-thefl/pewpew-watcher.git
cd pewpew-watcher

# Install dependencies
go mod download

# Build the application
go build -o pewpew-watcher main.go

# Run tests
go test ./...
```


## 📞 Support
For questions, support, and responsible disclosure:
- 📧 Email: Mahbodfl1@gmail.com
- 💬 Telegram: @Mahbodfl
- 🐛 Issues: GitHub Issues Page

## 📜 License
This project is licensed under the MIT License - see the LICENSE file for details.
```sql
MIT License
Copyright (c) 2024 Zoozanaghe Development Team

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## 🌟 Community
Join our Discord community for discussions
https://discord.gg/K2RdmqrM93

Participate in security research collaborations

Attend our virtual workshops and training sessions

Contribute to open source security projects

<div align="center"> ⭐ Show your support Give a star ⭐ if this project helped you in your security research!

"With great power comes great responsibility" - Use wisely and ethically 🧠

