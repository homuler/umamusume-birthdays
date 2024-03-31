import fs from "fs";
import puppeteer from "puppeteer";
import YAML from "yaml";
import yargs from "yargs";

import { Profile } from "./data";
import { openEmptyPage } from "./pages/empty";

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

  const yamlPath = argv.p;
  const data = readUmaYAML(yamlPath);
  console.log(data);

  const browser = await puppeteer.launch({
    headless: false,
    args: [`--window-size=1920,1080`],
    defaultViewport: {
      width:1920,
      height:1080
    },
  });

  const emptyPage = await openEmptyPage(browser);
  const charactersPage = await emptyPage.goToCharactersPage();
  console.log("opened charactersPage");
  const characters = await charactersPage.getCharacterCards();
  console.log(characters);

  const profiles: Profile[] = [];
  for (const character of characters) {
    const characterPage = await emptyPage.goToCharacterPage(character.url);
    console.log("opened characterPage");
    const profile = await characterPage.getProfile();
    console.log(profile);
    if (profile.name === null) {
      console.log("null name");
      continue;
    }
    profiles.push(profile);
  }
  profiles.sort((a, b) => (a.name as string).localeCompare(b.name as string));
  console.log(profiles);

  const newUmaList = updateUmaList(data, profiles);
  console.log(newUmaList);

  const newYAML = YAML.stringify(newUmaList);
  fs.writeFileSync(yamlPath, newYAML);
};

main().then(() => {
  process.exit(0);
});
