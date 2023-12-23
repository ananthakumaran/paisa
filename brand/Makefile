images:
	steam-run svgexport logo.svg ../desktop/build/linux/usr/share/icons/hicolor/512x512/apps/paisa.png 100% 512:512
	steam-run svgexport logo.svg ../desktop/build/appicon.png 100% 256:256
	steam-run svgexport logo.svg ../docs/images/favicon.png 100% 256:256
	steam-run svgexport logo.svg ../docs/apple-touch-icon.png 100% 180:180
	steam-run svgexport logo.svg ../docs/android-chrome-512x512.png 100% 512:512
	steam-run svgexport logo.svg ../docs/android-chrome-192x192.png 100% 192:192
	steam-run svgexport logo.svg ../docs/favicon-16x16.png 100% 16:16
	steam-run svgexport logo.svg ../docs/favicon-32x32.png 100% 32:32
	steam-run svgexport logo.svg ../docs/favicon-48x48.png 100% 48:48
	steam-run svgexport logo.svg ../docs/favicon-64x64.png 100% 64:64
	steam-run svgexport logo.svg ../docs/favicon-256x256.png 100% 256:256
	nix-shell -p imagemagick --command 'convert ../docs/favicon-16x16.png ../docs/favicon-32x32.png ../docs/favicon-64x64.png ../docs/favicon.ico'
	nix-shell -p imagemagick --command 'convert ../docs/favicon-16x16.png ../docs/favicon-32x32.png ../docs/favicon-48x48.png ../docs/favicon-64x64.png ../docs/favicon-256x256.png ../desktop/build/windows/icon.ico'
	rm ../docs/favicon-48x48.png
	rm ../docs/favicon-64x64.png
	rm ../docs/favicon-256x256.png

social-images:
	steam-run svgexport social.svg social-1500x500.png 100% 1500:500
