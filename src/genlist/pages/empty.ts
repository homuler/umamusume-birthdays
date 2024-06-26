import { Browser, Page } from 'puppeteer';
import { CharactersPage } from './characters';
import { BasePage, umamusumeTopUrl } from './common';
import { CharacterPage } from './character';

const charactersPageUrl = `${umamusumeTopUrl}/character`;

class EmptyPage extends BasePage {
  public constructor(page: Page) {
    super(page);
  }

  public async goToCharactersPage(): Promise<CharactersPage> {
    await this.goto(charactersPageUrl);
    return new CharactersPage(this.page);
  }

  public async goToCharacterPage(url: string): Promise<CharacterPage> {
    await this.goto(url);
    return new CharacterPage(this.page);
  }

  public async reset(): Promise<void> {
    return super.reset();
  }
}

export const openEmptyPage = async (browser: Browser): Promise<EmptyPage> => {
  const page = await browser.newPage();
  return new EmptyPage(page);
};
