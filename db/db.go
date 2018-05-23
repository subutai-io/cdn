package db

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bytes"

	"github.com/boltdb/bolt"
	"github.com/subutai-io/agent/log"
	"github.com/subutai-io/cdn/config"
)

var (
	MyBucket    = []byte("MyBucket")
	SearchIndex = []byte("SearchIndex")
	Users       = []byte("Users")
	Tokens      = []byte("Tokens")
	AuthID      = []byte("AuthID")
	Tags        = []byte("Tags")
	db          = InitDB()
)

var (
	publicScope  = []byte("94205120b9aa305d3167085d26735f1b") // MD5 Hash of "public-scope"
	privateScope = []byte("06e3ef83aafe325400bdd4b0321be4ad") // MD5 Hash of "private-scope"
)

// AddShare adds user to share scope of file if the file wasn't shared with him yet
func AddShare(hash, owner, user string) {
	log.Debug(fmt.Sprintf("Sharing %+v's file %+v (filename: %+v) with user %+v", owner, hash, NameByHash(hash), user))
	db.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket(MyBucket).Bucket([]byte(hash)); b != nil {
			if b := b.Bucket([]byte("scope")); b != nil {
				if b.Get([]byte(user)) == nil {
					log.Debug(fmt.Sprintf("Adding %+v to tx.Bucket(%+v).Bucket(%+v).Bucket(%+v)", user, string(MyBucket), hash, "scope"))
					b.Put([]byte(user), []byte("w"))
				}
			}
			if e := tx.Bucket(Users).Bucket([]byte(user)); e != nil {
				if f, _ := e.CreateBucketIfNotExists([]byte("files")); f != nil {
					if f.Get([]byte(hash)) == nil {
						log.Debug(fmt.Sprintf("Putting %+v's file %+v (filename: %+v) to %+v's files Bucket", owner, hash, NameByHash(hash), user))
						f.Put([]byte(hash), []byte(NameByHash(hash)))
					} else {
						log.Debug(fmt.Sprintf("File already put in files Bucket"))
					}
				}
			}
		}
		return nil
	})
	log.Debug(fmt.Sprintf("Sharing finished"))
}

func CheckAuthID(token string) (name string) {
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(AuthID)
		if value := b.Get([]byte(token)); value != nil {
			name = string(value)
			b.Delete([]byte(token))
		}
		return nil
	})
	return
}

func CheckRepo(owner string, repo []string, hash string) (val int) {
	log.Debug(fmt.Sprintf("CheckRepo (repo: \"%+v (len: %+v)\", owner: \"%+v\", file: \"%+v\" (name: %+v))", repo, len(repo), owner, hash, NameByHash(hash)))
	if len(repo) == 0 {
		repo = []string{"apt", "template", "raw"}
		log.Debug(fmt.Sprintf("Provided empty repo. New repo: %+v", repo))
	}
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(MyBucket).Bucket([]byte(hash)); b != nil {
			if len(owner) > 0 && !CheckShare(hash, owner) {
				log.Debug(fmt.Sprintf("File %+v (name: %+v) doesn't belong to and is not shared with %+v", hash, NameByHash(hash), owner))
				return nil
			}
			if c := b.Bucket([]byte("type")); c != nil {
				for _, v := range repo {
					log.Debug(fmt.Sprintf("Checking file %+v (name: %+v) in repo %+v", hash, NameByHash(hash), v))
					if d := c.Bucket([]byte(v)); d != nil {
						log.Debug(fmt.Sprintf("File %+v (name: %+v, owner: %+v) found in repo %+v", hash, NameByHash(hash), owner, v))
						val++
					} else {
						log.Debug(fmt.Sprintf("File %+v (name: %+v) is not in repo %+v", hash, NameByHash(hash), v))
					}
				}
			}
		}
		return nil
	})
	log.Debug(fmt.Sprintf("Found %+v matches", val))
	return
}

