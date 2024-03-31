import winston from 'winston';

class UmaLogger {
  private readonly consoleTransport: winston.transports.ConsoleTransportInstance;
  private readonly logger: winston.Logger;

  public constructor() {
    const { combine, timestamp, json } = winston.format;
    this.consoleTransport = new winston.transports.Console({
      format: combine(
        timestamp(),
        winston.format((info) => {
          const { timestamp, level, message, ...rest } = info;
          return { timestamp, level, message, ...rest };
        })(),
        json({ deterministic: false }),
      ),
    });

    this.logger = winston.createLogger({
      transports: [this.consoleTransport],
    });
  }

  public setLevel(verbosity: number): void {
    if (verbosity === 0) {
      this.consoleTransport.level = 'warn';
    } else if (verbosity === 1) {
      this.consoleTransport.level = 'info';
    } else if (verbosity >= 2) {
      this.consoleTransport.level = 'debug';
    }
  }

  public error(message: string, meta: { [key: string]: any } = {}): void {
    this.log('error', message, meta);
  }

  public warn(message: string, meta: { [key: string]: any } = {}): void {
    this.log('warn', message, meta);
  }

  public info(message: string, meta: { [key: string]: any } = {}): void {
    this.log('info', message, meta);
  }

  public debug(message: string, meta: { [key: string]: any } = {}): void {
    this.log('debug', message, meta);
  }

  private log(level: string, message: string, meta: { [key: string]: any }): void {
    this.logger.log({ level, message, ...meta });
  }
}

export const logger = new UmaLogger();
