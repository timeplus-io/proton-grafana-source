ID = timeplus-proton-datasource
Version = 1.0.2

init:
	npm install

dev:
	npm run dev

build:
	mage -v

package:
	mv dist/ $(ID)
	zip $(ID)-$(Version).zip $(ID) -r

validate:
	npx -y @grafana/plugin-validator@latest ./$(ID)-$(Version).zip

run:
	docker-compose up