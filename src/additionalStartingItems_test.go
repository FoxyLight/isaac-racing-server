package server

import (
	"reflect"
	"strings"
	"testing"
)

func testAdditionalStartingItemsPolicy() AdditionalStartingItemsPolicy {
	return AdditionalStartingItemsPolicy{
		Version:                1,
		MaxAdditionalItems:     5,
		SeededMaxResolvedItems: 4,
		VanillaCollectibleIDRange: AdditionalStartingItemsIDRange{
			Min: 1,
			Max: 732,
		},
		SupportedRaceFormats: []string{
			string(RaceFormatUnseeded),
			string(RaceFormatSeeded),
			string(RaceFormatCustom),
		},
		RankedSoloSupported:     false,
		TaintedLazarusSupported: false,
		GloballyBannedCollectibleIDs: []int{
			590,
			721,
		},
		QuestProgressionCollectibleIDs: []int{
			238,
			239,
			327,
			328,
			550,
			551,
			552,
			626,
			627,
			633,
			668,
		},
		Eligibility: AdditionalStartingItemsEligibility{
			RequireShown:               true,
			ExcludeActiveItems:         true,
			ExcludeOutsideVanillaRange: true,
		},
		SeededEligibilitySource: "BUILDS",
		CanonicalOrder:          "ascending-id",
	}
}

func testAdditionalStartingItemsItems() map[string]*JSONItem {
	return map[string]*JSONItem{
		"4":   {Name: "Cricket's Head", Shown: true},
		"8":   {Name: "Brother Bobby", Shown: true},
		"12":  {Name: "Magic Mushroom", Shown: true},
		"33":  {Name: "The Bible", Shown: true, SpaceBar: true},
		"50":  {Name: "Steven", Shown: true},
		"67":  {Name: "Sister Maggy", Shown: true},
		"182": {Name: "Sacred Heart", Shown: true},
		"238": {Name: "Key Piece 1", Shown: true},
		"590": {Name: "Global Ban A", Shown: true},
		"721": {Name: "Global Ban B", Shown: true},
		"700": {Name: "Hidden Test Item", Shown: false},
	}
}

func testAdditionalStartingItemsBuilds() []Build {
	return []Build{
		{
			Name: "One Item",
			Collectibles: []Collectible{
				{ID: 4, Name: "Cricket's Head"},
			},
		},
		{
			Name: "Two Items",
			Collectibles: []Collectible{
				{ID: 8, Name: "Brother Bobby"},
				{ID: 50, Name: "Steven"},
			},
		},
	}
}

func TestAdditionalStartingItemsNormalizesAndCanonicalizes(t *testing.T) {
	policy := testAdditionalStartingItemsPolicy()
	items := testAdditionalStartingItemsItems()
	builds := testAdditionalStartingItemsBuilds()

	empty := &Ruleset{Format: RaceFormatUnseeded}
	if err := validateAndCanonicalizeAdditionalStartingItems(empty, items, builds, policy); err != nil {
		t.Fatal(err)
	}
	if empty.AdditionalStartingItems == nil || len(empty.AdditionalStartingItems) != 0 {
		t.Fatalf("expected missing items to normalize to an empty slice: %#v", empty.AdditionalStartingItems)
	}

	ruleset := &Ruleset{
		Format:                  RaceFormatUnseeded,
		AdditionalStartingItems: []int{182, 4, 12},
	}
	if err := validateAndCanonicalizeAdditionalStartingItems(ruleset, items, builds, policy); err != nil {
		t.Fatal(err)
	}

	expected := []int{4, 12, 182}
	if !reflect.DeepEqual(ruleset.AdditionalStartingItems, expected) {
		t.Fatalf("expected canonical order %v, got %v", expected, ruleset.AdditionalStartingItems)
	}
}

