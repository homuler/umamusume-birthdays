import { Browser, Page } from 'puppeteer';
import { CharactersPage } from './characters';
import { BasePage, umamusumeTopUrl } from './common';
import { CharacterPage } from './character';
import { logger } from '../log';

const charactersPageUrl = `${umamusumeTopUrl}/character`;

class EmptyPage extends BasePage {
  public constructor(page: Page) {
    super(page);
  }

  protected async goto(url: string): Promise<void> {
    logger.debug("going to another page", { url });
    await this.page.goto(url);
  }

  public async goToCharactersPage(): Promise<CharactersPage> {
    await this.goto(charactersPageUrl);
    return new CharactersPage(this.page);
  }

  public async goToCharacterPage(url: string): Promise<CharacterPage> {
    await this.goto(url);
    return new CharacterPage(this.page);
  }
}

export const openEmptyPage = async (browser: Browser): Promise<EmptyPage> => {
  const page = await browser.newPage();
  return new EmptyPage(page);
};