// CheckShare returns true if user has access to file, otherwise - false
func CheckShare(hash, user string) (shared bool) {
	log.Debug(fmt.Sprintf("Checking if user %+v has access to file %+v (%+v)", user, hash, NameByHash(hash)))
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(MyBucket).Bucket([]byte(hash)); b != nil {
			if c := b.Bucket([]byte("scope")); c != nil {
				if c.Get(publicScope) != nil {
					shared = true
				} else if c.Get(privateScope) != nil {
					//					log.Debug(fmt.Sprintf("Iterating through tx.Bucket(%+v).Bucket(%+v).Bucket(%+v)", string(MyBucket), hash, "scope"))
					if c.Get([]byte(user)) != nil || b.Bucket([]byte("owner")).Get([]byte(user)) != nil {
						shared = true
					}
				} else {
					log.Debug(fmt.Sprintf("Availability scopes are missing"))
					if d := b.Bucket([]byte("owner")); d != nil {
						owner, _ := d.Cursor().First()
						RebuildShare(hash, string(owner))
						shared = CheckShare(hash, user)
					} else {
						log.Panic(fmt.Sprintf("Cannot find compromise: line 123, info_test.go"))
					}
				}
			}
		}
		return nil
	})
	log.Debug(fmt.Sprintf("Verdict: %+v", shared))
	return
}

func Close() {
	db.Close()
}

// Count all artifacts that have MD5 equal to hash
func CountMd5(hash string) (md5 int) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(MyBucket); b != nil {
			b.ForEach(func(k, v []byte) error {
				if b := b.Bucket(k).Bucket([]byte("hash")); b != nil {
					if string(b.Get([]byte("md5"))) == hash {
						md5++
					}
				}
				return nil
			})
		}
		return nil
	})
	return
}

// CountTotal counts and sets user's total quota usage
func CountTotal(user string) (total int) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(Users).Bucket([]byte(user)); b != nil {
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

func DebugDatabase() {
	//	log.Debug(fmt.Sprintf("\ndb.GoString():\n%+v\n", db.GoString()))
	//	log.Debug(fmt.Sprintf("\ndb.Stats():\n%+v\n", db.Stats()))
	//	log.Debug(fmt.Sprintf("\ndb.Info():\n%+v\n", db.Info()))
	db.View(func(tx *bolt.Tx) error {
		tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			log.Debug(fmt.Sprintf("\nPrinting tx:\n(name: %+v, b: %+v)\n", string(name), b))
			PrintBuckets(b, []string{string(name)})
			return nil
		})
		return nil
	})
}

// Delete removes record about file from DB*2

