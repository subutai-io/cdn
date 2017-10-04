package db

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/subutai-io/agent/log"
	"github.com/subutai-io/gorjun/config"
)

var (
	bucket = []byte("MyBucket")
	search = []byte("SearchIndex")
	users  = []byte("Users")
	tokens = []byte("Tokens")
	authID = []byte("AuthID")
	tags   = []byte("Tags")
	db     = initDB()
)

func initDB() *bolt.DB {
	os.MkdirAll(filepath.Dir(config.DB.Path), 0755)
	os.MkdirAll(config.Storage.Path, 0755)
	db, err := bolt.Open(config.DB.Path, 0600, &bolt.Options{Timeout: 3 * time.Second})
	log.Check(log.FatalLevel, "Opening DB: "+config.DB.Path, err)
	err = db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{bucket, search, users, tokens, authID, tags} {
			_, err := tx.CreateBucketIfNotExists(b)
			log.Check(log.FatalLevel, "Creating bucket: "+string(b), err)
		}
		return nil
	})
	log.Check(log.FatalLevel, "Finishing update transaction", err)

	return db
}

// Write create record about file in DB
func Write(owner, key, value string, options ...map[string]string) {
	if len(owner) == 0 {
		owner = "subutai"
	}
	err := db.Update(func(tx *bolt.Tx) error {
		now, _ := time.Now().MarshalText()

		// Associating files with user
		b, _ := tx.Bucket(users).CreateBucketIfNotExists([]byte(owner))
		if b, err := b.CreateBucketIfNotExists([]byte("files")); err == nil {
			if v := b.Get([]byte(key)); v == nil {
				// log.Warn("Associating: " + owner + " with " + value + " (" + key + ")")
				b.Put([]byte(key), []byte(value))
			}
		}

		// Creating new record about file
		if b, err := tx.Bucket(bucket).CreateBucket([]byte(key)); err == nil {
			b.Put([]byte("date"), now)
			b.Put([]byte("name"), []byte(value))

			// Adding search index for files
			b, _ = tx.Bucket(search).CreateBucketIfNotExists([]byte(strings.ToLower(value)))
			b.Put(now, []byte(key))
		}

		// Adding owners, shares and tags to files
		if b := tx.Bucket(bucket).Bucket([]byte(key)); b != nil {
			if c, err := b.CreateBucket([]byte("owner")); err == nil {
				c.Put([]byte(owner), []byte("w"))
			}
			for i := range options {
				for k, v := range options[i] {
					switch k {
					case "type":
						if c, err := b.CreateBucketIfNotExists([]byte("type")); err == nil {
							if c, err := c.CreateBucketIfNotExists([]byte(v)); err == nil {
								c.Put([]byte(owner), []byte("w"))
							}
						}
					case "md5", "sha256":
						if c, err := b.CreateBucketIfNotExists([]byte("hash")); err == nil {
							c.Put([]byte(k), []byte(v))
							// Getting file size
							if f, err := os.Open(config.Storage.Path + v); err == nil {
								fi, _ := f.Stat()
								f.Close()
								b.Put([]byte("size"), []byte(fmt.Sprint(fi.Size())))
							}
						}
					case "tags":
						if c, err := b.CreateBucketIfNotExists([]byte("tags")); err == nil && len(v) > 0 {
							for _, v := range strings.Split(v, ",") {
								tag := []byte(strings.ToLower(strings.TrimSpace(v)))
								t, _ := tx.Bucket(tags).CreateBucketIfNotExists(tag)
								c.Put(tag, []byte("w"))
								t.Put([]byte(key), []byte("w"))
							}
						}
					case "signature":
						if c, err := b.CreateBucketIfNotExists([]byte("owner")); err == nil {
							c.Put([]byte(owner), []byte(v))
						}
					default:
						if b.Get([]byte(k)) == nil {
							b.Put([]byte(k), []byte(v))
						}
					}
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

// Delete removes record about file from DB
func Delete(owner, repo, key string) (total int) {
	db.Update(func(tx *bolt.Tx) error {
		var filename []byte

		owned := CheckRepo(owner, "", key)
		md5, _ := Hash(key)
		if key == md5 {
			total = CheckRepo("", "", key)
		} else {
			total = CheckRepo("", "", md5)
		}

		// Deleting user association with file
		if b := tx.Bucket(bucket).Bucket([]byte(key)); b != nil {
			if d := b.Bucket([]byte("type")); d != nil {
				if d := d.Bucket([]byte(repo)); d != nil {
					d.Delete([]byte(owner))
				}
			}

			if c := b.Bucket([]byte("scope")); owned == 1 && c != nil {
				c.Delete([]byte(owner))
			}
			if b := b.Bucket([]byte("owner")); owned == 1 && b != nil {
				b.Delete([]byte(owner))
			}
		}

		// Deleting file association with user
		if b := tx.Bucket(users).Bucket([]byte(owner)); owned == 1 && b != nil {
			if b := b.Bucket([]byte("files")); b != nil {
				filename = b.Get([]byte(key))
				b.Delete([]byte(key))
			}
		}

		// Removing indexes and file only if no file owners left
		if total == 1 || key != md5 {
			// Deleting search index
			if b := tx.Bucket(search).Bucket(bytes.ToLower(filename)); b != nil {
				b.ForEach(func(k, v []byte) error {
					if string(v) == key {
						b.Delete(k)
					}
					return nil
				})
			}

			for _, tag := range FileField(key, "tags") {
				if s := tx.Bucket(tags).Bucket([]byte(tag)); s != nil {
					log.Check(log.DebugLevel, "Removing tag "+tag+" from index bucket", s.Delete([]byte(key)))
				}
			}

			// Removing file from DB
			tx.Bucket(bucket).DeleteBucket([]byte(key))
		}
		return nil
	})
	return total - 1
}

// Read returns name by ID
func Read(key string) (name string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket).Bucket([]byte(key)); b != nil {
			if value := b.Get([]byte("name")); value != nil {
				name = string(value)
			}
		}
		return nil
	})
	return name
}

// Hash returns hash sums by ID
func Hash(key string) (md5, sha256 string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket).Bucket([]byte(key)); b != nil {
			if b := b.Bucket([]byte("hash")); b != nil {
				if value := b.Get([]byte("md5")); value != nil {
					md5 = string(value)
				}
				if value := b.Get([]byte("sha256")); value != nil {
					sha256 = string(value)
				}
			}
		}
		return nil
	})
	return md5, sha256
}

