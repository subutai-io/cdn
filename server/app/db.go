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

func FileWrite(result *Result) (err error) {
	log.Info(fmt.Sprintf("Writing to DB: %+v", result))
	if result.Owner == "" {
		err = fmt.Errorf("owner wasn't provided")
		log.Warn(fmt.Sprintf("FileWrite error: %v", err))
		return err
	}
	err = db.DB.Update(func(tx *bolt.Tx) error {
		now, _ := time.Now().MarshalText()
		if b := tx.Bucket(db.Users).Bucket([]byte(result.Owner)); b != nil {
			if b, err := b.CreateBucketIfNotExists([]byte("files")); err == nil {
				if v := b.Get([]byte(result.FileID)); v == nil {
					b.Put([]byte(result.FileID), []byte(result.Filename))
				}
			} else {
				log.Warn(fmt.Sprintf("FileWrite error: %v", err))
				return err
			}
		} else {
			err := fmt.Errorf("user doesn't exist")
			log.Warn(fmt.Sprintf("WriteDB error: %v", err))
			return err
		}
		if b, err := tx.Bucket(db.MyBucket).CreateBucket([]byte(result.FileID)); err == nil {
			b.Put([]byte("date"), now)
			b.Put([]byte("name"), []byte(result.Filename))
			b, _ = tx.Bucket(db.SearchIndex).CreateBucketIfNotExists([]byte(strings.ToLower(result.Filename)))
			b.Put(now, []byte(result.Filename))
		} else {
			log.Warn(fmt.Sprintf("WriteDB error: %v", err))
			return err
		}
		if b := tx.Bucket(db.MyBucket).Bucket([]byte(result.FileID)); b != nil {
			if c, err := b.CreateBucket([]byte("owner")); err == nil {
				c.Put([]byte(result.Owner), []byte("w"))
				log.Info(fmt.Sprintf("Bucket owner set up successfully"))
			} else {
				log.Warn(fmt.Sprintf("WriteDB error: %v", err))
				return err
			}
			if _, err := b.CreateBucket([]byte("scope")); err == nil {
				log.Info(fmt.Sprintf("Bucket scope set up successfully"))
			} else {
				log.Warn(fmt.Sprintf("WriteDB error: %v", err))
				return err
			}
			if c, err := b.CreateBucket([]byte("hash")); err == nil {
				c.Put([]byte("md5"), []byte(result.Md5))
				c.Put([]byte("sha256"), []byte(result.Sha256))
				log.Info(fmt.Sprintf("Bucket hash set up successfully"))
			} else {
				log.Warn(fmt.Sprintf("WriteDB error: %v", err))
				return err
			}
			if c, err := b.CreateBucket([]byte("type")); err == nil {
				if d, err := c.CreateBucket([]byte(result.Repo)); err == nil {
					d.Put([]byte(result.Owner), []byte("w"))
					log.Info(fmt.Sprintf("Bucket type set up successfully"))
				} else {
					log.Warn(fmt.Sprintf("WriteDB error: %v", err))
					return err
				}
			} else {
				log.Warn(fmt.Sprintf("WriteDB error: %v", err))
				return err
			}
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
				err := fmt.Errorf("unrecognized repo %s", result.Repo)
				log.Warn(fmt.Sprintf("WriteDB error: %v", err))
				return err
			}
		} else {
			err := fmt.Errorf("couldn't open file's bucket")
			log.Warn(fmt.Sprintf("WriteDB error: %v", err))
			return err
		}
		log.Info(fmt.Sprintf("File successfully added to repo %s", result.Repo))
		return nil
	})
	return err
}
