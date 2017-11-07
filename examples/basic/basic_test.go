package basic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type echo func(context.Context, *Input) (*Output, error)

func (e echo) Repeat(ctx context.Context, i *Input) (*Output, error) {
	return e(ctx, i)
}

func TestPlainGet(t *testing.T) {
	var e EchoServer
	e = echo(func(context.Context, *Input) (*Output, error) {
		return &Output{Message: "foo"}, nil
	})

	echoHandler := NewEchoHandler(e)
	svr := httptest.NewServer(http.HandlerFunc(echoHandler.HandleRepeat))
	defer svr.Close()

	resp, err := svr.Client().Get(svr.URL)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}

	v := map[string]interface{}{}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&v)
	if err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}

	if v["message"] != "foo" {
		t.Fatalf("expected this output, got: %#v", v)
	}
}
