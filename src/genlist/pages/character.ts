import { Page } from 'puppeteer';
import { BasePage } from './common';
import { Costume, Profile } from '../data';

const detailTextSelector = 'div.character-detail__text';
const profileSelector = `${detailTextSelector} > dl.profile`;
const nameSelector = `${detailTextSelector} > div > div.name`;

const detailVisualSelector = 'div.character-detail__visual';
const costumeImageSelector = `${detailVisualSelector} > div.character-detail__visual-main ul.splide__list > li > img`;

export class CharacterPage extends BasePage {
  private async waitForDetail() {
    await this.page.waitForSelector(detailVisualSelector, { timeout: 10000 });
    await this.page.waitForSelector(detailTextSelector, { timeout: 10000 });
    await this.page.waitForNetworkIdle();
  }

  private async getName(): Promise<string | null> {
    const nameNode = await this.page.$(nameSelector);
    if (nameNode == null) {
      throw new Error('Name not found');
    }
    return nameNode.evaluate((node) => node.textContent);
  }

  private async getBirthday(): Promise<string | null> {
    const profileNode = await this.page.$(profileSelector);
    if (profileNode == null) {
      throw new Error('Profile not found');
    }
    const dts = await profileNode.$$('dt');
    const dds = await profileNode.$$('dd');

    if (dts.length == 0 || dds.length == 0) {
      throw new Error('Unexpected profile format');
    }
    const birthdayLabel = await dts[0].evaluate((node) => node.textContent);
    if (birthdayLabel != '誕生日') {
      throw new Error('Birthday not found');
    }
    const birthdayText = await dds[0].evaluate((node) => node.textContent);
    const normalizedBirthday = birthdayText && normalizeBirthday(birthdayText);

    return normalizedBirthday;
  }

  private async getCostume(): Promise<Costume> {
    const costumeImages = await this.page.$$(costumeImageSelector);
    const costume: Costume = { school: null, racing: null, original: null, staringFuture: null };

    for (const costumeImage of costumeImages) {
      const src = await costumeImage.evaluate((node) => node.getAttribute('src'));
      const alt = await costumeImage.evaluate((node) => node.getAttribute('alt'));
      const altTexts = (alt || "").split(" ");
      const type = altTexts[altTexts.length - 1];

      switch (type) {
        case '制服': costume.school = src; break;
        case '勝負服': costume.racing = src; break;
        case '原案': costume.original = src; break;
        case '<small>STARTING<br>FUTURE</small>': costume.staringFuture = src; break;
        default: {
          console.warn(`Unknown costume type: ${type}`);
          break;
        }
      }
    }

    return costume;
  }

  public async getProfile(): Promise<Profile> {
    await this.waitForDetail();

    const name = await this.getName();
    const birthday = await this.getBirthday();
    const costume = await this.getCostume();

    return {
      name,
      birthday,
      url: this.page.url(),
      costume,
    };
  }
}

const birthdayRegex = /(\d+)月(\d+)日/;

function normalizeBirthday(birthday: string): string | null {
  const matched = birthday.match(birthdayRegex);
  if (matched == null) {
    // ？？？のことがある
    return null;
  }
  return `${matched[1].padStart(2, '0')}/${matched[2].padStart(2, '0')}`;
}
