package platforms

import (
	"database/sql"
	"fmt"
	"log"
	"pewpew-watcher/utils"

	"github.com/tidwall/gjson"
)

func MonitorBugcrowd(config *utils.Config, db *sql.DB, firstRun bool) {
	platformConfig := config.Platforms["bugcrowd"]
	body, err := utils.FetchData(platformConfig.URL)
	if err != nil {
		log.Printf("  Failed to fetch Bugcrowd data: %v", err)
		return
	}

	log.Printf("🔍 Bugcrowd JSON preview: %s", string(body[:200]))

	programs := gjson.ParseBytes(body).Array()
	log.Printf("📊 Found %d Bugcrowd programs in root array", len(programs))

	var currentKeys []string
	programsCount := 0
	newPrograms := 0
	updatedPrograms := 0

	for _, program := range programs {
		name := program.Get("name").String()
		programURL := program.Get("url").String()

		if name == "" || programURL == "" {
			log.Printf(" 🍀 Skipping program with missing name/url: name=%s, url=%s", name, programURL)
			continue
		}

		url := programURL
		if url == "" {
			briefUrl := program.Get("briefUrl").String()
			if briefUrl != "" {
				url = "https://bugcrowd.com" + briefUrl
			}
		}

		logo := "https://asset.brandfetch.io/idZPL+3f8a/idw6hFgY3p.png"
		logoUrl := program.Get("logo").String()
		if logoUrl == "" {
			logoUrl = program.Get("logoUrl").String()
		}
		if logoUrl != "" {
			logo = logoUrl
		}

		key := utils.GenerateProgramKey(name, url)
		currentKeys = append(currentKeys, key)

		existingProgram, err := utils.GetProgram(db, key)

		programType := "vdp"
		var reward utils.Reward

		scope := make(map[string]string)

		if program.Get("offers_bounties").Bool() || program.Get("bounty").Bool() {
			programType = "rdp"
			reward = utils.Reward{
				Min: program.Get("min_bounty").String(),
				Max: program.Get("max_bounty").String(),
			}
		}

		targets := program.Get("targets").Array()
		for j, target := range targets {
			targetName := target.Get("name").String()
			if targetName == "" {
				targetName = target.String()
			}

			if targetName != "" {
				targetID := fmt.Sprintf("target-%d", j)
				scope[targetID] = targetName
			}
		}

		if len(scope) == 0 {
			targetGroups := program.Get("target_groups").Array()
			targetIndex := 0
			for _, group := range targetGroups {
				targets := group.Get("targets").Array()
				for _, target := range targets {
					targetName := target.Get("name").String()
					if targetName != "" {
						targetID := fmt.Sprintf("target-%d", targetIndex)
						scope[targetID] = targetName
						targetIndex++
					}
				}
			}
		}

		scopeJSON, _ := utils.SerializeScope(scope)
		rewardJSON, _ := utils.SerializeReward(reward)

		newProgram := &utils.Program{
			Name:     name,
			URL:      url,
			Type:     programType,
			Key:      key,
			Platform: "bugcrowd",
			Logo:     logo,
			Scope:    scopeJSON,
			Reward:   rewardJSON,
		}

		alert := &utils.Alert{
			Program:   newProgram,
			IsNew:     false,
			IsRemoved: false,
		}

		if err == sql.ErrNoRows {
			alert.IsNew = true
			alert.NewScope = getScopeValuesBugcrowd(scope)

			if err := utils.SaveProgram(db, newProgram); err != nil {
				log.Printf("  Failed to save new program %s: %v", name, err)
				continue
			}

			if utils.ShouldSendNotification(true, alert, platformConfig, firstRun) {
				utils.SendAlert(alert, config)
				newPrograms++
			}
			programsCount++
			continue
		}

		hasChanged := false
		alert.IsNew = false

		if existingProgram.Type != programType {
			alert.NewType = programType
			hasChanged = true
		}

		existingScope, _ := utils.DeserializeScope(existingProgram.Scope)
		newScope, removedScope, _ := utils.CompareScopes(existingScope, scope)

		if len(newScope) > 0 {
			alert.NewScope = newScope
			hasChanged = true
		}

		if len(removedScope) > 0 {
			alert.RemovedScope = removedScope
			hasChanged = true
		}

		existingReward, _ := utils.DeserializeReward(existingProgram.Reward)
		if existingReward.Min != reward.Min || existingReward.Max != reward.Max {
			alert.Reward = &reward
			hasChanged = true
		}

		if hasChanged {
			if err := utils.SaveProgram(db, newProgram); err != nil {
				log.Printf("  Failed to update program %s: %v", name, err)
				continue
			}

			if utils.ShouldSendNotification(false, alert, platformConfig, firstRun) {
				utils.SendAlert(alert, config)
				updatedPrograms++
			}
		}

		programsCount++
	}

	removedCount := checkRemovedProgramsBugcrowd("bugcrowd", currentKeys, db, platformConfig, firstRun, config)

	log.Printf(" 🍀 Bugcrowd: %d programs processed, %d new, %d updated, %d removed",
		programsCount, newPrograms, updatedPrograms, removedCount)
}

func getScopeValuesBugcrowd(scope map[string]string) []string {
	var values []string
	for _, value := range scope {
		values = append(values, value)
	}
	return values
}

func checkRemovedProgramsBugcrowd(platform string, currentKeys []string, db *sql.DB,
	platformConfig utils.Platform, firstRun bool, config *utils.Config) int {

	existingKeys, err := utils.GetAllProgramKeys(db, platform)
	if err != nil {
		log.Printf("  Failed to get existing keys for %s: %v", platform, err)
		return 0
	}

	removedCount := 0
	for _, existingKey := range existingKeys {
		if !containsBugcrowd(currentKeys, existingKey) {
			removedProgram, err := utils.GetProgram(db, existingKey)
			if err != nil {
				continue
			}

			alert := &utils.Alert{
				Program:   removedProgram,
				IsRemoved: true,
			}

			if utils.ShouldSendNotification(false, alert, platformConfig, firstRun) {
				utils.SendAlert(alert, config)
			}

			if err := utils.DeleteProgram(db, existingKey); err != nil {
				log.Printf("  Failed to delete removed program %s: %v", existingKey, err)
			}
			removedCount++
		}
	}

	return removedCount
}

func containsBugcrowd(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
