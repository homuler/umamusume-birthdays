package ocr

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/otiai10/gosseract/v2"
	"golang.org/x/exp/slog"
	"golang.org/x/text/unicode/norm"
)

type Client interface {
	Read([]byte) (*Result, error)
	Close() error
}

type birthday struct {
	Month int
	Day   int
}

type Result struct {
	Name     string
	Birthday birthday
}

type client struct {
	client *gosseract.Client
	logger *slog.Logger
}

var _ Client = (*client)(nil)

type contextKey struct{}

func WithClient(ctx context.Context, client Client) context.Context {
	return context.WithValue(ctx, contextKey{}, client)
}

func FromContext(ctx context.Context) *client {
	c, _ := ctx.Value(contextKey{}).(*client)
	return c
}

func NewClient(logger *slog.Logger) (*client, error) {
	c := gosseract.NewClient()
	if err := c.SetLanguage("eng", "jpn"); err != nil {
		defer c.Close()
		return nil, err
	}

	return &client{client: c, logger: logger}, nil
}

func (c *client) Read(img []byte) (*Result, error) {
	if err := c.client.SetImageFromBytes(img); err != nil {
		return nil, fmt.Errorf("failed to set the image for OCR: %w", err)
	}

	text, err := c.client.Text()
	if err != nil {
		return nil, fmt.Errorf("failed to read texts from the image: %w", err)
	}

	res, err := c.buildResult(text)
	if err != nil {
		return nil, fmt.Errorf("failed to read the profile: %w\n\n%v", err, text)
	}
	return res, nil
}

func (c *client) Close() error {
	return c.client.Close()
}

func (c *client) buildResult(text string) (*Result, error) {
	ls := strings.Split(text, "\n")
	var res Result

	for _, l := range ls {
		if name, ok := readAsUmaName(l); ok {
			res.Name = name
			continue
		}
		if b, ok := readAsBirthday(l); ok {
			res.Birthday = b
			break
		}
	}

	if res.Birthday.Month == 0 {
		c.logger.Debug("birthday is unknown", slog.String("data", text))
	}

	// 何も読めてない可能性があるが、今のところ許容する
	return &res, nil
}

func readAsUmaName(line string) (string, bool) {
	var sb strings.Builder

	for _, r := range line {
		if unicode.IsSpace(r) {
			continue
		}
		if !unicode.In(r, unicode.Katakana) && r != 'ー' {
			return "", false
		}
		sb.WriteRune(r)
	}
	return sb.String(), true
}

var (
	birthdayPrefixPattern = regexp.MustCompile(`^誕\s*生\s*日`)
	birthdayPattern1      = regexp.MustCompile(`(\d+)\s*月\s*(\d+)\s*日$`)
	birthdayPattern2      = regexp.MustCompile(`(\d+)8(\d+)8$`) // e.g. マルゼンスキー
	birthdayPattern3      = regexp.MustCompile(`(\d+)A(\d+)H$`) // e.g. マンハッタンカフェ
)

func readAsBirthday(line string) (birthday, bool) {
	normalized := norm.NFKC.String(line)

	if !strings.HasPrefix(normalized, "誕生 日") && !strings.HasPrefix(normalized, "HEH") &&
		!strings.HasPrefix(normalized, "HER") && !strings.HasPrefix(normalized, "BER") && !strings.HasPrefix(normalized, "#;ER") {
		return birthday{}, false
	}

	patterns := []regexp.Regexp{
		*birthdayPattern1,
		*birthdayPattern2,
		*birthdayPattern3,
	}

	for _, pattern := range patterns {
		ss := pattern.FindStringSubmatch(normalized)
		if len(ss) > 0 {
			m, _ := strconv.Atoi(ss[1])
			d, _ := strconv.Atoi(ss[2])
			return birthday{Month: m, Day: d}, true
		}
	}

	return birthday{}, false
}
