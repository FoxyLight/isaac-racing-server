package models

import (
	"reflect"
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
