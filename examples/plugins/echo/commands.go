package main

import (
	"math"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/bus"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"github.com/AlonaBot/AlonaBot/sdk/service"
)

const (
	askTimeout    = time.Minute
	keywordPhrase = "echo example keyword"
	rawPhrase     = "echo example raw"
)

// Calculation 以结构体标签声明必需参数和默认参数。
type Calculation struct {
	A int64 `arg:"a"`
	B int64 `arg:"b" default:"1"`
}

// NotesQuery 请求当前插件对象的笔记快照。
type NotesQuery struct{}

// NotesSnapshot 携带笔记列表，不包含调用者身份。
type NotesSnapshot struct {
	Notes []string `json:"notes"`
}

var (
	notesService   = service.JSON[NotesQuery, NotesSnapshot]("echo/notes")
	notesPublished = bus.Define[NotesSnapshot]("echo/notes_published")
)

type notebook struct {
	mu        sync.Mutex // 命令、Service 和 Bus 回调可并发访问同一个对象。
	notes     []string
	published []string
}

func registerCommands(p *sdk.Plugin) {
	book := &notebook{}
	p.Command("echo").Menu().Args("text").Optional("text", "echo").RawTail().Handle(echoText)
	notes := p.Sub("notebook").Permission(sdk.User).Group("note")
	notes.Command("add").Alias("n", "add").Args("text").RawTail().Handle(book.add)
	notes.Command("list").Alias("n", "list").Handle(book.list)
	notes.Command("publish").Handle(func(c *sdk.Context) error {
		_, err := notesPublished.Publish(c, book.snapshot())
		if err != nil {
			return err
		}
		return c.Reply(c.T("echo.published"))
	})
	notes.Command("published").Handle(book.listPublished)
	p.Command("calc").Menu().Handle(calculate)
	p.Regex(`^roll (\d+)d(\d+)$`).Handle(func(c *sdk.Context) error {
		match := c.Match()
		return c.Reply(c.T("echo.roll", "count", match[1], "sides", match[2]))
	})
	p.Keyword(keywordPhrase).IgnoreCase().Handle(func(c *sdk.Context) (message.Content, error) {
		return message.Text(c.T("echo.keyword")), nil
	})
	p.Raw().Rule(func(c *sdk.Context) (bool, error) {
		return c.Text() == rawPhrase, nil
	}).Handle(func(c *sdk.Context) error { return c.Reply(c.T("echo.raw")) })
	p.Command("ask").Menu().Permission(sdk.User).Continue().Handle(ask)
	service.Serve(p, notesService).Access(service.AccessPublic).Handle(func(_ *sdk.Context, _ NotesQuery) (NotesSnapshot, error) {
		return book.snapshot(), nil
	})
	bus.Subscribe(p, notesPublished).Handle(func(_ *sdk.Context, snapshot NotesSnapshot) error {
		book.mu.Lock()
		defer book.mu.Unlock()
		book.published = slices.Clone(snapshot.Notes)
		return nil
	})
}

func echoText(c *sdk.Context, text string) error { return c.Reply(text) }

func (b *notebook) add(c *sdk.Context, text string) error {
	b.mu.Lock()
	b.notes = append(b.notes, text)
	b.mu.Unlock()
	return c.Reply(c.T("echo.added"))
}

func (b *notebook) snapshot() NotesSnapshot {
	b.mu.Lock()
	defer b.mu.Unlock()
	return NotesSnapshot{Notes: append([]string{}, b.notes...)}
}

func (b *notebook) list(c *sdk.Context) error { return replyNotes(c, b.snapshot().Notes) }

func (b *notebook) listPublished(c *sdk.Context) error {
	b.mu.Lock()
	notes := slices.Clone(b.published)
	b.mu.Unlock()
	return replyNotes(c, notes)
}

func replyNotes(c *sdk.Context, notes []string) error {
	if len(notes) == 0 {
		return c.Reply(c.T("echo.empty"))
	}
	return c.Reply(strings.Join(notes, "\n"))
}

func calculate(c *sdk.Context, args Calculation) (message.Content, error) {
	if args.B > 0 && args.A > math.MaxInt64-args.B || args.B < 0 && args.A < math.MinInt64-args.B {
		return nil, errcode.InvalidArgument.New("sum overflows int64")
	}
	return message.Text(strconv.FormatInt(args.A+args.B, 10)), nil
}

func ask(c *sdk.Context) error {
	if err := c.Reply(c.T("echo.question")); err != nil {
		return err
	}
	next, err := c.Wait(askTimeout)
	if err != nil {
		return err
	}
	return c.Reply(c.T("echo.answer", "text", next.Text()))
}
