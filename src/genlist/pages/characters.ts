import { BasePage, umamusumeTopUrl } from './common';

const charactersSelector = 'div.chara-umamusume > div.character-index__list > ul > li';

interface CharacterCard {
  url: string;
  name: string | null;
}

export class CharactersPage extends BasePage {
  private async waitForCharacters() {
    // liが少なくとも1つ、動的に生成されるのを待つ.
    await this.page.waitForSelector(charactersSelector, { timeout: 10000 });
    await this.page.waitForNetworkIdle();
  }

  public async getCharacterCards(): Promise<CharacterCard[]> {
    await this.waitForCharacters();
    const list = await this.page.$$(charactersSelector);

    if (list == null) {
      throw new Error('Unexpected DOM: Character list not found');
    }

    const cards: CharacterCard[] = [];

    for (const li of list) {
      const anchor = await li.$('a');
      if (anchor == null) {
        throw new Error('Unexpected DOM: Anchor not found');
      }

      const href = await anchor.evaluate((node) => node.getAttribute('href'));
      const nameNode = await anchor.$('div.inner p.name');
      if (nameNode == null) {
        throw new Error('Unexpected DOM: p.name not found');
      }
      const name = await nameNode.evaluate((node) => node.textContent);

      cards.push({ url: `${umamusumeTopUrl}${href}`, name })
    }

    return cards;
  }
}
