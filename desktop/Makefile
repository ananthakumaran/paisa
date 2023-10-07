.PHONY: build

develop:
	wails dev -tags webkit2_40

build:
	wails build -tags webkit2_40

dmg:
	rm -rf Paisa.dmg
	wails build
	create-dmg --volname "Paisa" --window-pos 200 120 --window-size 800 400 --icon-size 100  --icon "Paisa.app" 200 190 --app-drop-link 600 185 --hide-extension "Paisa.app" --volicon build/bin/Paisa.app/Contents/Resources/iconfile.icns  "Paisa.dmg" build/bin/Paisa.app
