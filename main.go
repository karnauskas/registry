package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

func createTables() {
	db, err := gorm.Open("mysql", os.Getenv("MYSQL_CONN"))
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer db.Close()
	db.AutoMigrate(&Repo{})
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Manifest{})
	db.AutoMigrate(&Permission{})
}
func notFound(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v %v %v\n", r.Method, r.URL, r.Proto)
	w.WriteHeader(404)
}
func Version(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(204)
}
func main() {
	os.MkdirAll("/opt/registry/images", os.ModePerm)
	os.MkdirAll("/opt/registry/tmp", os.ModePerm)
	createTables()
	go func() {
		t := time.NewTimer(time.Hour)
		for {
			cleanImages()
			<-t.C
		}
	}()
	r := mux.NewRouter()
	r.HandleFunc("/", basicAuth(Version))
	r.HandleFunc("/v2/", basicAuth(Version))
	r.HandleFunc("/v2/{name:(?:.*)}/blobs/uploads/", basicAuth(Upload)).Methods("POST")
	r.HandleFunc("/v2/{name:(?:.*)}/blobs/uploads/{uuid}", basicAuth(Upload)).Methods("PATCH", "PUT")
	r.HandleFunc("/v2/{name:(?:.*)}/blobs/{uuid}", basicAuth(getBlob)).Methods("HEAD", "GET")
	r.HandleFunc("/v2/{name:(?:.*)}/manifests/{reference}", basicAuth(Manifests)).Methods("PUT", "GET", "HEAD")
	r.NotFoundHandler = http.HandlerFunc(notFound)
	log.Fatal(http.ListenAndServe(":5000", r))
}
