
init:
	npm install

dev:
	npm run dev

build:
	mage -v build:linux

run:
	docker-compose up