package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/chromedp/chromedp"
	umamusume "github.com/homuler/umamusume-birthdays/src"
	"github.com/homuler/umamusume-birthdays/src/ocr"
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

var (
	charactersPath = flag.String("p", "", "path to characters.yml")
	outputPath     = flag.String("o", "", "output path")
	verbose        = flag.Bool("v", false, "enable verbose logging")
)

func main() {
	flag.Parse()

	if *charactersPath == "" {
		flag.Usage()
		panic("the path to the characters.yml must be specified")
	}
	if *outputPath == "" {
		*outputPath = *charactersPath
	}

	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}

	ctx := umamusume.WithLogger(context.Background(), umamusume.NewLogger(level))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	ocrClient, err := ocr.NewClient(umamusume.GetLogger(ctx))
	if err != nil {
		panic(fmt.Errorf("failed to initialize an OCR client: %v", err))
	}
	ctx = ocr.WithClient(ctx, ocrClient)

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	if err := run(ctx, *charactersPath, *outputPath); err != nil {
		panic(err)
	}
}

func run(ctx context.Context, inpath string, outpath string) error {
	logger := umamusume.GetLogger(ctx)

	logger.Debug("start running")

	f, err := os.ReadFile(inpath)
	if err != nil {
		return err
	}

	orig, err := umamusume.ReadYAML(bytes.NewReader(f))
	if err != nil {
		return err
	}

	tasks, err := genUmaTasks(ctx)
	if err != nil {
		return err
	}
	logger.Debug("generated tasks", slog.Any("data", tasks))

	new := make([]*umamusume.Uma, 0, len(orig))
	for i, task := range tasks {
		uma, err := task.do(ctx)
		if err != nil {
			return err
		}
		new = append(new, uma)
		logger.Info("task done", slog.Int("count", i+1), slog.Any("result", *uma))
	}

	new = umamusume.Update(orig, new)
	out, err := yaml.Marshal(new)

	dir := path.Dir(outpath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	if err := os.WriteFile(outpath, out, 0644); err != nil {
		return err
	}
	return nil
}