func Delete(owner, repo, key string) (total int) {
	db.Update(func(tx *bolt.Tx) error {
		var filename []byte
		owned := CheckRepo(owner, []string{}, key)
		md5, _ := Hash(key)
		total = CheckRepo("", []string{}, key)
		// Deleting user association with file
		if b := tx.Bucket(MyBucket).Bucket([]byte(key)); b != nil {
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
		if b := tx.Bucket(Users).Bucket([]byte(owner)); owned == 1 && b != nil {
			if b := b.Bucket([]byte("files")); b != nil {
				filename = b.Get([]byte(key))
				b.Delete([]byte(key))
			}
		}
		// Removing indexes and file only if no file owners left
		if total == 1 || key != md5 {
			// Deleting SearchIndex index
			if b := tx.Bucket(SearchIndex).Bucket(bytes.ToLower(filename)); b != nil {
				b.ForEach(func(k, v []byte) error {
					if string(v) == key {
						b.Delete(k)
					}
					return nil
				})
			}
			for _, tag := range FileField(key, "tags") {
				if s := tx.Bucket(Tags).Bucket([]byte(tag)); s != nil {
					log.Check(log.DebugLevel, "Removing tag "+tag+" from index MyBucket", s.Delete([]byte(key)))
				}
			}
			// Removing file from DB
			tx.Bucket(MyBucket).DeleteBucket([]byte(key))
		}
		return nil
	})
	return
}

// Edit record about file in DB
func Edit(owner, key, value string, options ...map[string]string) {
	if len(owner) == 0 {
		owner = "subutai"
	}
	err := db.Update(func(tx *bolt.Tx) error {
		//		Associating files with user
		b, _ := tx.Bucket(Users).CreateBucketIfNotExists([]byte(owner))
		if b, err := b.CreateBucketIfNotExists([]byte("files")); err == nil {
			if v := b.Get([]byte(key)); v == nil {
				//				log.Info("Associating: " + owner + " with " + value + " (" + key + ")")
				b.Put([]byte(key), []byte(value))
			}
		}
		//		Editing record about file
		if len(value) > 0 {
			b.Put([]byte("name"), []byte(value))
		}
		//		Editing owners, shares and tags to files
		if b := tx.Bucket(MyBucket).Bucket([]byte(key)); b != nil {
			if c := b.Bucket([]byte("owner")); len(owner) > 0 {
				if c.Get([]byte(owner)) == nil {
					c.Put([]byte(owner), []byte("w"))
				}
			}
			for i := range options {
				for k, v := range options[i] {
					switch k {
					case "type":
						if c := b.Bucket([]byte("type")); c != nil {
							if c := c.Bucket([]byte(v)); len(owner) > 0 {
								if c.Get([]byte(owner)) == nil {
									c.Put([]byte(owner), []byte("w"))
								}
							}
						}
					case "md5", "sha256":
						if c := b.Bucket([]byte("hash")); len(k) > 0 {
							c.Put([]byte(k), []byte(v))
							// Getting file size
							if f, err := os.Open(config.Storage.Path + v); err == nil {
								fi, _ := f.Stat()
								f.Close()
								b.Put([]byte("size"), []byte(fmt.Sprint(fi.Size())))
							}
						}
					case "tags":
						if c := b.Bucket([]byte("tags")); len(v) > 0 {
							for _, v := range strings.Split(v, ",") {
								tag := []byte(strings.ToLower(strings.TrimSpace(v)))
								t, _ := tx.Bucket(Tags).CreateBucketIfNotExists(tag)
								c.Put(tag, []byte("w"))
								t.Put([]byte(key), []byte("w"))
							}
						}
					case "signature":
						if c := b.Bucket([]byte("owner")); len(v) > 0 {
							c.Put([]byte(owner), []byte(v))
						}
					default:
						{
							b.Put([]byte(k), []byte(v))
						}
					}
				}
			}
			if b = b.Bucket([]byte("scope")); b != nil {
				if b = b.Bucket([]byte(owner)); b != nil {
				}
			}
		}
		return nil
	})
	log.Check(log.WarnLevel, "Editing data in db", err)
}

// FileField provides list of file's field properties
func FileField(hash, field string) (list []string) {
	log.Debug(fmt.Sprintf("FileField: providing field %+v for file %+v", field, NameByHash(hash)))
	list = []string{}
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(MyBucket).Bucket([]byte(hash)); b != nil {
			if f := b.Bucket([]byte(field)); f != nil {
				//				log.Debug(fmt.Sprintf("Iterating through MyBucket tx.Bucket(%+v).Bucket(%+v).Bucket(%+v)", string(MyBucket), NameByHash(hash), field))
				f.ForEach(func(k, v []byte) error {
					//					log.Debug(fmt.Sprintf("Appending (key: %+v, value: %+v)", string(k), string(v)))
					list = append(list, string(k))
					return nil
				})
			} else if b.Get([]byte(field)) != nil {
				//				log.Debug(fmt.Sprintf("Bucket %+v does not exist. Executing tx.Bucket(%+v).Bucket(%+v).Get(%+v)", string(MyBucket), NameByHash(hash), field))
				list = append(list, string(b.Get([]byte(field))))
			}
		} else {
			//			log.Debug(fmt.Sprintf("File with hash %+v does not exist", hash))
		}
		return nil
	})
	log.Debug(fmt.Sprintf("Field %+v for file %+v:\n%+v", field, NameByHash(hash), list))
	return
}

// FileSignatures returns map with file's owners and their signatures
func FileSignatures(hash string) (list map[string]string) {
	log.Debug(fmt.Sprintf("Gathering owners and their signatures of file %+v", NameByHash(hash)))
	list = map[string]string{}
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(MyBucket).Bucket([]byte(hash)); b != nil {
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
	log.Debug(fmt.Sprintf("Owners and their signatures: %+v", list))
	return
}

// GetFileScope shows users with whom owner shared a file with particular hash
func GetFileScope(hash, owner string) (scope []string) {
	scope = []string{}
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(MyBucket).Bucket([]byte(hash)); b != nil {
			if b := b.Bucket([]byte("scope")); b != nil {
				if b.Get(publicScope) == nil && b.Get(privateScope) == nil {
					RebuildShare(hash, owner)
					scope = GetFileScope(hash, owner)
				} else {
					b.ForEach(func(k, v []byte) error {
						if string(k) != string(publicScope) && string(k) != string(privateScope) {
							scope = append(scope, string(k))
						}
						return nil
					})
				}
			}
		}
		return nil
	})
	return
}

func GetUserToken(user string) (token string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(Tokens); b != nil {
			b.ForEach(func(k, v []byte) error {
				log.Debug(fmt.Sprintf("(GetUserToken): Current token %s", string(k)))
				if c := b.Bucket(k); c != nil {
					date := new(time.Time)
					date.UnmarshalText(c.Get([]byte("date")))
					if date.Add(time.Hour * 24).Before(time.Now()) {
						return nil
					}
					if value := c.Get([]byte("name")); value != nil && string(value) == user {
						token = string(k)
					}
				}
				return nil
			})
		}
		return nil
	})
	if token != "" {
		log.Debug(fmt.Sprintf("(GetUserToken): Found %s's valid token %s", user, token))
	} else {
		log.Debug(fmt.Sprintf("(GetUserToken): Couldn't find valid token of %s", user))
	}
	return
}

