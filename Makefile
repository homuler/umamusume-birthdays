.PHONY: build

build:
	go run src/cmd/gen.go -p ./data/characters.yml -o ./resources/birthdays.ics
