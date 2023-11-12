.PHONY: docs
.PHONY: fixture/main.transactions.json

develop:
	./node_modules/.bin/concurrently --names "GO,JS" -c "auto" "make serve" "npm run dev"

serve:
	./node_modules/.bin/nodemon --signal SIGTERM --delay 2000ms --watch '.' --ext go,json --exec 'go run . serve || exit 1'

debug:
	./node_modules/.bin/concurrently --names "GO,JS" -c "auto" "make serve-now" "npm run dev"

serve-now:
	./node_modules/.bin/nodemon --signal SIGTERM --delay 2000ms --watch '.' --ext go,json --exec 'TZ=UTC go run . serve --now 2022-02-07 || exit 1'


watch:
	npm run "build:watch"
docs:
	mkdocs serve -a 0.0.0.0:8000

sample:
	go build && ./paisa init && ./paisa update

publish:
	nix develop --command bash -c 'mkdocs build'

lint:
	./node_modules/.bin/prettier --check src
	npm run check
	test -z $$(gofmt -l .)

regen:
	go build
	unset PAISA_CONFIG && REGENERATE=true TZ=UTC bun test tests

jstest:
	bun test --preload ./src/happydom.ts src
	go build
	unset PAISA_CONFIG && TZ=UTC bun test tests

jsbuild:
	npm run build

test: jsbuild jstest
	go test ./...

windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc go build


deploy:
	fly scale count 2 --region lax --yes
	docker build -t paisa . --file Dockerfile.demo
	fly deploy -i paisa:latest --local-only
	fly scale count 1 --region lax --yes

install:
	npm run build
	go build
	go install

fixture/main.transactions.json:
	cd /tmp && paisa init
	cp fixture/main.ledger /tmp/main.ledger
	cd /tmp && paisa update --journal && paisa serve -p 6500 &
	sleep 1
	curl http://localhost:6500/api/transaction | jq .transactions > fixture/main.transactions.json
	pkill -f 'paisa serve -p 6500'

generate-fonts:
	bun download-svgs.js
	node generate-font.js

node2nix:
	node2nix --development -18 --input package.json \
	--lock package-lock.json \
	--node-env ./flake/node-env.nix \
	--composition ./flake/default.nix \
	--output ./flake/node-package.nix
