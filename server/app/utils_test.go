package app

import (
	"testing"

	"strconv"

	"github.com/boltdb/bolt"
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
			DB.DB.Update(func(tx *bolt.Tx) error {
				if b := tx.Bucket(DB.Users); b != nil {
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
	tokens := []string{"RandomStuff", Ipsum.Token, Lorem.Token, Subutai.Token}
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
	PreDownloadFiles(0, Lorem)
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
		"99b79c3764cee9740583d14b377dc393",
		"99b79c3764cee9740583d14b377dc393",
		"d2fb80e1629180f0788056967beced556bcd9a24",
		"8a74dec550a12658beafa1a92be38888b78016566e2683632e647582935e6a2b",
		"c4fd87ee78740529e792ae6541a54084805e79a19176901c39ad33ec5aec78b63b14eec8c5ca66d74fe4754070e9531d572e7e1f048b8417994138872fa1c5b0",
		"",
	}
	for i := 1; i <= 5; i++ {
		test := tests[0]
		test.name += strconv.Itoa(i)
		test.args = args{
			file: Dirs[PublicScope][Lorem.Username] + Files[PublicScope][Lorem.Username][NamesLayer][0],
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
			file: Dirs[PublicScope][Lorem.Username] + Files[PublicScope][Lorem.Username][NamesLayer][0],
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
	tests[0].name += "0"
	{
		tests[0].args = args{
			item: "ipsum",
			list: []string{"Lorem", "ipsum", "ipsum", "sit", "amet"},
		}
		tests[0].want = true
	}
	{
		tests[1].args = args{
			item: "",
			list: []string{"Lorem", "ipsum", "ipsum", "sit", "amet"},
		}
		tests[1].want = false
	}
	{
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

func TestSetLogLevel(t *testing.T) {
	type args struct {
		logLevel string
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "TestSetLogLevel-"},
		// TODO: Add test cases.
	}
	logLevels := []string{"", "panic", "fatal", "error", "warn", "info", "debug"}
	for i := 1; i <= 6; i++ {
		test := tests[0]
		test.args = args{
			logLevel: logLevels[i],
		}
		test.name += strconv.Itoa(i)
		tests = append(tests, test)
	}
	tests[0].name += "0"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLogLevel(tt.args.logLevel)
		})
	}
}
