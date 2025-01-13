ID = timeplus-proton-datasource
Version = 2.1.0

init:
	npm install

dev:
	npm run dev

build:
	mage -v

package:
	mv dist/ $(ID)
	zip $(ID)-$(Version).zip $(ID) -r
	mv $(ID) dist/

validate:
	npx -y @grafana/plugin-validator@latest ./$(ID)-$(Version).zip

run:
	docker-compose up
