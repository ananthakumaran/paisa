FROM docker.io/ananthakumaran/paisa:v0.7.4

RUN apk --no-cache add hledger beancount

WORKDIR /root/

CMD ["paisa", "serve"]
