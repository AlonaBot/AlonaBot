package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/AlonaBot/AlonaBot/plugintest"
	"github.com/AlonaBot/AlonaBot/sdk/sqlitedb"
)

func TestConfigDB(t *testing.T) {
	p := New()
	var ctx context.Context
	p.OnReady(func(ready context.Context) error { ctx = context.WithoutCancel(ready); return nil })
	h := plugintest.Start(t, plugintest.Plugin(p), plugintest.Config(p, `{"greeting":"Configured hello"}`))
	user := plugintest.User{ID: "alice"}
	for _, scenario := range []struct{ command, want string }{
		{"/note hello Alice", "Hello, Alice"},
		{`/note set lunch "noodle soup"`, "Saved lunch"},
		{"/note get lunch", "lunch: noodle soup"},
	} {
		result := h.Text(plugintest.Private(user.ID), user, scenario.command)
		if len(result.Sends) != 1 || result.Sends[0].Text() != scenario.want || len(result.Outcomes) != 1 || result.Outcomes[0].Err != nil {
			t.Fatalf("%s = %+v", scenario.command, result)
		}
	}
	db, err := sqlitedb.Open(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var greeting string
	if err := db.QueryRowContext(ctx, "SELECT value FROM configdb_startup WHERE key = ?", startupKey).Scan(&greeting); err != nil {
		t.Fatal(err)
	}
	if greeting != "Configured hello" || p.Config().String("greeting") != greeting {
		t.Fatalf("configured greeting = %q", greeting)
	}
	var sequence int
	var name, path string
	if err := db.QueryRowContext(ctx, "PRAGMA database_list").Scan(&sequence, &name, &path); err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "dev.sqlite" {
		t.Fatalf("managed database = %q", path)
	}
}
