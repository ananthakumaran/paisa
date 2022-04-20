.PHONY: docs

serve:
	./node_modules/.bin/nodemon --signal SIGTERM --watch '.' --ext go --exec 'go run . serve || exit 1'
watch:
	./node_modules/.bin/esbuild web/src/index.ts --bundle --watch --sourcemap --outfile=web/static/dist.js
docs:
	cd docs && mdbook serve --open

publish:
	nix-shell --run 'cd docs && mdbook build'
