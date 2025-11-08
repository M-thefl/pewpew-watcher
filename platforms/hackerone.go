package platforms

import (
	"database/sql"
	"fmt"
	"log"
	"pewpew-watcher/utils"
	"strings"

	"github.com/tidwall/gjson"
)

func MonitorHackerOne(config *utils.Config, db *sql.DB, firstRun bool) {
	platformConfig := config.Platforms["hackerone"]
	body, err := utils.FetchData(platformConfig.URL)
	if err != nil {
		log.Printf("  Failed to fetch HackerOne data: %v", err)
		return
	}

	log.Printf("🔍 HackerOne JSON preview: %s", string(body[:200]))

	programs := gjson.ParseBytes(body).Array()
	log.Printf("📊 Found %d HackerOne programs in root array", len(programs))

	if len(programs) == 0 {
		programs = gjson.GetBytes(body, "data").Array()
		log.Printf("📊 Found %d HackerOne programs in 'data' path", len(programs))
	}

	var currentKeys []string
	programsCount := 0
	newPrograms := 0
	updatedPrograms := 0

	for _, program := range programs {
		name := program.Get("name").String()
		handle := program.Get("handle").String()
		url := program.Get("url").String()

		if url == "" && handle != "" {
			url = fmt.Sprintf("https://hackerone.com/%s", handle)
		}

		if name == "" || url == "" {
			log.Printf(" 🍀 Skipping program with missing name/url: name=%s, handle=%s", name, handle)
			continue
		}

		logo := "https://asset.brandfetch.io/idhUp0l1vN/id7Vk4WqZc.png"
		profilePic := program.Get("profile_picture").String()
		if profilePic != "" && !strings.Contains(profilePic, "hackerone-us-west-2-p") {
			logo = profilePic
		}

		key := utils.GenerateProgramKey(name, url)
		currentKeys = append(currentKeys, key)

		existingProgram, err := utils.GetProgram(db, key)

		programType := "vdp"
		if program.Get("offers_bounties").Bool() {
			programType = "rdp"
		}
		scope := make(map[string]string)
		scopeData := program.Get("targets.in_scope").Array()

		for j, target := range scopeData {
			targetType := target.Get("type").String()
			assetID := target.Get("asset_identifier").String()

			if assetID == "" {
				assetID = target.Get("asset").String()
			}

			scopeValue := fmt.Sprintf("%s (%s)", assetID, targetType)

			targetID := fmt.Sprintf("%s-%d", handle, j)
			scope[targetID] = scopeValue
		}
		if len(scope) == 0 {
			allTargets := program.Get("targets").Array()
			for j, target := range allTargets {
				if target.Get("eligible_for_submission").Bool() {
					targetType := target.Get("type").String()
					assetID := target.Get("asset_identifier").String()

					if assetID == "" {
						assetID = target.Get("asset").String()
					}

					scopeValue := fmt.Sprintf("%s (%s)", assetID, targetType)
					targetID := fmt.Sprintf("%s-%d", handle, j)
					scope[targetID] = scopeValue
				}
			}
		}

		scopeJSON, _ := utils.SerializeScope(scope)

		newProgram := &utils.Program{
			Name:     name,
			URL:      url,
			Type:     programType,
			Key:      key,
			Platform: "hackerone",
			Logo:     logo,
			Scope:    scopeJSON,
		}

		alert := &utils.Alert{
			Program:   newProgram,
			IsNew:     false,
			IsRemoved: false,
		}

		if err == sql.ErrNoRows {
			alert.IsNew = true
			alert.NewScope = getScopeValuesHackerOne(scope)

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

	removedCount := checkRemovedProgramsHackerOne("hackerone", currentKeys, db, platformConfig, firstRun, config)

	log.Printf(" 🍀 HackerOne: %d programs processed, %d new, %d updated, %d removed",
		programsCount, newPrograms, updatedPrograms, removedCount)
}

func getScopeValuesHackerOne(scope map[string]string) []string {
	var values []string
	for _, value := range scope {
		values = append(values, value)
	}
	return values
}

func checkRemovedProgramsHackerOne(platform string, currentKeys []string, db *sql.DB,
	platformConfig utils.Platform, firstRun bool, config *utils.Config) int {

	existingKeys, err := utils.GetAllProgramKeys(db, platform)
	if err != nil {
		log.Printf("  Failed to get existing keys for %s: %v", platform, err)
		return 0
	}

	removedCount := 0
	for _, existingKey := range existingKeys {
		if !containsHackerOne(currentKeys, existingKey) {
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

func containsHackerOne(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
