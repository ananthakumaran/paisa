.PHONY: docs
.PHONY: fixture/main.transactions.json

develop:
	./node_modules/.bin/concurrently --names "GO,JS" -c "auto" "make serve" "npm run dev"

serve:
	./node_modules/.bin/nodemon --signal SIGTERM --delay 2000ms --watch '.' --ext go,json --exec 'go run . serve || exit 1'
watch:
	npm run "build:watch"
docs:
	cd docs && mdbook serve --open

sample:
	go build && ./paisa init && ./paisa update

publish:
	nix develop --command bash -c 'cd docs && mdbook build'

lint:
	./node_modules/.bin/prettier --check src
	npm run check
	test -z $$(gofmt -l .)

test:
	NODE_OPTIONS=--experimental-vm-modules npm run test
	npm run build
	go test ./...

windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc go build


deploy:
	docker build -t paisa .
	fly deploy -i paisa:latest --local-only

install:
	npm run build
	go build
	cp paisa ~/.local/bin

fixture/main.transactions.json:
	cd /tmp && paisa init
	cp fixture/main.ledger /tmp/personal.ledger
	cd /tmp && paisa update --journal && paisa serve -p 6500 &
	sleep 1
	curl http://localhost:6500/api/transaction | jq .transactions > fixture/main.transactions.json
	pkill -f 'paisa serve -p 6500'
