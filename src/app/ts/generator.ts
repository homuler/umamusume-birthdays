import fs from "fs";
import moment from "moment";
import { Birthday } from "./models/Birthday";
import { CalendarEvent } from "./models/CalendarEvent";

export class Generator {
  run() {
    const NEWLINE = "\r\n";

    // データファイル読み込み
    // TODO: パス指定の方法がイマイチ
    const yamlText = fs.readFileSync("../../birthdays.yaml", "utf-8");

    // YAMLを解析
    const birthdays = Birthday.parse(yamlText);
    console.log(birthdays);

    // iCalendar形式の予定定義に変換
    const timestamp = moment().format("YYYYMMDDTHHmmssZ");
    const events = birthdays.map((birthday: Birthday) => {
      return birthday.names
        .map((name: string) => new CalendarEvent(name, birthday.date))
        .map((event: CalendarEvent) => {
          const _ = [];
          _.push("BEGIN:VEVENT");

          _.push("CLASS:PUBLIC");
          _.push(`UID:${event.uniqueId}`);
          _.push(`DTSTAMP:${timestamp}`);
          _.push(`SUMMARY:${event.name}の誕生日`);
          _.push(`DESCRIPTION:${event.name}の誕生日です。`);
          _.push(`RRULE:FREQ=YEARLY`);
          _.push(
            `DTSTART;VALUE=DATE:${moment(event.datetime).format("YYYYMMDD")}`
          );
          _.push(
            `DTEND;VALUE=DATE:${moment(event.datetime).format("YYYYMMDD")}`
          );

          _.push("END:VEVENT");
          return _.join(NEWLINE);
        })
        .join(NEWLINE);
    });

    // iCalendar形式のカレンダーを生成
    const _prodId = "ushibutatory-umamusume-birthdays-calendar";
    const iCal = [];
    iCal.push("BEGIN:VCALENDAR");
    iCal.push(`PRODID:${_prodId}`);
    iCal.push("VERSION:2.0");
    iCal.push("METHOD:PUBLISH");
    {
      iCal.push("BEGIN:VTIMEZONE");
      iCal.push("TZID:Asia/Tokyo");
      {
        iCal.push("BEGIN:STANDARD");
        iCal.push("DTSTART:19390101T000000");
        iCal.push("TZOFFSETFROM:+0900");
        iCal.push("TZOFFSETTO:+0900");
        iCal.push("TZNAME:JST");
        iCal.push("END:STANDARD");
      }
      iCal.push("END:VTIMEZONE");
    }
    iCal.push(events.join(NEWLINE));
    iCal.push("END:VCALENDAR");

    // ファイル書き込み
    // TODO: パス指定がイマイチ
    if (!fs.existsSync("data")) {
      fs.mkdirSync("data");
    }
    fs.writeFileSync("data/birthdays.ics", iCal.join(NEWLINE), {
      encoding: "utf-8",
    });
  }
}

new Generator().run();
