FROM golang:1.22.3-alpine3.19 AS dev

RUN apk add --no-cache git

RUN go install github.com/githubnemo/CompileDaemon@v1.4.0

COPY ./ /app
WORKDIR /app

RUN go mod tidy && go build main.go
ENTRYPOINT CompileDaemon --build="go build main.go" --command=./main

FROM golang:1.22.3-alpine3.19 AS prod
COPY --from=dev /app/main .
EXPOSE 5000
CMD ["./main" ]