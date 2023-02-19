package main

import (
	"bytes"
	"flag"
	"fmt"
	_ "image/png"
	"os"
	"path"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	umamusume "github.com/homuler/umamusume-birthdays/src"
	"golang.org/x/exp/slog"
)

const prodID = "homuler-umamusume-birthdays-calendar"
const propertyAltDesc ics.Property = "X-ALT-DESC"

var (
	charactersPath = flag.String("p", "", "path to characters.yml")
	outputPath     = flag.String("o", "birthdays.ics", "output path")
	verbose        = flag.Bool("v", false, "enable verbose logging")
)

func main() {
	flag.Parse()

	logLevel := slog.LevelInfo
	if *verbose {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(
		slog.HandlerOptions{
			AddSource: true,
			Level:     logLevel,
		}.NewJSONHandler(os.Stdout))

	if *charactersPath == "" {
		panic("path must be specified")
	}

	contents, err := os.ReadFile(*charactersPath)
	if err != nil {
		panic(fmt.Sprintf("failed to read %v: %v", *charactersPath, err))
	}
	characters, err := umamusume.ReadYAML(bytes.NewReader(contents))
	if err != nil {
		panic(err)
	}

	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetProductId(prodID)
	cal.SetTzid("Asia/Tokyo")

	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	y := time.Now().In(jst).Year()

	for _, uma := range characters {
		logger.Debug("Processing new uma", slog.Any("uma", uma))

		if len(uma.Birthday) == 0 {
			logger.Info(fmt.Sprintf("The birthday of %s is unknown for now", uma.Name))
			continue
		}

		t, err := time.ParseInLocation("2006/01/02", fmt.Sprintf("%v/%s", y, uma.Birthday), jst)
		if err != nil {
			// ログ出力だけして続行
			logger.Error(fmt.Sprintf("failed to parse birthday: %s", uma.Birthday), err)
			continue
		}

		evt := cal.AddEvent(uma.Name)
		evt.SetClass(ics.ClassificationPublic)
		evt.SetDtStampTime(time.Now())
		evt.SetSummary(fmt.Sprintf("%sの誕生日", uma.Name))
		evt.SetDescription(renderString(uma))
		evt.SetProperty(ics.ComponentProperty(propertyAltDesc), renderHTML(uma), ics.WithFmtType("text/html"))
		evt.SetURL(uma.Url)
		evt.AddRrule("FREQ=YEARLY")
		evt.SetProperty(ics.ComponentPropertyDtStart, t.In(jst).Format("20060102"), ics.WithValue(string(ics.ValueDataTypeDate)))
	}

	dir := path.Dir(*outputPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.WriteFile(*outputPath, []byte(cal.Serialize()), 0644); err != nil {
		panic(err)
	}
}

func renderString(uma *umamusume.Uma) string {
	return fmt.Sprintf("<a href='%s'>%s</a>の誕生日です", uma.Url, uma.Name)
}

func renderHTML(uma *umamusume.Uma) string {
	var sb strings.Builder

	sb.WriteString("<!doctype html><html><body>")
	sb.WriteString(fmt.Sprintf("<p><a href='%s'>%s</a>の誕生日です.</p>", uma.Url, uma.Name))
	sb.WriteString("<div style='display: flex; justify-content: flex-start; height: 200px'>")
	if len(uma.Costumes.Racing) > 0 {
		sb.WriteString(fmt.Sprintf("<img src='%s' alt='勝負服' />", uma.Costumes.Racing))
	}
	if len(uma.Costumes.School) > 0 {
		sb.WriteString(fmt.Sprintf("<img src='%s' alt='制服' />", uma.Costumes.School))
	}
	if len(uma.Costumes.SF) > 0 {
		sb.WriteString(fmt.Sprintf("<img src='%s' alt='STARTING FUTURE' />", uma.Costumes.SF))
	}
	if len(uma.Costumes.Original) > 0 {
		sb.WriteString(fmt.Sprintf("<img src='%s' alt='原案' />", uma.Costumes.Original))
	}
	sb.WriteString("</div></body></html>")

	return sb.String()
}
