.PHONY: gencal synclist clean

GENCAL = src/gencal/gencal
GENLIST = src/genlist/dist/index.js

$(GENCAL):
	cd src/gencal && go build

gencal: $(GENCAL)
	./$(GENCAL) -p ./data/characters.yml -o ./resources/birthdays.ics

$(GENLIST):
	cd src/genlist && npm run build

synclist: $(GENLIST)
	node ./$(GENLIST) -p ./data/characters.yml -vv

clean:
	rm -f $(GENCAL)
	rm -f $(GENLIST)