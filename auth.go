package main

import (
	"encoding/base64"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/context"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"strings"
)

type User struct {
	gorm.Model
	Username string `sql:"not null;unique"`
	Password string
}
type Repo struct {
	gorm.Model
	AccountId int
	Name      string `sql:"not null;unique"`
	Type      string `sql:"not null"`
}
type Manifest struct {
	gorm.Model
	Name      string
	Reference string
	Digest    string
	Content   []byte `sql:"size:2000,type:varbinary(2000)"`
}
type Permission struct {
	gorm.Model
	RepoId int
	Repo   *Repo
	UserId int
	User   *User
	Role   string `sql:"not null"`
}

func isAuthorized(userId uint, name string, write bool) bool {
	s := strings.SplitN(name, "/", 2)
	if len(s) != 2 {
		return false
	}
	db, err := gorm.Open("mysql", os.Getenv("MYSQL_CONN"))
	if err != nil {
		log.Println(err.Error())
	}
	defer db.Close()
	perm := []Permission{}
	db.Preload("Repo").Where("user_id=?", int(userId)).Find(&perm)
	for _, p := range perm {
		if p.Repo.Name == s[0] {
			if p.Repo.Type == "public" && !write {
				return true
			}
			if p.Role == "readonly" && write {
				return false
			}
			return true
		}
	}
	return false
}
func basicAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(s) != 2 {
			http.Error(w, "Not authorized", 401)
			return
		}

		b, err := base64.StdEncoding.DecodeString(s[1])
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		pair := strings.SplitN(string(b), ":", 2)
		if len(pair) != 2 {
			http.Error(w, "Not authorized", 401)
			return
		}
		user, err := checkUser(pair[0], pair[1])
		if err != nil {
			http.Error(w, "Not authorized", 401)
			return
		}
		if r.URL.Path == "/v2/" {
			w.WriteHeader(200)
			return
		}
		context.Set(r, "userId", user.ID)
		h.ServeHTTP(w, r)
	}
}
func checkUser(username string, password string) (User, error) {
	db, err := gorm.Open("mysql", os.Getenv("MYSQL_CONN"))
	if err != nil {
		log.Println(err.Error())
	}
	defer db.Close()
	user := User{}
	db.Where("username=?", username).First(&user)
	if user.Username == "" {
		return User{}, errors.New("Username is blank")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return User{}, errors.New("Incorrect Password")
	}
	return user, nil
}
