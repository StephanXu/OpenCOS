FROM golang:1.18.2-alpine3.16 as build

WORKDIR /usr/src

RUN go env -w GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/filesrv .


FROM golang:1.18.2-alpine3.16 as final
WORKDIR /app/context

COPY --from=build /usr/local/bin/filesrv /app/filesrv

ENV GIN_MODE=release
EXPOSE 4366

CMD ["/app/filesrv"]