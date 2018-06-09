package download

import (
	"testing"
	"github.com/subutai-io/agent/log"
	"fmt"
)

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

func Test_checkVersion(t *testing.T) {
	type args struct {
		items []ListItem
		item  ListItem
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "Test_checkVersion-0"},
		{name: "Test_checkVersion-1"},
		{name: "Test_checkVersion-2"},
		{name: "Test_checkVersion-3"},
		// TODO: Add test cases.
	}
	tests[0].args.items = append(tests[0].args.items, ListItem{Name: "debian-stretch", Owner: []string{"subutai"}, Version: "0.4.1"})
	tests[0].args.item = ListItem{Name: "debian-stretch", Owner: []string{"subutai"}, Version: "0.4.5"}
	tests[0].want = 0
	tests[1].args.items = append(tests[1].args.items, ListItem{Name: "debian-stretch", Owner: []string{"subutai"}, Version: "0.4.5"})
	tests[1].args.item = ListItem{Name: "debian-stretch", Owner: []string{"subutai"}, Version: "0.4.1"}
	tests[1].want = -1
	tests[2].args.items = append(tests[2].args.items, ListItem{Name: "debian-stretch", Owner: []string{"somebody"}, Version: "0.4.1"})
	tests[2].args.item = ListItem{Name: "debian-stretch", Owner: []string{"subutai"}, Version: "0.4.5"}
	tests[2].want = 1
	tests[3].args.items = append(tests[3].args.items, ListItem{Name: "debian-stretch", Owner: []string{"somebody"}, Version: "0.4.1"})
	tests[3].args.items = append(tests[3].args.items, ListItem{Name: "debian-stretch", Owner: []string{"subutai"}, Version: "0.4.2"})
	tests[3].args.item = ListItem{Name: "debian-stretch", Owner: []string{"subutai"}, Version: "0.4.3"}
	tests[3].want = 1
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.Warn(fmt.Sprintf("Starting test"))
			log.Warn(fmt.Sprintf("tt.args.items = %+v", tt.args.items))
			log.Warn(fmt.Sprintf("tt.args.item = %+v", tt.args.item))
			if got := checkVersion(tt.args.items, tt.args.item); got != tt.want {
				t.Errorf("checkVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
