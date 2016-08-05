package db

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/subutai-io/base/agent/log"

	"github.com/subutai-io/gorjun/config"
)

var (
	bucket = []byte("MyBucket")
	search = []byte("SearchIndex")
	users  = []byte("Users")
	tokens = []byte("Tokens")
	authid = []byte("AuthID")
	db     = initdb()
)

func initdb() *bolt.DB {
	db, err := bolt.Open(config.Path+"my.db", 0600, nil)
	log.Check(log.FatalLevel, "Openning DB: my.db", err)

	err = db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{bucket, search, users, tokens, authid} {
			_, err := tx.CreateBucketIfNotExists(b)
			log.Check(log.FatalLevel, "Creating bucket: "+string(b), err)
		}
		return nil
	})
	log.Check(log.FatalLevel, "Finishing update transaction", err)

	return db
}

func Write(owner, key, value string, options ...map[string]string) {
	if len(owner) == 0 {
		owner = "public"
	}
	err := db.Update(func(tx *bolt.Tx) error {
		now, _ := time.Now().MarshalText()

		// Associating files with user
		b, _ := tx.Bucket(users).CreateBucketIfNotExists([]byte(owner))
		if b, err := b.CreateBucketIfNotExists([]byte("files")); err == nil {
			b.Put([]byte(key), []byte(value))
		}

		// Creating new record about file
		if b, err := tx.Bucket(bucket).CreateBucket([]byte(key)); err == nil {
			b.Put([]byte("date"), now)
			b.Put([]byte("name"), []byte(value))

			// Getting file size
			if f, err := os.Open(config.Filepath + key); err == nil {
				fi, _ := f.Stat()
				f.Close()
				b.Put([]byte("size"), []byte(fmt.Sprint(fi.Size())))
			}

			// Writing optional parameters for file
			for i, _ := range options {
				for k, v := range options[i] {
					b.Put([]byte(k), []byte(v))
				}
			}

			// Adding search index for files
			b, _ = tx.Bucket(search).CreateBucketIfNotExists([]byte(value))
			b.Put(now, []byte(key))

		}

		// Adding owners and shares to files
		if b := tx.Bucket(bucket).Bucket([]byte(key)); b != nil {
			if b, _ = b.CreateBucketIfNotExists([]byte("owner")); b != nil {
				//If value is not empty, we are assuming that it is a signature (or any other personal info)
				//Otherwise we are just adding new owner
				if len(value) != 0 && len(options) == 0 {
					b.Put([]byte(owner), []byte(value))
				} else {
					b.Put([]byte(owner), []byte("w"))
				}
			}
			if b, _ = b.CreateBucketIfNotExists([]byte("scope")); b != nil {
				if b, _ = b.CreateBucketIfNotExists([]byte(owner)); b != nil {
				}
			}
		}
		return nil
	})
	log.Check(log.WarnLevel, "Writing data to db", err)
}

func Delete(owner, key string) (remains int) {
	db.Update(func(tx *bolt.Tx) error {
		var filename []byte

		// Deleting file association with user
		if b := tx.Bucket(users).Bucket([]byte(owner)); b != nil {
			if b := b.Bucket([]byte("files")); b != nil {
				filename = b.Get([]byte(key))
				b.Delete([]byte(key))
			}
		}

		// Deleting user association with file
		if b := tx.Bucket(bucket).Bucket([]byte(key)); b != nil {
			if b := b.Bucket([]byte("owner")); b != nil {
				b.Delete([]byte(owner))
				remains = b.Stats().KeyN - 1
			}
		}

		// Removing indexes and file only if no file owners left
		if remains <= 0 {
			// Deleting search index
			if b := tx.Bucket(search).Bucket([]byte(filename)); b != nil {
				b.ForEach(func(k, v []byte) error {
					if string(v) == key {
						b.Delete(k)
					}
					return nil
				})
			}

			// Removing file from DB
			tx.Bucket(bucket).DeleteBucket([]byte(key))
		}
		return nil
	})

	return
}

func Read(key string) (val string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket).Bucket([]byte(key)); b != nil {
			if value := b.Get([]byte("name")); value != nil {
				val = string(value)
			}
		}
		return nil
	})
	return val
}

func List() map[string]string {
	list := make(map[string]string)
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		b.ForEach(func(k, v []byte) error {
			if value := b.Bucket(k).Get([]byte("name")); value != nil {
				list[string(k)] = string(value)
			}
			return nil
		})
		return nil
	})
	return list
}

func Info(hash string) map[string]string {
	list := make(map[string]string)
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket).Bucket([]byte(hash)); b != nil {
			b.ForEach(func(k, v []byte) error {
				list[string(k)] = string(v)
				return nil
			})
		}
		return nil
	})
	list["owner"] = "public"
	return list
}

func Close() {
	db.Close()
}

func Search(query string) map[string]string {
	list := make(map[string]string)
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(search)
		c := b.Cursor()
		for k, _ := c.Seek([]byte(query)); len(k) > 0 && bytes.HasPrefix(k, []byte(query)); k, _ = c.Next() {
			b.Bucket(k).ForEach(func(kk, vv []byte) error {
				list[string(vv)] = string(k)
				return nil
			})
		}
		return nil
	})
	return list
}

func LastHash(name, t string) (hash string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(search).Bucket([]byte(name)); b != nil {
			c := b.Cursor()
			for k, v := c.Last(); k != nil; k, v = c.Prev() {
				if t == "" || t == Info(string(v))["type"] {
					hash = string(v)
					break
				}
			}
		}
		return nil
	})
	return hash
}

