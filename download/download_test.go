/*
Generated TestHandler
Generated TestInfo
Generated TestList
Generated Test_processVersion
Generated TestIn
Generated TestGetVerified
Generated TestFormatItem
Generated Test_intersect
Generated Test_unique
*/

package download

import (
	"net/http"
	"reflect"
	"testing"
)

func TestHandler(t *testing.T) {
	type args struct {
		repo string
		w    http.ResponseWriter
		r    *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Handler(tt.args.repo, tt.args.w, tt.args.r)
		})
	}
}

func TestInfo(t *testing.T) {
	type args struct {
		repo string
		r    *http.Request
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Info(tt.args.repo, tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Info() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestList(t *testing.T) {
	type args struct {
		repo string
		r    *http.Request
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := List(tt.args.repo, tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("List() = %v, want %v", got, tt.want)
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

func TestGetVerified(t *testing.T) {
	type args struct {
		list            []string
		name            string
		repo            string
		versionTemplate string
	}
	tests := []struct {
		name string
		args args
		want ListItem
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetVerified(tt.args.list, tt.args.name, tt.args.repo, tt.args.versionTemplate); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetVerified() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatItem(t *testing.T) {
	type args struct {
		info map[string]string
		repo string
	}
	tests := []struct {
		name string
		args args
		want ListItem
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatItem(tt.args.info, tt.args.repo); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FormatItem() = %v, want %v", got, tt.want)
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
