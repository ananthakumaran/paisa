.PHONY: docs clean
.PHONY: fixture/main.transactions.json

clean:
	deno task clean

develop:
	@if [ ! -f web/static/index.html ]; then deno task build; fi
	deno task develop

serve:
	deno task serve

debug:
	@if [ ! -f web/static/index.html ]; then deno task build; fi
	deno task debug

serve-now:
	deno task serve:now


watch:
	deno task build:watch
docs:
	mkdocs serve -a 0.0.0.0:8000

sample:
	go build && ./paisa init && ./paisa update

publish:
	nix develop --command bash -c 'mkdocs build'

parser:
	deno task parser-build-debug

lint:
	deno task lint
	deno task check
	test -z $$(gofmt -l .)

regen:
	go build
	unset PAISA_CONFIG && REGENERATE=true TZ=UTC deno task test:integration

jstest:
	deno task test:unit
	go build
	unset PAISA_CONFIG && TZ=UTC deno task test:integration

jsbuild:
	deno task build

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
	deno task build
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
	deno run -A download-svgs.js
	node generate-font.js

node2nix:
	npm install --lockfile-version 2
	node2nix --development -18 --input package.json \
	--lock package-lock.json \
	--node-env ./flake/node-env.nix \
	--composition ./flake/default.nix \
	--output ./flake/node-package.nix
