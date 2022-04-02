serve:
	./node_modules/.bin/nodemon --signal SIGTERM --watch '.' --ext go --exec 'go run . serve'
watch:
	./node_modules/.bin/esbuild web/src/index.ts --bundle --watch --sourcemap --outfile=web/static/dist.js
