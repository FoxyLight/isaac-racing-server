package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"sort"
	"strconv"
)

type AdditionalStartingItemsPolicy struct {
	Version                         int                               `json:"version"`
	MaxAdditionalItems              int                               `json:"maxAdditionalItems"`
	SeededMaxResolvedItems          int                               `json:"seededMaxResolvedItems"`
	VanillaCollectibleIDRange       AdditionalStartingItemsIDRange    `json:"vanillaCollectibleIdRange"`
	SupportedRaceFormats            []string                          `json:"supportedRaceFormats"`
	RankedSoloSupported             bool                              `json:"rankedSoloSupported"`
	TaintedLazarusSupported         bool                              `json:"taintedLazarusSupported"`
	GloballyBannedCollectibleIDs    []int                             `json:"globallyBannedCollectibleIds"`
	QuestProgressionCollectibleIDs  []int                             `json:"questProgressionCollectibleIds"`
	Eligibility                     AdditionalStartingItemsEligibility `json:"eligibility"`
	SeededEligibilitySource         string                            `json:"seededEligibilitySource"`
	CanonicalOrder                  string                            `json:"canonicalOrder"`
}

type AdditionalStartingItemsIDRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type AdditionalStartingItemsEligibility struct {
	RequireShown                bool `json:"requireShown"`
	ExcludeActiveItems          bool `json:"excludeActiveItems"`
	ExcludeOutsideVanillaRange  bool `json:"excludeOutsideVanillaRange"`
}

var additionalStartingItemsPolicy AdditionalStartingItemsPolicy

func loadAdditionalStartingItemsPolicy() {
	jsonFilePath := path.Join(libPath, "additionalStartingItems.json")
	jsonFile, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		logger.Fatal("Failed to open \""+jsonFilePath+"\":", err)
	}

	if err := json.Unmarshal(jsonFile, &additionalStartingItemsPolicy); err != nil {
		logger.Fatal("Failed to unmarshal the Additional Starting Items policy:", err)
	}
}

func validateAndCanonicalizeAdditionalStartingItems(
	ruleset *Ruleset,
	items map[string]*JSONItem,
	builds []Build,
	policy AdditionalStartingItemsPolicy,
) error {
	if ruleset.AdditionalStartingItems == nil {
		ruleset.AdditionalStartingItems = make([]int, 0)
	}

	if len(ruleset.AdditionalStartingItems) == 0 {
		return nil
	}

	if len(ruleset.AdditionalStartingItems) > policy.MaxAdditionalItems {
		return fmt.Errorf("you cannot select more than %d additional starting items", policy.MaxAdditionalItems)
	}

	if !stringInSlice(string(ruleset.Format), policy.SupportedRaceFormats) {
		return fmt.Errorf("additional starting items are not supported for the %s race format", ruleset.Format)
	}

	if ruleset.Ranked && ruleset.Solo && !policy.RankedSoloSupported {
		return fmt.Errorf("additional starting items are not supported for ranked solo races")
	}

	if ruleset.Character == "Tainted Lazarus" && !policy.TaintedLazarusSupported {
		return fmt.Errorf("additional starting items are not supported for Tainted Lazarus")
	}

	seen := make(map[int]bool)
	for _, itemID := range ruleset.AdditionalStartingItems {
		if seen[itemID] {
			return fmt.Errorf("additional starting item %d was selected more than once", itemID)
		}
		seen[itemID] = true

		if policy.Eligibility.ExcludeOutsideVanillaRange &&
			(itemID < policy.VanillaCollectibleIDRange.Min || itemID > policy.VanillaCollectibleIDRange.Max) {
			return fmt.Errorf("additional starting item %d is outside the supported collectible range", itemID)
		}

		item, ok := items[strconv.Itoa(itemID)]
		if !ok {
			return fmt.Errorf("additional starting item %d does not exist", itemID)
		}

		if policy.Eligibility.RequireShown && !item.Shown {
			return fmt.Errorf("additional starting item %d is not selectable", itemID)
		}

		if policy.Eligibility.ExcludeActiveItems && item.SpaceBar {
			return fmt.Errorf("additional starting item %d is an active item", itemID)
		}

		if intInSlice(itemID, policy.GloballyBannedCollectibleIDs) {
			return fmt.Errorf("additional starting item %d is globally banned", itemID)
		}

		if intInSlice(itemID, policy.QuestProgressionCollectibleIDs) {
			return fmt.Errorf("additional starting item %d is a quest or progression item", itemID)
		}
	}

	if ruleset.Format == RaceFormatSeeded {
		if ruleset.StartingBuildIndex < 0 || ruleset.StartingBuildIndex >= len(builds) {
			return fmt.Errorf("seeded races with additional starting items require a concrete valid starting build")
		}

		build := builds[ruleset.StartingBuildIndex]
		if len(build.Collectibles)+len(ruleset.AdditionalStartingItems) > policy.SeededMaxResolvedItems {
			return fmt.Errorf(
				"seeded races may resolve to at most %d total starting items",
				policy.SeededMaxResolvedItems,
			)
		}

		for _, collectible := range build.Collectibles {
			if seen[collectible.ID] {
				return fmt.Errorf(
					"additional starting item %d duplicates an item in the seeded build",
					collectible.ID,
				)
			}
		}
	}

	sort.Ints(ruleset.AdditionalStartingItems)
	return nil
}

func intInSlice(value int, values []int) bool {
	for _, candidate := range values {
		if value == candidate {
			return true
		}
	}
	return false
}
