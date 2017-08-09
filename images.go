package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

func createDockerHash(content []byte) string {
	sha_256 := sha256.New()
	sha_256.Write(content)
	return "sha256:" + fmt.Sprintf("%x", sha_256.Sum(nil))
}
func Upload(w http.ResponseWriter, r *http.Request) {
	db, err := gorm.Open("mysql", os.Getenv("MYSQL_CONN"))
	if err != nil {
		log.Println(err.Error())
	}
	defer db.Close()
	digest := r.URL.Query().Get("digest")
	vars := mux.Vars(r)
	userId := context.Get(r, "userId")
	if !isAuthorized(userId.(uint), vars["name"], true) {
		w.WriteHeader(403)
		return
	}
	if r.Method == "POST" && vars["uuid"] == "" {
		u1 := uuid.NewV4()
		w.Header().Set("Content-Length", "0")
		w.Header().Add("Location", "/v2/"+vars["name"]+"/blobs/uploads/"+u1.String())
		w.Header().Set("Range", "0-0")
		w.Header().Add("Docker-Upload-UUID", u1.String())
		w.WriteHeader(202)
	} else if r.Method == "PATCH" {
		fi, err := os.OpenFile("/opt/registry/tmp/"+vars["uuid"], os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Println(err.Error())
			return
		}
		defer fi.Close()
		size, _ := io.Copy(fi, r.Body)
		w.Header().Set("Content-Length", "0")
		w.Header().Add("Location", "/v2/"+vars["name"]+"/blobs/uploads/"+vars["uuid"])
		w.Header().Set("Range", "0-"+strconv.Itoa(int(size)))
		w.Header().Add("Docker-Upload-UUID", vars["uuid"])
		w.WriteHeader(204)
	} else if r.Method == "PUT" {
		fi, err := os.OpenFile("/opt/registry/tmp/"+vars["uuid"], os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return
		}
		defer fi.Close()
		os.Rename("/opt/registry/tmp/"+vars["uuid"], "/opt/registry/images/"+digest)
		size, _ := io.Copy(fi, r.Body)
		w.Header().Set("Content-Length", "0")
		w.Header().Add("Location", "/v2/"+vars["name"]+"/blobs/"+digest)
		w.Header().Set("Content-Range", "0-"+strconv.Itoa(int(size)))
		w.Header().Add("Docker-Content-Digest", digest)
		w.WriteHeader(204)
	}
}
func getBlob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if r.Method == "HEAD" {
		if _, err := os.Stat("/opt/registry/images/" + vars["uuid"]); err == nil {
			w.Header().Set("Content-Length", "0")
			w.Header().Set("Docker-Content-Digest", vars["uuid"])
			w.WriteHeader(200)
		} else {
			w.WriteHeader(404)
		}
	} else if r.Method == "GET" {
		if fi, err := os.Stat("/opt/registry/images/" + vars["uuid"]); err == nil {
			w.Header().Set("Content-Length", strconv.Itoa(int(fi.Size())))
			w.Header().Set("Docker-Content-Digest", vars["uuid"])
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(200)
			f, err := os.OpenFile("/opt/registry/images/"+vars["uuid"], os.O_RDONLY, 0600)
			defer f.Close()
			if err != nil {
				w.WriteHeader(404)
				return
			}
			io.Copy(w, f)
		} else {
			w.WriteHeader(404)
		}
	}
}
func cleanImages() {
	db, err := gorm.Open("mysql", os.Getenv("MYSQL_CONN"))
	if err != nil {
		log.Println(err.Error())
	}
	defer db.Close()
	manifests := make([]Manifest, 0)
	db.Find(&manifests)
	digests := make([]string, 0)
	m := map[string]interface{}{}
	for _, v := range manifests {
		err = json.Unmarshal(v.Content, &m)
		if err != nil {
			log.Error(err)
			continue
		}
		s := m["layers"].([]interface{})
		for i := 0; i < len(s); i++ {
			mv := s[i].(map[string]interface{})
			digests = append(digests, mv["digest"].(string))
		}
	}
	files, _ := ioutil.ReadDir("/opt/registry/images")
	for _, f := range files {
		if !in_array(f.Name(), digests) {
			os.Remove("/opt/registry/images/" + f.Name())
		}
	}
}
func in_array(name string, arr []string) bool {
	for _, v := range arr {
		if name == v {
			return true
		}
	}
	return false
}