func Info(id string) map[string]string {
	list := make(map[string]string)
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket).Bucket([]byte(id)); b != nil {
			b.ForEach(func(k, v []byte) error {
				list[string(k)] = string(v)
				return nil
			})
			if hash := b.Bucket([]byte("hash")); hash != nil {
				list["md5"] = string(hash.Get([]byte("md5")))
				list["sha256"] = string(hash.Get([]byte("sha256")))
			}
		}
		return nil
	})
	list["id"] = id
	return list
}

func Close() {
	db.Close()
}

func Search(query string) (list []string) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(search)
		b.ForEach(func(k, v []byte) error {
			if strings.Contains(strings.ToLower(string(k)), strings.ToLower(query)) {
				// for k, _ := c.Seek([]byte(query)); len(k) > 0 && bytes.HasPrefix(k, []byte(query)); k, _ = c.Next() {
				//Shitty search index contains lots of outdated and invalid records and we must return all of them. Need to fix it.
				b.Bucket(k).ForEach(func(kk, vv []byte) error {
					for _, l := range list {
						if l == string(vv) {
							return nil
						}
					}
					list = append(list, string(vv))
					return nil
				})
				// _, kk := b.Bucket(k).Cursor().First()
			}
			return nil
		})
		return nil
	})
	return
}

func LastHash(name, t string) (hash string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(search).Bucket([]byte(strings.ToLower(name))); b != nil {
			c := b.Cursor()
			for k, v := c.Last(); k != nil; k, v = c.Prev() {
				if CheckRepo("", t, string(v)) > 0 {
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
		b, err := tx.Bucket(users).CreateBucketIfNotExists([]byte(strings.ToLower(string(name))))
		if !log.Check(log.WarnLevel, "Registering user "+string(name), err) {
			b.Put([]byte("key"), key)
			if b, err := b.CreateBucketIfNotExists([]byte("keys")); err == nil {
				b.Put(key, nil)
			}
		}
		return err
	})
}

// UserKeys returns list of users GPG keys
func UserKeys(name string) (keys []string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(users).Bucket([]byte(strings.ToLower(name))); b != nil {
			if k := b.Bucket([]byte("keys")); k != nil {
				return k.ForEach(func(k, v []byte) error { keys = append(keys, string(k)); return nil })
			}
			keys = append(keys, string(b.Get([]byte("key"))))
		}
		return nil
	})
	return keys
}

// UserKey is replaced by UserKeys and left for compatibility. This function should be removed later.
func UserKey(name string) (key string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(users).Bucket([]byte(strings.ToLower(name))); b != nil {
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
		tx.Bucket(authID).Put([]byte(token), []byte(name))
		return nil
	})
}

func CheckAuthID(token string) (name string) {
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(authID)
		if value := b.Get([]byte(token)); value != nil {
			name = string(value)
			b.Delete([]byte(token))
		}
		return nil
	})
	return name
}

