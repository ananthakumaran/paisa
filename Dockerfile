FROM node:18-alpine3.18 AS web
WORKDIR /usr/src/paisa
COPY package.json package-lock.json* ./
RUN npm install
COPY . .
RUN npm run build

FROM golang:1.21-alpine3.18 AS go
WORKDIR /usr/src/paisa
RUN apk --no-cache add sqlite gcc g++
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
COPY --from=web /usr/src/paisa/web/static ./web/static
RUN CGO_ENABLED=1 go build

FROM alpine:3.18
RUN apk --no-cache add ca-certificates ledger tzdata
WORKDIR /root/
COPY --from=go /usr/src/paisa/paisa /usr/bin
EXPOSE 7500
CMD ["paisa", "serve"]
