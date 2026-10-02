package models

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestAdditionalStartingItemsEncoding(t *testing.T) {
	tests := []struct {
		name  string
		items []int
		want  string
	}{
		{name: "none", items: []int{}, want: ""},
		{name: "one", items: []int{4}, want: "4"},
		{name: "canonical list", items: []int{4, 12, 182}, want: "4,12,182"},
		{name: "canonicalizes unsorted input", items: []int{182, 4, 12}, want: "4,12,182"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := encodeAdditionalStartingItems(test.items); got != test.want {
				t.Fatalf("expected %q, got %q", test.want, got)
			}
		})
	}
}

func TestAdditionalStartingItemsDecoding(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  []int
	}{
		{name: "none", value: "", want: []int{}},
		{name: "one", value: "4", want: []int{4}},
		{name: "canonical list", value: "4,12,182", want: []int{4, 12, 182}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := decodeAdditionalStartingItems(test.value)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("expected %v, got %v", test.want, got)
			}
		})
	}
}

func TestAdditionalStartingItemsDecodingRejectsInvalidStorage(t *testing.T) {
	if _, err := decodeAdditionalStartingItems("4,nope,12"); err == nil {
		t.Fatal("expected invalid stored data to fail decoding")
	}
}

func TestAdditionalStartingItemsSchemaContract(t *testing.T) {
	for _, filePath := range []string{
		"../../install/database_schema.sql",
		"../../install/additional_starting_items_v1.sql",
	} {
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatal(err)
		}

		sql := string(content)
		if !strings.Contains(sql, "additional_starting_items") ||
			!strings.Contains(sql, "VARCHAR(64)") ||
			!strings.Contains(sql, "NOT NULL") ||
			!strings.Contains(sql, "DEFAULT ''") {
			t.Fatalf("%s does not preserve the Additional Starting Items storage contract", filePath)
		}
	}
}