// FileField provides list of file properties
func FileField(hash, field string) (list []string) {
	list = []string{}
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket).Bucket([]byte(hash)); b != nil {
			if f := b.Bucket([]byte(field)); f != nil {
				f.ForEach(func(k, v []byte) error {
					list = append(list, string(k))
					return nil
				})
			} else if b.Get([]byte(field)) != nil {
				list = append(list, string(b.Get([]byte(field))))
			}
		}
		return nil
	})
	return list
}

// FileSignatures returns map with file owners and theirs signatures
func FileSignatures(hash string, opt ...string) (list map[string]string) {
	//workaround to hide signatures from full list of artifacts and to show it only when certain name is specified
	if len(opt[0]) == 0 {
		return nil
	}
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
		owner = "subutai"
	}
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(users).Bucket([]byte(owner)); b != nil {
			if files := b.Bucket([]byte("files")); files != nil {
				files.ForEach(func(k, v []byte) error {
					if Read(string(k)) == file {
						list = append(list, string(k))
					}
					return nil
				})
			}
		}
		return nil
	})
	return list
}

// GetScope shows users with whom shared a certain owner of the file
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

// ShareWith adds user to share scope of file
func ShareWith(hash, owner, user string) {
	db.Update(func(tx *bolt.Tx) error {
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

// UnshareWith removes user from share scope of file
func UnshareWith(hash, owner, user string) {
	db.Update(func(tx *bolt.Tx) error {
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

// CheckShare returns true if user has access to file, otherwise - false
func CheckShare(hash, user string) (shared bool) {
	// log.Warn("hash: " + hash + ", user: " + user)
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket).Bucket([]byte(hash)); b != nil {
			if b := b.Bucket([]byte("scope")); b != nil {
				b.ForEach(func(k, v []byte) error {
					// log.Warn("Owner: " + string(k))
					if strings.EqualFold(string(k), user) {
						shared = true
					} else if b := b.Bucket(k); b != nil {
						b.ForEach(func(k1, v1 []byte) error {
							// log.Warn("+++" + string(k1))
							if strings.EqualFold(string(k1), user) {
								shared = true
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
	return
}

// Public returns true if file is publicly accessible
func Public(hash string) (public bool) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket).Bucket([]byte(hash)); b != nil {
			if b := b.Bucket([]byte("scope")); b != nil {
				b.ForEach(func(k, v []byte) error {
					if b := b.Bucket(k); b != nil {
						k, _ := b.Cursor().First()
						if k == nil {
							public = true
						}
					}
					return nil
				})
			} else {
				public = true
			}
		}
		return nil
	})
	return
}

// countTotal counts and sets user's total quota usage
func countTotal(user string) (total int) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(users).Bucket([]byte(user)); b != nil {
			if c := b.Bucket([]byte("files")); c != nil {
				c.ForEach(func(k, v []byte) error {
					tmp, _ := strconv.Atoi(Info(string(k))["size"])
					total += tmp
					return nil
				})
			}
		}
		return nil
	})
	return
}

// QuotaLeft returns user's quota left space
func QuotaLeft(user string) int {
	var quota, stored int
	db.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket(users).Bucket([]byte(user)); b != nil {
			if q := b.Get([]byte("quota")); q != nil {
				quota, _ = strconv.Atoi(string(q))
			} else {
				quota = config.DefaultQuota()
			}
			if s := b.Get([]byte("stored")); s != nil {
				stored, _ = strconv.Atoi(string(s))
			} else {
				stored = countTotal(user)
				b.Put([]byte("stored"), []byte(strconv.Itoa(stored)))
			}
		}
		return nil
	})
	if quota == -1 {
		return -1
	} else if quota <= stored {
		return 0
	}
	return quota - stored
}

// QuotaUsageSet accepts size of added/removed file and updates quota usage for user
func QuotaUsageSet(user string, value int) {
	var stored int
	db.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket(users).Bucket([]byte(user)); b != nil {
			if s := b.Get([]byte("stored")); s != nil {
				stored, _ = strconv.Atoi(string(s))
			} else {
				stored = countTotal(user)
			}
			b.Put([]byte("stored"), []byte(strconv.Itoa(stored+value)))
		}
		return nil
	})
}

// QuotaGet returns value of user's disk quota
func QuotaGet(user string) (quota int) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(users).Bucket([]byte(user)); b != nil {
			if q := b.Get([]byte("quota")); q != nil {
				quota, _ = strconv.Atoi(string(q))
			} else {
				quota = config.DefaultQuota()
			}
		}
		return nil
	})
	return
}

