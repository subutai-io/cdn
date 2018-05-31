package auto

import "testing"

func TestCleanGarbage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"TestCleanGarbage-1"},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CleanGarbage()
		})
	}
}

func Test_stringInSlice(t *testing.T) {
	type args struct {
		a    string
		list []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Test_stringInSlice-1", args{"in", []string{"string", "in", "slice"}}, true},
		{"Test_stringInSlice-2", args{"in", []string(nil)}, false},
		{"Test_stringInSlice-3", args{"", []string{"", "in", "slice"}}, true},
		{"Test_stringInSlice-4", args{"outside", []string{"", "in", "slice"}}, false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringInSlice(tt.args.a, tt.args.list); got != tt.want {
				t.Errorf("stringInSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
