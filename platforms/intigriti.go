package platforms

import (
	"database/sql"
	"fmt"
	"log"
	"pewpew-watcher/utils"

	"github.com/tidwall/gjson"
)

func MonitorIntigriti(config *utils.Config, db *sql.DB, firstRun bool) {
	platformConfig := config.Platforms["intigriti"]
	body, err := utils.FetchData(platformConfig.URL)
	if err != nil {
		log.Printf("Failed to fetch Intigriti data: %v", err)
		return
	}

	log.Printf("🔍 Intigriti JSON preview: %s", string(body[:200]))

	programs := gjson.ParseBytes(body).Array()
	log.Printf("📊 Found %d Intigriti programs in root array", len(programs))

	var currentKeys []string
	programsCount := 0
	newPrograms := 0
	updatedPrograms := 0

	for _, program := range programs {
		name := program.Get("name").String()
		url := program.Get("url").String()

		if name == "" || url == "" {
			log.Printf(" 🍀 Skipping program with missing name/url: name=%s, url=%s", name, url)
			continue
		}

		logo := "https://api.intigriti.com/file/api/file/public_bucket_d23a1f29-c2fe-4d03-8daf-df24d1e076ea-c2449aa2-3a08-4bf5-a430-441a11020851"
		programLogo := program.Get("logo").String()
		if programLogo != "" {
			logo = programLogo
		}

		key := utils.GenerateProgramKey(name, url)
		currentKeys = append(currentKeys, key)
		existingProgram, err := utils.GetProgram(db, key)

		programType := "vdp"
		if program.Get("maxBounty").Exists() || program.Get("bounty").Bool() {
			programType = "rdp"
		}
		scope := make(map[string]string)

		targets := program.Get("targets").Map()
		if inScope, exists := targets["in_scope"]; exists {
			inScopeArray := inScope.Array()
			for i, target := range inScopeArray {
				targetName := target.Get("name").String()
				targetType := target.Get("type").String()

				if targetName == "" {
					targetName = target.Get("endpoint").String()
				}
				if targetName == "" {
					targetName = target.Get("target").String()
				}
				if targetType == "" {
					targetType = "unknown"
				}

				targetID := fmt.Sprintf("inscope-%d", i)
				scope[targetID] = fmt.Sprintf("%s (%s)", targetName, targetType)
			}
		}
		if len(scope) == 0 {
			inScopeArray := program.Get("in_scope").Array()
			for i, target := range inScopeArray {
				targetName := target.Get("name").String()
				targetType := target.Get("type").String()

				if targetName == "" {
					targetName = target.Get("endpoint").String()
				}
				if targetName == "" {
					targetName = target.String() // just a string
				}
				if targetType == "" {
					targetType = "unknown"
				}

				targetID := fmt.Sprintf("inscope-%d", i)
				scope[targetID] = fmt.Sprintf("%s (%s)", targetName, targetType)
			}
		}
		if len(scope) == 0 {
			domains := program.Get("domains").Array()
			for i, domain := range domains {
				domainStr := domain.String()
				if domainStr != "" {
					targetID := fmt.Sprintf("domain-%d", i)
					scope[targetID] = fmt.Sprintf("%s (domain)", domainStr)
				}
			}
		}

		scopeJSON, _ := utils.SerializeScope(scope)

		newProgram := &utils.Program{
			Name:     name,
			URL:      url,
			Type:     programType,
			Key:      key,
			Platform: "intigriti",
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
			alert.NewScope = getScopeValuesIntigriti(scope)

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

	removedCount := checkRemovedProgramsIntigriti("intigriti", currentKeys, db, platformConfig, firstRun, config)

	log.Printf("Intigriti: %d programs processed, %d new, %d updated, %d removed",
		programsCount, newPrograms, updatedPrograms, removedCount)
}

func getScopeValuesIntigriti(scope map[string]string) []string {
	var values []string
	for _, value := range scope {
		values = append(values, value)
	}
	return values
}

func checkRemovedProgramsIntigriti(platform string, currentKeys []string, db *sql.DB,
	platformConfig utils.Platform, firstRun bool, config *utils.Config) int {

	existingKeys, err := utils.GetAllProgramKeys(db, platform)
	if err != nil {
		log.Printf("  Failed to get existing keys for %s: %v", platform, err)
		return 0
	}

	removedCount := 0
	for _, existingKey := range existingKeys {
		if !containsIntigriti(currentKeys, existingKey) {
			// Program removed
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

func containsIntigriti(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
