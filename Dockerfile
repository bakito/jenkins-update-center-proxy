FROM docker.io/library/golang:1.18 as builder

WORKDIR /go/src/app

RUN apt-get update && apt-get install -y upx

ARG VERSION=master
ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux

ADD . /go/src/app/

RUN go build -a -installsuffix cgo -ldflags="-w -s -X github.com/bakito/jenkins-update-center-proxy/version.Version=${VERSION}" -o jenkins-update-center-proxy . \
  && upx -q jenkins-update-center-proxy

# application image
FROM scratch
WORKDIR /opt/go

LABEL maintainer="bakito <github@bakito.ch>"
EXPOSE 8080
ENTRYPOINT ["/opt/go/jenkins-update-center-proxy"]

COPY --from=builder /go/src/app/jenkins-update-center-proxy  /opt/go/jenkins-update-center-proxy