// Hash returns MD5 and SHA256 hashes by ID
func Hash(key string) (md5, sha256 string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(MyBucket).Bucket([]byte(key)); b != nil {
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
	return
}

func Info(id string) map[string]string {
	log.Debug(fmt.Sprintf("\n\nGathering %+v file's info", NameByHash(id)))
	list := make(map[string]string)
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(MyBucket).Bucket([]byte(id)); b != nil {
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
	if len(list) != 0 {
		list["id"] = id
	}
	log.Debug(fmt.Sprintf("\n\nGathered %+v file's info. list: %+v\n\n", NameByHash(id), list))
	return list
}

func InitDB() *bolt.DB {
	os.MkdirAll(filepath.Dir(config.DB.Path), 0755)
	os.MkdirAll(config.Storage.Path, 0755)
	db, err := bolt.Open(config.DB.Path, 0600, &bolt.Options{Timeout: 3 * time.Second})
	log.Check(log.FatalLevel, "Opening DB: "+config.DB.Path, err)
	err = db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{MyBucket, SearchIndex, Users, Tokens, AuthID, Tags} {
			_, err := tx.CreateBucketIfNotExists(b)
			log.Check(log.FatalLevel, "Creating bucket: "+string(b), err)
		}
		return nil
	})
	log.Check(log.FatalLevel, "Finishing update transaction", err)
	return db
}

// IsPublic returns true if file is publicly accessible
func IsPublic(hash string) (public bool) {
	log.Debug(fmt.Sprintf("Checking if file %+v (hash: %+v) is public", NameByHash(hash), hash))
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(MyBucket).Bucket([]byte(hash)); b != nil {
			if c := b.Bucket([]byte("scope")); c != nil {
				if c.Get(publicScope) == nil && c.Get(privateScope) == nil {
					log.Debug(fmt.Sprintf("Both public and private scopes are missing. Rebuilding scope"))
					if d := b.Bucket([]byte("owner")); d != nil {
						owner, _ := d.Cursor().First()
						RebuildShare(hash, string(owner))
					} else {
						log.Panic(fmt.Sprintf("The file %+v has no owner", hash))
					}
				} else {
					log.Debug(fmt.Sprintf("Checking %+v (name: %+v) file's scope for publicScope", hash, NameByHash(hash)))
					public = c.Get(publicScope) != nil
				}
			}
		}
		return nil
	})
	log.Debug(fmt.Sprintf("Verdict: %+v", public))
	return
}

// LastHash returns hash of the last uploaded file
func LastHash(name, t string) (hash string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(SearchIndex).Bucket([]byte(strings.ToLower(name))); b != nil {
			c := b.Cursor()
			for k, v := c.Last(); k != nil; k, v = c.Prev() {
				if CheckRepo("", []string{t}, string(v)) > 0 {
					hash = string(v)
					break
				}
			}
		}
		return nil
	})
	return
}

func MakePublic(hash, owner string) {
	log.Debug(fmt.Sprintf("MakePublic(%+v, %+v) started", hash, owner))
	RemoveShare(hash, owner, string(privateScope))
	AddShare(hash, owner, string(publicScope))
	log.Debug(fmt.Sprintf("MakePublic(%+v, %+v) ended", hash, owner))
}

func MakePrivate(hash, owner string) {
	log.Debug(fmt.Sprintf("MakePrivate(%+v, %+v) started", hash, owner))
	RemoveShare(hash, owner, string(publicScope))
	AddShare(hash, owner, string(privateScope))
	log.Debug(fmt.Sprintf("MakePrivate(%+v, %+v) ended", hash, owner))
}

// NameByHash returns file's name by its ID
func NameByHash(hash string) (name string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(MyBucket).Bucket([]byte(hash)); b != nil {
			if value := b.Get([]byte("name")); value != nil {
				name = string(value)
			}
		}
		return nil
	})
	return
}

