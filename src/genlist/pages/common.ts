import { Page } from "puppeteer";

export const umamusumeTopUrl = "https://umamusume.jp";

export class BasePage {
  protected constructor(protected readonly page: Page) {}

  public url(): string {
    return this.page.url();
  }
}
