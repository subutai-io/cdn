package app

import (
	"testing"

	"github.com/boltdb/bolt"
	"github.com/subutai-io/cdn/db"
)

func TestCheckOwner(t *testing.T) {
	SetUp(false)
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
	TearDown(false)
}
