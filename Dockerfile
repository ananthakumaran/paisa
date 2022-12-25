FROM node:18-alpine3.16 as web
WORKDIR /usr/src/paisa
COPY . .
RUN npm install && \
    PAISA_HOST=https://paisa-demo.ananthakumaran.in npm run build

FROM golang:1.18-alpine3.16 as go
WORKDIR /usr/src/paisa
RUN apk --no-cache add sqlite gcc g++
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
COPY --from=web /usr/src/paisa/web/static ./web/static
RUN CGO_ENABLED=1 go build

FROM alpine:3.16
RUN apk --no-cache add ca-certificates ledger
WORKDIR /root/
COPY --from=go /usr/src/paisa ./
RUN ./paisa init && ./paisa update
EXPOSE 7500
CMD ["./paisa", "serve"]
