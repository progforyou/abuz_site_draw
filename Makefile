PLATFORM=macosx_x64

build:
	terser parts/abuz_site_draw/pkg/client/static/js/main_t.js --output parts/abuz_site_draw/pkg/client/static/js/main.js
	go run -mod vendor ./build-tools/bin-maker/bin-maker.go