func TestAdditionalStartingItemsRejectsInvalidSelections(t *testing.T) {
	policy := testAdditionalStartingItemsPolicy()
	items := testAdditionalStartingItemsItems()
	builds := testAdditionalStartingItemsBuilds()

	tests := []struct {
		name    string
		ruleset Ruleset
		want    string
	}{
		{
			name: "too many",
			ruleset: Ruleset{
				Format:                  RaceFormatUnseeded,
				AdditionalStartingItems: []int{4, 8, 12, 50, 67, 182},
			},
			want: "more than 5",
		},
		{
			name: "duplicate",
			ruleset: Ruleset{
				Format:                  RaceFormatUnseeded,
				AdditionalStartingItems: []int{4, 4},
			},
			want: "more than once",
		},
		{
			name: "active",
			ruleset: Ruleset{
				Format:                  RaceFormatUnseeded,
				AdditionalStartingItems: []int{33},
			},
			want: "active item",
		},
		{
			name: "hidden",
			ruleset: Ruleset{
				Format:                  RaceFormatUnseeded,
				AdditionalStartingItems: []int{700},
			},
			want: "not selectable",
		},
		{
			name: "outside range",
			ruleset: Ruleset{
				Format:                  RaceFormatUnseeded,
				AdditionalStartingItems: []int{733},
			},
			want: "outside",
		},
		{
			name: "global ban",
			ruleset: Ruleset{
				Format:                  RaceFormatUnseeded,
				AdditionalStartingItems: []int{590},
			},
			want: "globally banned",
		},
		{
			name: "quest item",
			ruleset: Ruleset{
				Format:                  RaceFormatUnseeded,
				AdditionalStartingItems: []int{238},
			},
			want: "quest or progression",
		},
		{
			name: "diversity",
			ruleset: Ruleset{
				Format:                  RaceFormatDiversity,
				AdditionalStartingItems: []int{4},
			},
			want: "not supported",
		},
		{
			name: "ranked solo",
			ruleset: Ruleset{
				Format:                  RaceFormatUnseeded,
				Ranked:                  true,
				Solo:                    true,
				AdditionalStartingItems: []int{4},
			},
			want: "ranked solo",
		},
		{
			name: "tainted lazarus",
			ruleset: Ruleset{
				Format:                  RaceFormatUnseeded,
				Character:               "Tainted Lazarus",
				AdditionalStartingItems: []int{4},
			},
			want: "Tainted Lazarus",
		},
		{
			name: "seeded duplicate of build",
			ruleset: Ruleset{
				Format:                  RaceFormatSeeded,
				StartingBuildIndex:      0,
				AdditionalStartingItems: []int{4},
			},
			want: "duplicates",
		},
		{
			name: "seeded total limit",
			ruleset: Ruleset{
				Format:                  RaceFormatSeeded,
				StartingBuildIndex:      1,
				AdditionalStartingItems: []int{4, 12, 182},
			},
			want: "at most 4",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ruleset := test.ruleset
			err := validateAndCanonicalizeAdditionalStartingItems(&ruleset, items, builds, policy)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected error containing %q, got %v", test.want, err)
			}
		})
	}
}

func TestAdditionalStartingItemsSupportsAllowedRulesets(t *testing.T) {
	policy := testAdditionalStartingItemsPolicy()
	items := testAdditionalStartingItemsItems()
	builds := testAdditionalStartingItemsBuilds()

	tests := []Ruleset{
		{
			Format:                  RaceFormatUnseeded,
			Ranked:                  true,
			Solo:                    false,
			AdditionalStartingItems: []int{12},
		},
		{
			Format:                  RaceFormatCustom,
			AdditionalStartingItems: []int{12},
		},
		{
			Format:                  RaceFormatSeeded,
			StartingBuildIndex:      0,
			AdditionalStartingItems: []int{12, 182},
		},
	}

	for _, ruleset := range tests {
		ruleset := ruleset
		if err := validateAndCanonicalizeAdditionalStartingItems(&ruleset, items, builds, policy); err != nil {
			t.Fatalf("expected ruleset to be allowed, got %v", err)
		}
	}
}
