FROM golang:latest as builder
ENV USER=lingualeo
RUN mkdir /build 
ADD . /build/
WORKDIR /build 
RUN go get github.com/PuerkitoBio/goquery \
    github.com/alyu/configparser \
    github.com/sirupsen/logrus \
    gopkg.in/yaml.v2 \
    github.com/wsxiaoys/terminal/color && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o ${USER} . && \
    useradd -M -N -o -u 1000 ${USER}
FROM scratch
ENV USER=lingualeo APP=/app
COPY --from=builder /build/${USER} ${APP}/
COPY --from=builder /etc/passwd /etc/passwd
WORKDIR ${APP}
USER ${USER}
ENTRYPOINT ["/app/lingualeo"]
