import { Page } from "puppeteer";
import { logger } from "../log";

export const umamusumeTopUrl = "https://umamusume.jp";

export class BasePage {
  protected constructor(private _page: Page) {}

  protected get page(): Page {
    return this._page;
  }

  protected async goto(url: string): Promise<void> {
    logger.debug("going to another page", { url });
    await this.page.goto(url);
  }

  public url(): string {
    return this.page.url();
  }

  public close(): Promise<void> {
    return this.page.close();
  }

  protected async reset(): Promise<void> {
    const browser = this.page.browser();

    logger.debug("close and reopening an empty page");
    await this.close();

    this._page = await browser.newPage();
  }
}
