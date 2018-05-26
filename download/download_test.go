package download

import (
	"testing"
	"reflect"
)


func TestIn(t *testing.T) {
	type args struct {
		str  string
		list []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"TestIn-1", args{"in", []string{"string", "in", "slice"}}, true},
		{"TestIn-2", args{"in", []string(nil)}, false},
		{"TestIn-3", args{"", []string{"", "in", "slice"}}, true},
		{"TestIn-4", args{"outside", []string{"", "in", "slice"}}, false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := In(tt.args.str, tt.args.list); got != tt.want {
				t.Errorf("In() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_intersect(t *testing.T) {
	type args struct {
		listA []string
		listB []string
	}
	tests := []struct {
		name     string
		args     args
		wantList []string
	}{
		{"Test_intersect-1", args{[]string(nil), []string(nil)}, []string(nil)},
		{"Test_intersect-2", args{[]string{"in", "out"}, []string{"in", "in"}}, []string{"in"}},
		{"Test_intersect-3", args{[]string{"in"}, []string{"out"}}, []string(nil)},
		{"Test_intersect-4", args{[]string{"hello", "world"}, []string{"world", "hello"}}, []string{"world", "hello"}},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotList := intersect(tt.args.listA, tt.args.listB); !reflect.DeepEqual(gotList, tt.wantList) {
				t.Errorf("intersect() = %v, want %v", gotList, tt.wantList)
			}
		})
	}
}

func Test_unique(t *testing.T) {
	type args struct {
		list []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"Test_unique-1", args{[]string{"in", "in"}}, []string{"in"}},
		{"Test_unique-2", args{[]string{"in", "out"}}, []string{"in", "out"}},
		{"Test_unique-3", args{[]string(nil)}, []string(nil)},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unique(tt.args.list); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_union(t *testing.T) {
	type args struct {
		listA []string
		listB []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"Test_union-1", args{[]string{"in", "out"}, []string{"in", "out"}}, []string{"in", "out"}},
		{"Test_union-2", args{[]string{"in"}, []string{"in", "out"}}, []string{"in", "out"}},
		{"Test_union-3", args{[]string(nil), []string(nil)}, []string(nil)},
		{"Test_union-4", args{[]string(nil), []string{"in"}}, []string{"in"}},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := union(tt.args.listA, tt.args.listB); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("union() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_processVersion(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Test_processVersion-1", args{"latest"}, ""},
		{"Test_processVersion-2", args{"7.0.0"}, "7.0.0"},
		{"Test_processVersion-3", args{""}, ""},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := processVersion(tt.args.version); got != tt.want {
				t.Errorf("processVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}


