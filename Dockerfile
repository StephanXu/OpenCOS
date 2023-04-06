FROM golang:1.18.2-alpine3.16 as build

WORKDIR /usr/src

RUN go env -w GOPROXY=https://goproxy.cn,direct

COPY filesrv/go.mod filesrv/go.sum ./filesrv/
RUN pushd filesrv \
    && go mod download \
    && go mod verify \
    && popd

COPY filehasher/go.mod ./filehasher/
RUN pushdd filehasher \
    && go mod download \
    && go mod verify \
    && popd

COPY filesrv filehasher ./
RUN go build ./filesrv -v -o /usr/local/bin/filesrv .
RUN go build ./filehasher -o /usr/local/bin/filehasher


FROM golang:1.18.2-alpine3.16 as final
WORKDIR /app/context

COPY --from=build /usr/local/bin/filesrv /app/filesrv
COPY --from=build /usr/local/bin/filehasher /app/filehasher

ENV GIN_MODE=release
EXPOSE 4366

CMD ["/app/filesrv"]