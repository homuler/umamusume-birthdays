import fs from "fs";
import puppeteer from "puppeteer";
import YAML from "yaml";
import yargs from "yargs";

import { Profile } from "./data";
import { openEmptyPage } from "./pages/empty";
import { retriable } from "./promise";
import { logger } from "./log";

interface Uma {
  name: string;
  birthday: string;
  url: string;
  playable: boolean;
  costumes: {
    school: string;
    racing: string;
    original: string;
    sf: string;
  };
  variations: {
    url: string;
  }[];
}

const readUmaYAML = (path: string): Uma[] => {
  const file = fs.readFileSync(path, "utf8");
  return YAML.parse(file);
}

const updateUmaList = (umas: Uma[], profiles: Profile[]): Uma[] => {
  const result: Uma[] = [];

  for (const profile of profiles) {
    const newUma = {
      name: profile.name as string,
      birthday: profile.birthday || "",
      url: profile.url,
      playable: profile.costume.racing !== null,
      costumes: {
        school: profile.costume.school || "",
        racing: profile.costume.racing || "",
        original: profile.costume.original || "",
        sf: profile.costume.staringFuture || "",
      },
      variations: [],
    };
    const uma = umas.find((uma) => uma.name === newUma.name);

    if (uma === void 0) {
      // 新規ウマ娘
      result.push(newUma);
      continue;
    }

    result.push({
      ...newUma,
      variations: uma.variations,
    });
  }

  return result;
};

const parser = yargs(process.argv.slice(2)).options({
  p: { type: 'string', demandOption: true },
  v: { type: 'count' },
});

const main = async () => {
  const argv = await parser.parse();

  logger.setLevel(argv.v);

  const yamlPath = argv.p;
  logger.info("reading characters.yml", { path: yamlPath });
  const data = readUmaYAML(yamlPath);

  logger.info("launching puppeteer");
  const browser = await puppeteer.launch({
    headless: true,
    args: [`--window-size=1920,1080`],
    defaultViewport: {
      width:1920,
      height:1080
    },
  });

  logger.info("getting the character list");
  const emptyPage = await openEmptyPage(browser);
  const characters = await retriable(async () => {
    const charactersPage = await emptyPage.goToCharactersPage();
    return charactersPage.getCharacterCards();
  }, 3);

  logger.debug("got the character list", { count: characters.length, data: characters });

  logger.info("getting the profile of each character");

  const profiles: Profile[] = [];
  for (const character of characters) {
    logger.info("getting the profile", { target: character });

    const profile = await retriable(async () => {
      const characterPage = await emptyPage.goToCharacterPage(character.url);
      return characterPage.getProfile();
    }, 3);

    logger.debug("got the profile", { data: profile });
    if (profile.name === null) {
      logger.warn("got the profile with null name", { target: character, data: profile });
      continue;
    }
    profiles.push(profile);
  }

  logger.debug("closing the browser");
  await browser.close();

  profiles.sort((a, b) => (a.name as string).localeCompare(b.name as string));
  logger.debug("got all profiles", { count: profiles.length, data: profiles });

  logger.info("saving the updated character list", { path: yamlPath });
  const newUmaList = updateUmaList(data, profiles);
  const newYAML = YAML.stringify(newUmaList);
  fs.writeFileSync(yamlPath, newYAML);
};

main().then(() => {
  process.exit(0);
});
