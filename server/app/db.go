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

func CountDB(result *Result) int {
	answer := 0
	db.DB.View(func(tx *bolt.Tx) error {
		myBucket := tx.Bucket(db.MyBucket)
		myBucket.ForEach(func(k, v []byte) error {
			file := myBucket.Bucket(k)
			repo := RepoDB(GetResultByFileID(string(k)))
			if result.Repo == "raw" || result.Repo == "template" {
				if repo == "raw" || repo == "template" {
					if hash := file.Bucket([]byte("hash")); hash != nil {
						md5Hash := string(hash.Get([]byte("md5")))
						if md5Hash == result.Md5 {
							answer++
						}
					}
				}
			} else if result.Repo == "apt" {
				filename := string(file.Get([]byte("name")))
				if repo == "apt" && filename == result.Filename {
					answer++
				}
			}
			return nil
		})
		return nil
	})
	return answer
}

func DeleteDB(result *Result) {
	db.DB.Update(func(tx *bolt.Tx) error {
		tx.Bucket(db.MyBucket).DeleteBucket([]byte(result.FileID))
		if user := tx.Bucket(db.Users).Bucket([]byte(result.Owner)); user != nil {
			if files := user.Bucket([]byte("files")); files != nil {
				files.Delete([]byte(result.FileID))
			}
		}
		cleaned := false
		searchIndex := tx.Bucket(db.SearchIndex)
		searchIndex.ForEach(func(k, v []byte) error {
			if string(k) == result.Filename {
				if uploads := searchIndex.Bucket(k); uploads != nil {
					uploadDate := ""
					uploads.ForEach(func(k, v []byte) error {
						if string(v) == result.FileID {
							uploadDate = string(k)
						}
						return nil
					})
					if len(uploadDate) > 0 {
						uploads.Delete([]byte(uploadDate))
					}
					if uploads.Stats().KeyN == 0 {
						cleaned = true
					}
				}
			}
			return nil
		})
		if cleaned == true {
			searchIndex.DeleteBucket([]byte(result.Filename))
		}
		return nil
	})
}

func RepoDB(result *Result) (repo string) {
	db.DB.View(func(tx *bolt.Tx) error {
		if file := tx.Bucket(db.MyBucket).Bucket([]byte(result.FileID)); file != nil {
			if repos := file.Bucket([]byte("type")); repos != nil {
				key, _ := repos.Cursor().First()
				repo = string(key)
			}
		}
		return nil
	})
	return
}

func WriteDB(result *Result) (err error) {
	log.Info(fmt.Sprintf("Writing to DB: %+v", result))
	if result.Owner == "" {
		err = fmt.Errorf("owner wasn't provided")
		log.Warn(fmt.Sprintf("WriteDB error: %v", err))
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
				log.Warn(fmt.Sprintf("WriteDB error: %v", err))
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
			b.Put(now, []byte(result.FileID))
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
