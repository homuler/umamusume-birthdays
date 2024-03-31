import { logger } from "./log";

export async function retriable<T>(builder: () => Promise<T>, count: number, backoff = 0): Promise<T> {
  try {
    return await builder();
  } catch (e) {
    if (count <= 0) {
      throw e;
    }
    logger.warn("retrying...", { count, backoff });
    await sleep(50 * Math.pow(2, (backoff)));
    return retriable(builder, count - 1, backoff + 1);
  }
}

export const sleep = async (ms: number): Promise<void> => {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
};
