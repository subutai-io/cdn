package config

import (
	"testing"
	"strconv"
)

func TestDefaultQuota(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{name: "TestDefaultQuota-"},
		// TODO: Add test cases.
	}
	for i := 1; i <= 2; i++ {
		test := tests[0]
		test.name += strconv.Itoa(i)
		tests = append(tests, test)
	}
	tests[0].name += "0"
	wants := []int{1 << 31, 1 << 21, 1 << 11}
	for i := 0; i <= 2; i++ {
		tests[i].want = wants[i]
	}
	quotas := []string{"2G", "2M", "2K"}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Storage.Userquota = quotas[i]
			if got := DefaultQuota(); got != tt.want {
				t.Errorf("DefaultQuota() = %v, want %v", got, tt.want)
			}
		})
	}
}
