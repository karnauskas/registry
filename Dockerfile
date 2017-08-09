FROM golang:latest
WORKDIR /tmp/
RUN go get -d -v github.com/sirupsen/logrus
RUN go get -d -v golang.org/x/crypto/bcrypt
RUN go get -d -v github.com/jinzhu/gorm
RUN go get -d -v github.com/satori/go.uuid
RUN go get -d -v github.com/gorilla/mux
RUN go get -d -v github.com/gorilla/context
RUN go get -d -v github.com/go-sql-driver/mysql
COPY *.go /tmp/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o registry .

FROM scratch
WORKDIR /
COPY --from=0 /tmp/registry .
VOLUME ["/opt/registry"]
EXPOSE 5000
CMD ["./registry"]