func RegisterUser(name, key []byte) {
	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.Bucket(users).CreateBucket([]byte(name))
		if !log.Check(log.WarnLevel, "Registering user "+string(name), err) {
			b.Put([]byte("key"), key)
		}
		return err
	})
}

func UserKey(name string) (key string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(users).Bucket([]byte(name)); b != nil {
			if value := b.Get([]byte("key")); value != nil {
				key = string(value)
			}
		}
		return nil
	})
	return key
}

func SaveToken(name, token string) {
	db.Update(func(tx *bolt.Tx) error {
		if b, _ := tx.Bucket(tokens).CreateBucketIfNotExists([]byte(token)); b != nil {
			b.Put([]byte("name"), []byte(name))
			now, _ := time.Now().MarshalText()
			b.Put([]byte("date"), now)
		}
		return nil
	})
}

func CheckToken(token string) (name string) {
	token = fmt.Sprintf("%x", sha256.Sum256([]byte(token)))

	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(tokens).Bucket([]byte(token)); b != nil {
			date := new(time.Time)
			date.UnmarshalText(b.Get([]byte("date")))
			if date.Add(time.Minute * 60).Before(time.Now()) {
				return nil
			}
			if value := b.Get([]byte("name")); value != nil {
				name = string(value)
			}
		}
		return nil
	})
	return name
}

func SaveAuthID(name, token string) {
	db.Update(func(tx *bolt.Tx) error {
		tx.Bucket(authid).Put([]byte(token), []byte(name))
		return nil
	})
}

func CheckAuthID(token string) (name string) {
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(authid)
		if value := b.Get([]byte(token)); value != nil {
			name = string(value)
			b.Delete([]byte(token))
		}
		return nil
	})
	return name
}

// FileOwner provides list of file owners
func FileOwner(hash string) (list []string) {
	list = []string{}
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket).Bucket([]byte(hash)); b != nil {
			if b := b.Bucket([]byte("owner")); b != nil {
				b.ForEach(func(k, v []byte) error {
					list = append(list, string(k))
					return nil
				})
			}
		}
		return nil
	})
	return list
}

// CheckOwner checks if user owns particular file
func CheckOwner(owner, hash string) (res bool) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket).Bucket([]byte(hash)); b != nil {
			if b := b.Bucket([]byte("owner")); b != nil && b.Get([]byte(owner)) != nil {
				res = true
			}
		}
		return nil
	})
	return res
}

func FileSignatures(hash string) (list map[string]string) {
	list = map[string]string{}
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket).Bucket([]byte(hash)); b != nil {
			if b := b.Bucket([]byte("owner")); b != nil {
				b.ForEach(func(k, v []byte) error {
					if string(v) != "w" {
						list[string(k)] = string(v)
					}
					return nil
				})
			}
		}
		return nil
	})
	return list
}

// UserFile searching file at particular user. It returns list of hashes of files with required name.
func UserFile(owner, file string) (list []string) {
	if len(owner) == 0 {
		owner = "public"
	}
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(users).Bucket([]byte(owner)); b != nil {
			if files := b.Bucket([]byte("files")); files != nil {
				files.ForEach(func(k, v []byte) error {
					if string(v) == file {
						list = append(list, string(k))
					}
					fmt.Println(string(k), string(v))
					return nil
				})
			}
		}
		return nil
	})
	return list
}

func GetScope(hash, owner string) (scope []string) {
	scope = []string{}
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket).Bucket([]byte(hash)); b != nil {
			if b := b.Bucket([]byte("scope")); b != nil {
				if b := b.Bucket([]byte(owner)); b != nil {
					b.ForEach(func(k, v []byte) error {
						scope = append(scope, string(k))
						return nil
					})
				}
			}
		}
		return nil
	})
	return scope
}

func ShareWith(hash, owner, user string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket).Bucket([]byte(hash)); b != nil {
			if b := b.Bucket([]byte("scope")); b != nil {
				if b := b.Bucket([]byte(owner)); b != nil {
					b.Put([]byte(user), []byte("w"))
				}
			}
		}
		return nil
	})
}

func UnshareWith(hash, owner, user string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket).Bucket([]byte(hash)); b != nil {
			if b := b.Bucket([]byte("scope")); b != nil {
				if b := b.Bucket([]byte(owner)); b != nil {
					b.Delete([]byte(user))
				}
			}
		}
		return nil
	})
}

func CheckShare(hash, user string) bool {
	shared := false
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket).Bucket([]byte(hash)); b != nil {
			if b := b.Bucket([]byte("scope")); b != nil {
				b.ForEach(func(k, v []byte) error {
					if b := b.Bucket(k); b != nil {
						b.ForEach(func(k1, v1 []byte) error {
							if string(k1) == user {
								shared = true
								return nil
							}
							return nil
						})
					}
					return nil
				})

			}
		}
		return nil
	})
	return shared
}

func Public(hash string) bool {
	public := false
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket).Bucket([]byte(hash)); b != nil {
			if b := b.Bucket([]byte("scope")); b != nil {
				b.ForEach(func(k, v []byte) error {
					if b := b.Bucket([]byte(k)); b != nil {
						k, _ := b.Cursor().First()
						if k == nil {
							public = true
						}
					}
					return nil
				})
			}
		}
		return nil
	})
	return public
}
