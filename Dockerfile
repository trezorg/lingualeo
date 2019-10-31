FROM golang:alpine as builder
ENV USER=lingualeo USER_ID=1000

RUN adduser -D -H -u ${USER_ID} ${USER}

ADD go.mod /build/
RUN cd /build && go mod download

ADD . /build/
RUN cd /build && GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -a -tags static_all -o ${APP_NAME} .

FROM scratch
ENV USER=lingualeo APP=/app
COPY --from=builder /build/${USER} ${APP}/
COPY --from=builder /etc/passwd /etc/passwd
WORKDIR ${APP}
USER ${USER}
ENTRYPOINT ["/app/lingualeo"]