// OwnerFilesByRepo returns all public files of owner from specified repo
func OwnerFilesByRepo(owner string, repo string) (list []string) {
	log.Debug(fmt.Sprintf("(OwnerFilesByRepo): Gathering all %+v's files from repo %+v...", owner, repo))
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(Users).Bucket([]byte(owner)); b != nil {
			if files := b.Bucket([]byte("files")); files != nil {
				files.ForEach(func(k, v []byte) error {
					log.Debug(fmt.Sprintf("(OwnerFilesByRepo): File of %+v -- (key: %v, value: %v)", owner, string(k), string(v)))
					if IsPublic(string(k)) {
						if /* CheckShare(string(k), owner) */ filesNumber := CheckRepo(owner, []string{repo}, string(k)); filesNumber > 0 {
							log.Debug(fmt.Sprintf("(OwnerFilesByRepo): Found %+v files of %+v with key %v in repo %v", filesNumber, owner, string(k), repo))
							list = append(list, string(k))
						}
					}
					return nil
				})
			}
		}
		return nil
	})
	log.Debug(fmt.Sprintf("(OwnerFilesByRepo): list of all %+v's files from repo %+v: %+v", owner, repo, list))
	return
}

func PrintBucketName(buckets []string) (path string) {
	path = "tx"
	for i := range buckets {
		path += fmt.Sprintf(".Bucket(%+v)", buckets[i])
	}
	return path
}

func PrintBuckets(b *bolt.Bucket, parents []string) {
	log.Debug(fmt.Sprintf("	Printing %+v", PrintBucketName(parents)))
	b.ForEach(func(k, v []byte) error {
		log.Debug(fmt.Sprintf("		(k: %+v, v: %+v)", string(k), string(v)))
		return nil
	})
	b.ForEach(func(k, v []byte) error {
		nb := b.Bucket(k)
		if nb != nil {
			log.Debug(fmt.Sprintf("\n"))
			PrintBuckets(nb, append(parents, string(k)))
		}
		return nil
	})
}

// QuotaGet returns value of user's disk quota
func QuotaGet(user string) (quota int) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(Users).Bucket([]byte(user)); b != nil {
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

// QuotaLeft returns user's quota left space
func QuotaLeft(user string) int {
	var quota, stored int
	db.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket(Users).Bucket([]byte(user)); b != nil {
			if q := b.Get([]byte("quota")); q != nil {
				quota, _ = strconv.Atoi(string(q))
			} else {
				quota = config.DefaultQuota()
			}
			if s := b.Get([]byte("stored")); s != nil {
				stored, _ = strconv.Atoi(string(s))
			} else {
				stored = CountTotal(user)
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

// QuotaSet sets changes default storage quota for user
func QuotaSet(user, quota string) {
	db.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket(Users).Bucket([]byte(user)); b != nil {
			b.Put([]byte("quota"), []byte(quota))
		}
		return nil
	})
}

// QuotaUsageCorrect updates saved values of quota usage according to file index table
func QuotaUsageCorrect() {
	db.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket(Users); b != nil {
			b.ForEach(func(k, v []byte) error {
				if c := b.Bucket(k); c != nil {
					rVal := CountTotal(string(k))
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
		if b := tx.Bucket(Users).Bucket([]byte(user)); b != nil {
			if s := b.Get([]byte("stored")); s != nil {
				stored, _ = strconv.Atoi(string(s))
			} else {
				stored = CountTotal(user)
				b.Put([]byte("stored"), []byte(strconv.Itoa(stored)))
			}
		}
		return nil
	})
	return
}

// QuotaUsageSet accepts size of added/removed file and updates quota usage for user
func QuotaUsageSet(user string, value int) {
	var stored int
	db.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket(Users).Bucket([]byte(user)); b != nil {
			if s := b.Get([]byte("stored")); s != nil {
				stored, _ = strconv.Atoi(string(s))
			} else {
				stored = CountTotal(user)
			}
			b.Put([]byte("stored"), []byte(strconv.Itoa(stored+value)))
		}
		return nil
	})
}

func RebuildShare(hash, owner string) {
	log.Debug(fmt.Sprintf("RebuildShare(%+v, %+v) started", hash, owner))
	public := true
	db.Update(func(tx *bolt.Tx) error {
		log.Debug(fmt.Sprintf("Starting RebuildShare"))
		if b := tx.Bucket(MyBucket).Bucket([]byte(hash)); b != nil {
			/*			name := ""
						if value := b.Get([]byte("name")); value != nil {
							name = string(value)
						}
			*/if c := b.Bucket([]byte("scope")); c != nil {
				keyList := [][]byte{}
				valueList := [][]byte{}
				bucketList := [][]byte{}
				c.ForEach(func(k, v []byte) error {
					log.Debug(fmt.Sprintf("Iterating through scope: (key: %+v, value: %+v)", string(k), string(v)))
					if v == nil {
						d := c.Bucket(k)
						log.Debug(fmt.Sprintf("key: %+v is a Bucket", string(k)))
						log.Debug(fmt.Sprintf("Iterating through this Bucket:"))
						d.ForEach(func(kk, vv []byte) error {
							log.Debug(fmt.Sprintf("	(subkey: %+v, subvalue: %+v)", string(kk), string(vv)))
							if string(kk) == string(k) {
								log.Debug(fmt.Sprintf("Self-share detected. File is private"))
								public = false
							} else {
								keyList = append(keyList, kk)
								valueList = append(valueList, vv)
								/*if e := tx.Bucket(Users).Bucket(kk); e != nil {
									if f := e.Bucket([]byte("files")); f != nil {
										if f.Get([]byte(hash)) == nil {
											log.Debug(fmt.Sprintf("		Putting %+v's file %+v (name: %+v, actual name: %+v) in files of %+v", owner, hash, name, NameByHash(hash), string(kk)))
											f.Put([]byte(hash), []byte(name))
										}
									}
								}*/
							}
							return nil
						})
						bucketList = append(bucketList, k)
					}
					return nil
				})
				for _, k := range bucketList {
					c.DeleteBucket(k)
				}
				for i := 0; i < len(keyList); i++ {
					if c.Get(keyList[i]) == nil {
						c.Put(keyList[i], valueList[i])
					}
				}
			}
		}
		return nil
	})
	log.Debug(fmt.Sprintf("Scope rebuild almost finished. Setting file's availability: %+v", public))
	if public {
		MakePublic(hash, owner)
	} else {
		MakePrivate(hash, owner)
	}
	log.Debug(fmt.Sprintf("RebuildShare(%+v, %+v) ended", hash, owner))
}

func RegisterUser(name, key []byte) {
	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.Bucket(Users).CreateBucketIfNotExists([]byte(strings.ToLower(string(name))))
		if !log.Check(log.WarnLevel, "Registering user " + strings.ToLower(string(name)), err) {
			b.Put([]byte("key"), key)
			if b, err := b.CreateBucketIfNotExists([]byte("keys")); err == nil {
				log.Debug(fmt.Sprintf("Created user %+v", name))
				b.Put(key, nil)
			}
		}
		return err
	})
}

