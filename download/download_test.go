package download

import (
	"testing"
	"reflect"
	"github.com/subutai-io/agent/log"
	"fmt"
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
		{name: "TestFormatItem-0"},
		{name: "TestFormatItem-1"},
		// TODO: Add test cases.
	}
	{
		tests[0].args.info = make(map[string]string)
		tests[0].args.info["id"] = "some-id"
		tests[0].args.info["md5"] = "some-md5"
		tests[0].args.info["sha256"] = "some-sha256"
		tests[0].args.info["arch"] = "some-arch"
		tests[0].args.info["name"] = "some-name-subutai-template"
		tests[0].args.info["size"] = "123"
		tests[0].args.info["parent"] = "some-parent"
		tests[0].args.info["parent-owner"] = "some-parent-owner"
		tests[0].args.info["parent-version"] = "some-parent-version"
		tests[0].args.info["version"] = "some-version"
		tests[0].args.info["Description"] = "some-Description"
		tests[0].args.repo = "template"
	}
	{
		tests[1].args.info = make(map[string]string)
		tests[1].args.info["id"] = "some-id"
		tests[1].args.info["sha256"] = "some-sha256"
		tests[1].args.info["SHA256"] = "some-sha256"
		tests[1].args.info["arch"] = "some-arch"
		tests[1].args.info["Architecture"] = "some-arch"
		tests[1].args.info["name"] = "some-name-apt"
		tests[1].args.info["size"] = "123"
		tests[1].args.info["Size"] = "123"
		tests[1].args.info["parent"] = "some-parent"
		tests[1].args.info["parent-owner"] = "some-parent-owner"
		tests[1].args.info["parent-version"] = "some-parent-version"
		tests[1].args.info["version"] = "some-version"
		tests[1].args.info["Version"] = "some-version"
		tests[1].args.info["Description"] = "some-Description"
		tests[1].args.repo = "apt"
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatItem(tt.args.info, tt.args.repo)
			if got.ID == tt.args.info["id"] &&
				got.Filename == tt.args.info["name"] &&
				((tt.name == "TestFormatItem-0" && got.Name == "some-name") || (tt.name == "TestFormatItem-1" && got.Name == "some-name-apt")) &&
				((tt.name == "TestFormatItem-0" && got.Hash.Md5 == "some-md5") || (tt.name == "TestFormatItem-1" && got.Hash.Md5 == tt.args.info["id"])) &&
				got.Hash.Sha256 == tt.args.info["sha256"] &&
				got.Version == tt.args.info["version"] &&
				got.Parent == tt.args.info["parent"] &&
				got.ParentOwner == tt.args.info["parent-owner"] &&
				got.ParentVersion == tt.args.info["parent-version"] &&
				((tt.name == "TestFormatItem-0" && got.Prefsize == "tiny") || (tt.name == "TestFormatItem-1" && got.Prefsize == "")) &&
				got.Description == tt.args.info["Description"] &&
				((tt.name == "TestFormatItem-0" && got.Architecture == "SOME-ARCH") || (tt.name == "TestFormatItem-1" && got.Architecture == "some-arch")) {
				log.Info(fmt.Sprintf("Test %s passes", tt.name))
			} else {
				t.Errorf("Test %s didn't pass: check of %+v vs %+v didn't pass", tt.name, got, tt.args.info)
			}
		})
	}
}