// QuotaSet sets changes default storage quota for user
func QuotaSet(user, quota string) {
	db.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket(users).Bucket([]byte(user)); b != nil {
			b.Put([]byte("quota"), []byte(quota))
		}
		return nil
	})
}

// QuotaUsageCorrect updates saved values of quota usage according to file index table
func QuotaUsageCorrect() {
	db.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket(users); b != nil {
			b.ForEach(func(k, v []byte) error {
				if c := b.Bucket(k); c != nil {
					rVal := countTotal(string(k))
					if sVal := c.Get([]byte("stored")); sVal == nil && rVal != 0 || sVal != nil && string(sVal) != strconv.Itoa(rVal) {
						log.Info("Correcting quota usage for user " + string(k))
						log.Info("Stored value: " + string(sVal) + ", real value: " + strconv.Itoa(rVal))
						c.Put([]byte("stored"), []byte(strconv.Itoa(rVal)))
					}
				}
				return nil
			})
		}
		return nil
	})
}

// QuotaUsageGet returns value of used disk quota
func QuotaUsageGet(user string) (stored int) {
	db.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket(users).Bucket([]byte(user)); b != nil {
			if s := b.Get([]byte("stored")); s != nil {
				stored, _ = strconv.Atoi(string(s))
			} else {
				stored = countTotal(user)
				b.Put([]byte("stored"), []byte(strconv.Itoa(stored)))
			}
		}
		return nil
	})
	return
}

// SaveTorrent saves torrent file for particular template in DB for future usage to prevent regeneration same file again.
func SaveTorrent(hash, torrent []byte) {
	db.Update(func(tx *bolt.Tx) error {
		if b, err := tx.Bucket(bucket).CreateBucketIfNotExists(hash); err == nil {
			b.Put([]byte("torrent"), torrent)
		}
		return nil
	})
}

// Torrent retrieves torrent file for template from DB. If no torrent file found it returns nil.
func Torrent(hash []byte) (val []byte) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket).Bucket(hash); b != nil {
			if value := b.Get([]byte("torrent")); value != nil {
				val = value
			}
		}
		return nil
	})
	return val
}

// CheckRepo walks through specified repo (or all repos, if none is specified) and checks
// if particular file exists and owner is correct. Returns number of found matches
func CheckRepo(owner, repo, hash string) (val int) {
	reps := []string{repo}
	if len(repo) == 0 {
		reps = []string{"apt", "template", "raw"}
	}
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket); b != nil && strings.Contains(hash, "-") {
			b.ForEach(func(k, v []byte) error {
				if b := b.Bucket(k).Bucket([]byte("hash")); b != nil {
					if string(b.Get([]byte("md5"))) == hash {
						val++
					}
				}
				return nil
			})
		}
		if b := tx.Bucket(bucket).Bucket([]byte(hash)); b != nil {
			if c := b.Bucket([]byte("type")); c != nil {
				for _, v := range reps {
					if d := c.Bucket([]byte(v)); d != nil {
						if k, _ := d.Cursor().First(); len(owner) == 0 && k != nil {
							val += d.Stats().KeyN
						} else if d.Get([]byte(owner)) != nil {
							val++
						}
					}
				}
			} else if len(repo) == 0 || len(repo) != 0 && repo == string(b.Get([]byte("type"))) {
				if len(owner) == 0 {
					val = b.Bucket([]byte("owner")).Stats().KeyN
				} else if len(owner) != 0 && b.Bucket([]byte("owner")).Get([]byte(owner)) != nil {
					val++
				}
			}
		}
		return nil
	})
	return val
}

// RemoveTags deletes tag from index bucket and file information.
// It should be executed on every file deletion to keep DB consistant.
func RemoveTags(key, list string) error {
	return db.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket).Bucket([]byte(key)); b != nil {
			if t := b.Bucket([]byte("tags")); t != nil {

				for _, v := range strings.Split(list, ",") {
					tag := []byte(strings.ToLower(strings.TrimSpace(v)))
					if s := tx.Bucket(tags).Bucket(tag); s != nil {
						log.Check(log.DebugLevel, "Removing tag "+string(tag)+" from index bucket", s.Delete(tag))
					}
					log.Check(log.DebugLevel, "Removing tag "+string(tag)+" from file information", t.Delete([]byte(key)))
				}
			}
		}
		return nil
	})
}

// Tag returns a list of artifacts that contains requested tags.
// If no records found list will be empty.
func Tag(query string) (list []string, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(tags).Bucket([]byte(strings.ToLower(query))); b != nil {
			return b.ForEach(func(k, v []byte) error {
				list = append(list, string(k))
				return nil
			})
		}
		return fmt.Errorf("Tag not found")
	})
	return list, err
}
