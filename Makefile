ID = timeplus-proton-datasource
Version = 2.1.4

init:
	npm install

dev:
	npm run dev

build:
	mage -v

debug_build:
	mage build:debug

package:
	mv dist/ $(ID)
	zip $(ID)-$(Version).zip $(ID) -r
	mv $(ID) dist/

validate:
	npx -y @grafana/plugin-validator@latest ./$(ID)-$(Version).zip

run:
	docker-compose up
