package app

import (
	"testing"

	"github.com/boltdb/bolt"
	"github.com/subutai-io/cdn/db"
	"strconv"
)

func TestCheckOwner(t *testing.T) {
	Integration = 0
	SetUp()
	defer TearDown()
	type args struct {
		owner string
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "t1", args: args{owner: "ExistedOwner"}},
		{name: "t2", args: args{owner: "NotExistingOwner"}},
	}
	for _, tt := range tests {
		if tt.name == "t1" {
			db.DB.Update(func(tx *bolt.Tx) error {
				if b := tx.Bucket(db.Users); b != nil {
					b.CreateBucket([]byte(tt.args.owner))
				}
				return nil
			})
			if got := CheckOwner(tt.args.owner); got != true {
				t.Errorf("%q. CheckOwner. Owner exist", tt.name)
			}
		} else if tt.name == "t2" {
			if got := CheckOwner(tt.args.owner); got != false {
				t.Errorf("%q. CheckOwner. Owner is not exist", tt.name)
			}
		}
	}
}

func TestCheckToken(t *testing.T) {
	Integration = 0
	SetUp()
	PrepareUsersAndTokens()
	defer TearDown()
	type args struct {
		token string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "TestCheckToken-"},
		// TODO: Add test cases.
	}
	tokens := []string{"RandomStuff", Abaytulakova.Token, Akenzhaliev.Token, Subutai.Token}
	for i := 1; i <= 3; i++ {
		test := tests[0]
		test.name += strconv.Itoa(i)
		test.args = args{
			token: tokens[i],
		}
		test.want = true
		tests = append(tests, test)
	}
	tests[0].name += "0"
	tests[0].args = args{
		token: tokens[0],
	}
	tests[0].want = false
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckToken(tt.args.token); got != tt.want {
				t.Errorf("CheckToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHash(t *testing.T) {
	Integration = 0
	SetUp()
	PreDownloadFiles(0, Akenzhaliev)
	defer TearDown()
	type args struct {
		file string
		algo string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "TestHash-"},
		// TODO: Add test cases.
	}
	algos := []string{"", "md5", "sha1", "sha256", "sha512", ""}
	wants := []string{
		"a315ae66e49e95ccfb832f3852bd3827",
		"a315ae66e49e95ccfb832f3852bd3827",
		"f24f260763d351b0f3249574a9437ad6b15d100d",
		"2a8f000dc3c2e86174c98d721c652b9b2fdc6784931ca4b0a1dd2c7fbcffb851",
		"0dae80b90a1c82a90257772a78407b41132a40ebc8f84eb956222f636ca6cae1188438c9c9a2a5262b77b6543cf893b6a6ffd75433c262d99770ed27107bd153",
		"",
	}
	for i := 1; i <= 5; i++ {
		test := tests[0]
		test.name += strconv.Itoa(i)
		test.args = args{
			file: Dirs[PublicScope][Akenzhaliev.Username] + Files[PublicScope][Akenzhaliev.Username][NamesLayer][0],
			algo: algos[i],
		}
		test.want = wants[i]
		tests = append(tests, test)
	}
	{
		tests[5].args = args{
			file: "/foo/bar/zoo",
		}
	}
	{
		tests[0].name += "0"
		tests[0].args = args{
			file: Dirs[PublicScope][Akenzhaliev.Username] + Files[PublicScope][Akenzhaliev.Username][NamesLayer][0],
			algo: algos[0],
		}
		tests[0].want = wants[0]
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Hash(tt.args.file, tt.args.algo); got != tt.want {
				t.Errorf("Hash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIn(t *testing.T) {
	Integration = 0
	type args struct {
		item string
		list []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "TestIn-"},
		// TODO: Add test cases.
	}
	for i := 1; i <= 2; i++ {
		test := tests[0]
		test.name += strconv.Itoa(i)
		tests = append(tests, test)
	}
	{
		tests[0].name += "0"
		tests[0].args = args{
			item: "ipsum",
			list: []string{"Lorem", "ipsum", "dolor", "sit", "amet"},
		}
		tests[0].want = true
	}
	{
		tests[1].name += "0"
		tests[1].args = args{
			item: "",
			list: []string{"Lorem", "ipsum", "dolor", "sit", "amet"},
		}
		tests[1].want = false
	}
	{
		tests[2].name += "0"
		tests[2].args = args{
			item: "",
			list: []string{},
		}
		tests[2].want = false
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := In(tt.args.item, tt.args.list); got != tt.want {
				t.Errorf("In() = %v, want %v", got, tt.want)
			}
		})
	}
}