// RemoveShare removes user from share scope of file if the file was shared with him
func RemoveShare(hash, owner, user string) {
	log.Debug(fmt.Sprintf("RemoveShare(%+v, %+v, %+v) started", hash, owner, user))
	db.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket(MyBucket).Bucket([]byte(hash)); b != nil {
			if b := b.Bucket([]byte("scope")); b != nil {
				if b.Get([]byte(user)) != nil {
					b.Delete([]byte(user))
				}
				if e := tx.Bucket(Users).Bucket([]byte(user)); e != nil {
					if f := e.Bucket([]byte("files")); f != nil {
						if f.Get([]byte(hash)) != nil {
							f.Delete([]byte(hash))
						}
					}
				}
			}
		}
		return nil
	})
	log.Debug(fmt.Sprintf("RemoveShare(%+v, %+v, %+v) ended", hash, owner, user))
}

// RemoveTags deletes tag from index bucket and file information.
// It should be executed on every file deletion to keep DB consistant.
func RemoveTags(key, list string) error {
	return db.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket(MyBucket).Bucket([]byte(key)); b != nil {
			if t := b.Bucket([]byte("tags")); t != nil {
				for _, v := range strings.Split(list, ",") {
					tag := []byte(strings.ToLower(strings.TrimSpace(v)))
					if s := tx.Bucket(Tags).Bucket(tag); s != nil {
						log.Check(log.DebugLevel, "Removing tag "+string(tag)+" from index MyBucket", s.Delete(tag))
					}
					log.Check(log.DebugLevel, "Removing tag "+string(tag)+" from file information", t.Delete([]byte(key)))
				}
			}
		}
		return nil
	})
}

func SaveAuthID(name, token string) {
	db.Update(func(tx *bolt.Tx) error {
		tx.Bucket(AuthID).Put([]byte(token), []byte(name))
		return nil
	})
}

func SaveToken(name, token string) {
	db.Update(func(tx *bolt.Tx) error {
		if b, _ := tx.Bucket(Tokens).CreateBucketIfNotExists([]byte(token)); b != nil {
			b.Put([]byte("name"), []byte(name))
			now, _ := time.Now().MarshalText()
			b.Put([]byte("date"), now)
		}
		return nil
	})
}

