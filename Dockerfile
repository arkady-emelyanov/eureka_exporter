FROM golang:alpine AS build
RUN apk --update add git

WORKDIR /src
ADD go.mod .
ADD go.sum .
RUN go mod download

ENV GO111MODULE=on \
    CGO_ENABLED=0

COPY . .
RUN go build -o app ./main.go

FROM alpine
COPY --from=build /src/app /
ENTRYPOINT /app
