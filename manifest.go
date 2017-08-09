package main

import (
	"bytes"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func Manifests(w http.ResponseWriter, r *http.Request) {
	db, err := gorm.Open("mysql", os.Getenv("MYSQL_CONN"))
	if err != nil {
		log.Println(err.Error())
	}
	defer db.Close()
	vars := mux.Vars(r)
	if r.Method == "PUT" {
		userId := context.Get(r, "userId")
		if !isAuthorized(userId.(uint), vars["name"], true) {
			w.WriteHeader(403)
			return
		}
		b, _ := ioutil.ReadAll(r.Body)
		hash := createDockerHash(b)
		manifest := Manifest{}
		db.Where("name=? AND reference=?", vars["name"], vars["reference"]).First(&manifest)
		if manifest.Name == "" {
			manifest = Manifest{
				Name:      vars["name"],
				Reference: vars["reference"],
				Digest:    hash,
				Content:   b,
			}
			db.Create(&manifest)
		} else {
			manifest.Digest = hash
			manifest.Content = b
			db.Save(&manifest)
		}
		w.Header().Set("Content-Length", "0")
		w.Header().Set("Docker-Content-Digest", manifest.Digest)
		w.WriteHeader(201)
		if os.Getenv("WEBHOOK") != "" {
			data := url.Values{}
			data.Set("name", vars["name"])
			data.Add("reference", vars["reference"])
			data.Add("action", "uploaded")
			req, err := http.NewRequest("POST", os.Getenv("WEBHOOK"), bytes.NewBufferString(data.Encode()))
			if err != nil {
				log.Println(err.Error())
			} else {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Close = true
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					log.Println(err.Error())
				} else {
					defer resp.Body.Close()
				}
			}

		}
	} else if r.Method == "GET" {
		userId := context.Get(r, "userId")
		if !isAuthorized(userId.(uint), vars["name"], false) {
			w.WriteHeader(403)
			return
		}
		manifest := Manifest{}
		db.Where("name=? AND reference=?", vars["name"], vars["reference"]).First(&manifest)
		if manifest.Name != "" {
			w.Header().Set("Docker-Content-Digest", manifest.Digest)
			w.Header().Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")
			w.WriteHeader(200)
			w.Write(manifest.Content)
		} else {
			w.WriteHeader(400)
		}

	} else if r.Method == "HEAD" {
		userId := context.Get(r, "userId")
		if !isAuthorized(userId.(uint), vars["name"], false) {
			w.WriteHeader(403)
			return
		}
		manifest := Manifest{}
		db.Where("name=? AND reference=?", vars["name"], vars["reference"]).First(&manifest)
		if manifest.Name != "" {
			w.Header().Set("Content-Length", strconv.Itoa(len(manifest.Content)))
			w.Header().Set("Docker-Content-Digest", manifest.Digest)
			w.WriteHeader(200)
		} else {
			w.WriteHeader(404)
		}
	}
}
