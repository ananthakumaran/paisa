.PHONY: docs

develop:
	./node_modules/.bin/concurrently --names "GO,JS" -c "auto" "make serve" "npm run dev"

serve:
	./node_modules/.bin/nodemon --signal SIGTERM --delay 2000ms --watch '.' --ext go --exec 'go run . serve || exit 1'
watch:
	npm run "build:watch"
docs:
	cd docs && mdbook serve --open

sample:
	go build && ./paisa init && ./paisa update && ./paisa serve

publish:
	nix develop --command bash -c 'cd docs && mdbook build'

lint:
	./node_modules/.bin/prettier --check src
	npm run check
	test -z $$(gofmt -l .)

test:
	go test ./...
	npm run test

windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc go build


deploy:
	docker build -t paisa .
	fly deploy -i paisa:latest --local-only

install:
	npm run build
	go build
	cp paisa ~/.local/bin