// SaveTorrent saves torrent file for particular template in DB for future usage to prevent regeneration same file again.
func SaveTorrent(hash, torrent []byte) {
	db.Update(func(tx *bolt.Tx) error {
		if b, err := tx.Bucket(MyBucket).CreateBucketIfNotExists(hash); err == nil {
			b.Put([]byte("torrent"), torrent)
		}
		return nil
	})
}

// SearchName searches for all (public/private) files of all users that have "query" substring in their names
func SearchName(query string) (list []string) {
	log.Debug(fmt.Sprintf("Starting db.SearchName(%+v)", query))
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(SearchIndex)
		b.ForEach(func(k, v []byte) error {
			//			log.Debug(fmt.Sprintf("tx.Bucket(%+v).ForEach(key: %+v, value: %+v)", string(SearchIndex), string(k), string(v)))
			if strings.Contains(strings.ToLower(string(k)), strings.ToLower(query)) {
				b.Bucket(k).ForEach(func(kk, vv []byte) error {
					//					log.Debug(fmt.Sprintf("b.Bucket(%+v).ForEach(key: %+v, value: %+v)", string(k), string(kk), string(vv)))
					for _, l := range list {
						if l == string(vv) {
							return nil
						}
					}
					list = append(list, string(vv))
					//					log.Debug(fmt.Sprintf("%+v not found in list. list after appending: %+v", string(vv), list))
					return nil
				})
			}
			return nil
		})
		return nil
	})
	log.Debug(fmt.Sprintf("list of public/private files containing substring \"%+v\":\n\n\n%+v", query, list))
	for _, f := range list {
		log.Debug(fmt.Sprintf("		%+v (name: %+v)", f, NameByHash(f)))
	}
	return
}

// Tag returns a list of artifacts that contains requested tags.
// If no records found list will be empty.
func Tag(query string) (list []string, err error) {
	log.Debug(fmt.Sprintf("Starting db.Tag(%+v)", query))
	err = db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(Tags).Bucket([]byte(strings.ToLower(query))); b != nil {
			return b.ForEach(func(k, v []byte) error {
				list = append(list, string(k))
				log.Debug(fmt.Sprintf("tx.Bucket(%+v).ForEach(key: %+v, value: %+v) -- contains tag %+v", string(Tags), string(k), string(v), query))
				return nil
			})
		}
		return fmt.Errorf("Tag not found")
	})
	log.Debug(fmt.Sprintf("list: %+v\nerr: %+v", list, err))
	return
}

// TokenOwner returns the owner of the given token
func TokenOwner(token string) (name string) {
	tokenFormatted := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(Tokens).Bucket([]byte(tokenFormatted)); b != nil {
			date := new(time.Time)
			date.UnmarshalText(b.Get([]byte("date")))
			if date.Add(time.Hour * 24).Before(time.Now()) {
				return nil
			}
			if value := b.Get([]byte("name")); value != nil {
				name = string(value)
			}
		} else if b := tx.Bucket(Tokens).Bucket([]byte(token)); b != nil {
			date := new(time.Time)
			date.UnmarshalText(b.Get([]byte("date")))
			if date.Add(time.Hour * 24).Before(time.Now()) {
				return nil
			}
			if value := b.Get([]byte("name")); value != nil {
				name = string(value)
			}
		} else {
			log.Debug(fmt.Sprintf("Token %s (formatted: %s) not found", token, tokenFormatted))
		}
		return nil
	})
	log.Debug(fmt.Sprintf("Checking token %s (formatted: %s) finished. Token corresponds to name %s", token, tokenFormatted, name))
	return
}

// TokenFilesByRepo returns all public/private/shared files of token owner from specified repo
func TokenFilesByRepo(token string, repo string) (list []string) {
	owner := TokenOwner(token)
	if owner == "" {
		log.Debug(fmt.Sprintf("(TokenFilesByRepo): Couldn't find owner of token %s", token))
		return
	}
	log.Debug(fmt.Sprintf("(TokenFilesByRepo): Gathering all %+v's files from repo %+v...", owner, repo))
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(Users).Bucket([]byte(owner)); b != nil {
			if files := b.Bucket([]byte("files")); files != nil {
				files.ForEach(func(k, v []byte) error {
					if filesNumber := CheckRepo(owner, []string{repo}, string(k)); filesNumber > 0 {
						list = append(list, string(k))
					}
					return nil
				})
			}
		}
		return nil
	})
	log.Debug(fmt.Sprintf("(TokenFilesByRepo): list of all %+v's files from repo %+v: %+v", owner, repo, list))
	return
}

