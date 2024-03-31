import { Page } from "puppeteer";
import { logger } from "../log";

export const umamusumeTopUrl = "https://umamusume.jp";

export class BasePage {
  protected constructor(protected readonly page: Page) {}

  protected async goto(url: string): Promise<void> {
    logger.debug("going to another page", { url });
    await this.page.goto(url);
  }

  public url(): string {
    return this.page.url();
  }
}
