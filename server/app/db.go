package app

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/subutai-io/agent/log"
	"github.com/subutai-io/cdn/db"
)

func WriteDB(results ...*Result) error {
	for i := range results {
		var err error
		result := results[i]
		if len(result.Owner) == 0 {
			err = fmt.Errorf("WriteDB: Owner wasn't provided")
		} else {
			err = db.DB.Update(func(tx *bolt.Tx) error {
				now, _ := time.Now().MarshalText()
				// Associating files with user
				b, _ := tx.Bucket(db.Users).CreateBucketIfNotExists([]byte(result.Owner))
				if b, err := b.CreateBucketIfNotExists([]byte("files")); err == nil {
					if v := b.Get([]byte(result.FileID)); v == nil {
						// log.Warn("Associating: " + owner + " with " + value + " (" + key + ")")
						b.Put([]byte(result.FileID), []byte(result.Filename))
					}
				}
				// Creating new record about file
				if b, err := tx.Bucket(db.MyBucket).CreateBucket([]byte(result.FileID)); err == nil {
					b.Put([]byte("date"), now)
					b.Put([]byte("name"), []byte(result.Filename))
					// Adding SearchIndex index for files
					b, _ = tx.Bucket(db.SearchIndex).CreateBucketIfNotExists([]byte(strings.ToLower(result.Filename)))
					b.Put(now, []byte(result.Filename))
				}
				// Adding owners, shares and tags to files
				if b := tx.Bucket(db.MyBucket).Bucket([]byte(result.FileID)); b != nil {
					if c, err := b.CreateBucket([]byte("owner")); err == nil {
						log.Info(fmt.Sprintf("Bucket owner created successfully"))
						c.Put([]byte(result.Owner), []byte("w"))
					}
					if _, err := b.CreateBucket([]byte("scope")); err == nil {
						log.Info(fmt.Sprintf("Bucket scope created successfully"))
					}
					if c, err := b.CreateBucket([]byte("hash")); err == nil {
						log.Info(fmt.Sprintf("Bucket hash created successfully"))
						c.Put([]byte("md5"), []byte(result.Md5))
						c.Put([]byte("sha256"), []byte(result.Sha256))
					}
					if c, err := b.CreateBucket([]byte("type")); err == nil {
						log.Info(fmt.Sprintf("Bucket type created successfully"))
						if d, err := c.CreateBucket([]byte(result.Repo)); err == nil {
							d.Put([]byte(result.Owner), []byte("w"))
						}
					}
					//convert int64 to []byte and then write in db
					sz := make([]byte, 8)
					binary.LittleEndian.PutUint64(sz, uint64(result.Size))
					if result.Repo == "raw" {
						b.Put([]byte("size"), sz)
						b.Put([]byte("version"), []byte(result.Version))
						b.Put([]byte("tag"), []byte(result.Tags))
					} else if result.Repo == "template" {
						b.Put([]byte("size"), sz)
						b.Put([]byte("Description"), []byte(result.Description))
						b.Put([]byte("arch"), []byte(result.Architecture))
						b.Put([]byte("parent"), []byte(result.Parent))
						b.Put([]byte("parent-owner"), []byte(result.ParentOwner))
						b.Put([]byte("parent-version"), []byte(result.ParentVersion))
						b.Put([]byte("prefsize"), []byte(result.PrefSize))
						b.Put([]byte("version"), []byte(result.Version))
						c, _ := b.CreateBucket([]byte("tags"))
						for _, tag := range strings.Split(result.Tags, ",") {
							c.Put([]byte(tag), []byte("w"))
						}
					} else if result.Repo == "apt" {
						b.Put([]byte("Size"), sz)
						b.Put([]byte("Description"), []byte(result.Description))
						b.Put([]byte("Architecture"), []byte(result.Architecture))
						b.Put([]byte("Version"), []byte(result.Version))
						b.Put([]byte("tag"), []byte(result.Tags))
					} else {
						return fmt.Errorf("Unrecognized repo %s", result.Repo)
					}
				}
				return nil
			})
		}
		if err != nil {
			return err
		}
	}
	return nil
}