// Torrent retrieves torrent file for template from DB. If no torrent file found it returns nil.
func Torrent(hash []byte) (val []byte) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(MyBucket).Bucket(hash); b != nil {
			if value := b.Get([]byte("torrent")); value != nil {
				val = value
			}
		}
		return nil
	})
	return
}

// UserFile searches a file among particular user's files. It returns list of hashes of files with required name.
func UserFile(owner, file string) (list []string) {
	if len(owner) == 0 {
		owner = "subutai"
	}
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(Users).Bucket([]byte(owner)); b != nil {
			if files := b.Bucket([]byte("files")); files != nil {
				files.ForEach(func(k, v []byte) error {
					if NameByHash(string(k)) == file {
						list = append(list, string(k))
					}
					return nil
				})
			}
		}
		return nil
	})
	return
}

// UserKey is replaced by UserKeys and left for compatibility. This function should be removed later.
func UserKey(name string) (key string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(Users).Bucket([]byte(strings.ToLower(name))); b != nil {
			if value := b.Get([]byte("key")); value != nil {
				key = string(value)
			}
		}
		return nil
	})
	return
}

// UserKeys returns list of users' GPG keys
func UserKeys(name string) (keys []string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(Users).Bucket([]byte(strings.ToLower(name))); b != nil {
			if k := b.Bucket([]byte("keys")); k != nil {
				return k.ForEach(func(k, v []byte) error {
					keys = append(keys, string(k))
					return nil
				})
			}
			keys = append(keys, string(b.Get([]byte("key"))))
		}
		return nil
	})
	return
}

// Write create record about file in DB
func Write(owner, key, value string, options ...map[string]string) error {
	if len(owner) == 0 {
		owner = "subutai"
	}
	err := db.Update(func(tx *bolt.Tx) error {
		now, _ := time.Now().MarshalText()
		// Associating files with user
		b, _ := tx.Bucket(Users).CreateBucketIfNotExists([]byte(owner))
		if b, err := b.CreateBucketIfNotExists([]byte("files")); err == nil {
			if v := b.Get([]byte(key)); v == nil {
				// log.Warn("Associating: " + owner + " with " + value + " (" + key + ")")
				b.Put([]byte(key), []byte(value))
			}
		}
		// Creating new record about file
		if b, err := tx.Bucket(MyBucket).CreateBucket([]byte(key)); err == nil {
			b.Put([]byte("date"), now)
			b.Put([]byte("name"), []byte(value))
			// Adding SearchIndex index for files
			b, _ = tx.Bucket(SearchIndex).CreateBucketIfNotExists([]byte(strings.ToLower(value)))
			b.Put(now, []byte(key))
		}
		// Adding owners, shares and tags to files
		if b := tx.Bucket(MyBucket).Bucket([]byte(key)); b != nil {
			if c, err := b.CreateBucket([]byte("owner")); err == nil {
				log.Info(fmt.Sprintf("Bucket owner created successfully"))
				c.Put([]byte(owner), []byte("w"))
			}
			if _, err := b.CreateBucket([]byte("scope")); err == nil {
				log.Info(fmt.Sprintf("Bucket scope created successfully"))
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
								t, _ := tx.Bucket(Tags).CreateBucketIfNotExists(tag)
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
		}
		return nil
	})
	log.Check(log.WarnLevel, "Writing data to db", err)
	return err
}

// CheckRepoOfHash return the type of file by its hash
func CheckRepoOfHash(hash string) (repo string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(MyBucket).Bucket([]byte(hash)).Bucket([]byte("type")); b != nil {
			b.ForEach(func(k, v []byte) error {
				repo = string(k)
				return nil
			})
		}
		return nil
	})
	return repo
}

// SearchFileByTag performs search file by the specified tag. Return the list of files with such tag.
func SearchFileByTag(tag string, repo string) (listofIds []string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(MyBucket); b != nil {
			b.ForEach(func(k, v []byte) error {
				r := CheckRepoOfHash(string(k))
				t := string(b.Bucket(k).Get([]byte("tag")))
				if r == repo {
					if t == tag {
						listofIds = append(listofIds, string(k))
					} else {
						tt := strings.Split(t, ",")
						for _, s := range tt {
							if s == tag {
								listofIds = append(listofIds, string(k))
							}
						}
					}
				}
				return nil
			})
			return nil
		}
		return nil
	})
	return listofIds
}
