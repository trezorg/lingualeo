FROM golang:latest as builder
RUN mkdir /build 
ADD . /build/
WORKDIR /build 
RUN go get github.com/PuerkitoBio/goquery \
    github.com/alyu/configparser \
    github.com/wsxiaoys/terminal/color && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o lingualeo . && \
    useradd -M -N -o -u 1000 lingualeo
FROM scratch
ENV USER=lingualeo
COPY --from=builder /build/lingualeo /app/
COPY --from=builder /etc/passwd /etc/passwd
WORKDIR /app
USER lingualeo
ENTRYPOINT ["/app/lingualeo"]
