.PHONY: gencal genlist clean

GENCAL = src/cmd/gencal/gencal
GENLIST = src/cmd/genlist/genlist

$(GENCAL):
	cd src/cmd/gencal && go build

gencal: $(GENCAL)
	./$(GENCAL) -p ./data/characters.yml -o ./resources/birthdays.ics

$(GENLIST):
	cd src/cmd/genlist && go build

genlist: $(GENLIST)
	./$(GENLIST) -p ./data/characters.yml -v

clean:
	rm -f $(GENCAL)
	rm -f $(GENLIST)