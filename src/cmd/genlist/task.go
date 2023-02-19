package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	umamusume "github.com/homuler/umamusume-birthdays/src"
	"github.com/homuler/umamusume-birthdays/src/ocr"
	"golang.org/x/exp/slog"
)

const (
	umamusumeTopUrl   = "https://umamusume.jp"
	charactersPageUrl = umamusumeTopUrl + "/character"
)

var (
	ErrUnexpectedDom        = errors.New("unexpected DOM")
	ErrProfileImageNotFound = errors.New("profile image not found")
)

type UmaTask struct {
	name string
	url  string
}

func (task *UmaTask) do(ctx context.Context) (*umamusume.Uma, error) {
	logger := umamusume.GetLogger(ctx)

	if err := chromedp.Run(ctx, chromedp.Navigate(task.url)); err != nil {
		return nil, fmt.Errorf("failed to load the profile page of '%v'(%v): %v", task.name, task.url, err)
	}

	var nodes []*cdp.Node
	if err := chromedp.Run(ctx, chromedp.Nodes("//div[@class='character-detail__frame']/img", &nodes)); err != nil {
		return nil, fmt.Errorf("failed to load the profile image for '%v'(%v): %v", task.name, task.url, err)
	}

	src, found := getAttribute(nodes[0], "src")
	if !found {
		return nil, fmt.Errorf("failed to find the profile image for '%v'(%v)", task.name, task.url)
	}

	resp, err := http.Get(src)
	if err != nil {
		return nil, fmt.Errorf("failed to get the profile image for '%v'(%v): %v", task.name, task.url, err)
	}
	defer resp.Body.Close()

	img, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the profile image for '%v'(%v): %v", task.name, task.url, err)
	}

	ocrClient := ocr.FromContext(ctx)
	res, err := ocrClient.Read(img)
	if err != nil {
		return nil, err
	}

	uma := umamusume.Uma{
		Name: task.name,
		Url:  task.url,
	}
	if res.Birthday.Month > 0 {
		uma.Birthday = fmt.Sprintf("%02d/%02d", res.Birthday.Month, res.Birthday.Day)
	}

	// costumes
	if err := chromedp.Run(ctx, chromedp.Nodes("//div[@class='character-detail__image']/img", &nodes)); err != nil {
		return nil, fmt.Errorf("failed to load the costume images for '%v'(%v): %v", task.name, task.url, err)
	}

	for _, costume := range nodes {
		alt, found := getAttribute(costume, "alt")
		if !found {
			logger.Warn("alt text for a costume image is not found")
			continue
		}
		src, found := getAttribute(costume, "src")
		if !found {
			logger.Warn("src for a costume image is not found")
			continue
		}

		if strings.HasSuffix(alt, "制服") {
			uma.Costumes.School = src
		} else if strings.HasSuffix(alt, "勝負服") {
			uma.Costumes.Racing = src
			uma.Playable = true
		} else if strings.HasSuffix(alt, "原案") {
			uma.Costumes.Original = src
		} else if strings.HasSuffix(alt, "<small>STARTING<br>FUTURE</small>") {
			uma.Costumes.SF = src
		} else {
			logger.Warn("unknown alt text for a costume image", slog.String("alt", alt))
		}
	}
	return &uma, nil
}

func genUmaTasks(ctx context.Context) ([]UmaTask, error) {
	logger := umamusume.GetLogger(ctx)

	if err := chromedp.Run(ctx, chromedp.Navigate(charactersPageUrl)); err != nil {
		return nil, fmt.Errorf("failed to access the character page: %v", err)
	}

	logger.Info("waiting for the character list to show...")
	var nodes []*cdp.Node
	ts := chromedp.Tasks{
		chromedp.Nodes("//section[@class='character-umamusume']/ul[@class='character__list']", &nodes),
		waitForTimeout(ctx, 10*time.Second), // liが動的に生成されるので、表示の完了を適当に待つ
	}
	if err := chromedp.Run(ctx, ts); err != nil {
		return nil, fmt.Errorf("the character list is not found: %v", err)
	}
	if err := loadChildren(ctx, nodes[0]); err != nil {
		return nil, fmt.Errorf("failed to load the character list: %v", err)
	}

	root := nodes[0]
	tasks := make([]UmaTask, 0, root.ChildNodeCount)

	for _, li := range root.Children {
		if li.NodeName != "LI" {
			return nil, fmt.Errorf("%w: the character list has elements other than LI", ErrUnexpectedDom)
		}
		if li.ChildNodeCount == 0 {
			// プレースホルダー
			continue
		}
		anchor := li.Children[0]
		if anchor.NodeName != "A" {
			return nil, fmt.Errorf("%w: a child element of the character list has elements other than A", ErrUnexpectedDom)
		}
		url, found := getAttribute(anchor, "href")
		if !found {
			return nil, fmt.Errorf("%w: an anchor tag does not have href", ErrUnexpectedDom)
		}

		if anchor.ChildNodeCount == 0 {
			return nil, fmt.Errorf("%w: an anchor tag does not have children", ErrUnexpectedDom)
		}

		img := anchor.Children[0]
		alt, found := getAttribute(img, "alt")
		if !found {
			// 名前が不明のキャラクター
			// スキップしたほうが良いかも？
			logger.Warn("unknown character found")
		}
		tasks = append(tasks, UmaTask{name: alt, url: fmt.Sprintf("%v%v", umamusumeTopUrl, url)})
	}

	return tasks, nil
}

// chromedpのhelper

func waitForLoadEvent(ctx context.Context) chromedp.Action {
	ch := make(chan struct{})
	lctx, cancel := context.WithCancel(ctx)
	go chromedp.ListenTarget(lctx, func(ev interface{}) {
		if _, ok := ev.(*page.EventLoadEventFired); ok {
			cancel()
			close(ch)
		}
	})

	return chromedp.ActionFunc(func(ctx context.Context) error {
		select {
		case <-ch:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
}

func waitForTimeout(ctx context.Context, timeout time.Duration) chromedp.Action {
	ch := make(chan struct{})
	lctx, cancel := context.WithTimeout(ctx, timeout)
	go func() {
		select {
		case <-lctx.Done():
			cancel()
			close(ch)
		}
	}()

	return chromedp.ActionFunc(func(ctx context.Context) error {
		select {
		case <-ch:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
}

func readChildren(ctx context.Context, node *cdp.Node, sel interface{}) ([]*cdp.Node, error) {
	var children []*cdp.Node
	if err := chromedp.Run(ctx, chromedp.Nodes(sel, &children, chromedp.ByQueryAll, chromedp.FromNode(node))); err != nil {
		return nil, err
	}
	return children, nil
}

func loadChildren(ctx context.Context, node *cdp.Node) error {
	_, err := readChildren(ctx, node, "*")
	return err
}

func getAttribute(node *cdp.Node, key string) (string, bool) {
	found := false
	for _, v := range node.Attributes {
		if found {
			return v, true
		}
		if v == key {
			found = true
		}
	}
	return "", false
}
