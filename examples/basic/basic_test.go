package basic

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type echo func(context.Context, *Input) (*Output, error)

func (e echo) Repeat(ctx context.Context, i *Input) (*Output, error) {
	return e(ctx, i)
}

func assertResponseMap(tb testing.TB, resp *http.Response, s int, fn func(tb testing.TB, v map[string]interface{})) {
	if resp.StatusCode != s {
		tb.Fatalf("unexpected status code: %d", resp.StatusCode)
	}

	v := map[string]interface{}{}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	err := dec.Decode(&v)
	if err != nil {
		tb.Fatalf("failed to decode json: %v", err)
	}

	fn(tb, v)
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

	assertResponseMap(t, resp, http.StatusOK, func(tb testing.TB, v map[string]interface{}) {
		if v["message"] != "foo" {
			tb.Fatalf("expected this output, got: %#v", v)
		}
	})
}

func TestPostJSON(t *testing.T) {
	var e EchoServer
	e = echo(func(ctx context.Context, i *Input) (*Output, error) {
		if i.Message == "" {
			i.Message = "empty_input"
		}
		return &Output{Message: i.Message, OverwriteMe: i.OverwriteMe}, nil
	})

	echoHandler := NewEchoHandler(e)
	svr := httptest.NewServer(http.HandlerFunc(echoHandler.HandleRepeat))
	defer svr.Close()

	resp, err := svr.Client().Post(svr.URL, "application/json", bytes.NewBufferString(`{"message": "bar"}`))
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}

	assertResponseMap(t, resp, http.StatusOK, func(tb testing.TB, v map[string]interface{}) {
		if v["message"] != "bar" {
			tb.Fatalf("expected this output, got: %#v", v)
		}
	})

	t.Run("weird content-types", func(t *testing.T) {
		resp, err := svr.Client().Post(svr.URL, "applicaTION/json; charset=utf-8", bytes.NewBufferString(`{"message": "bar"}`))
		if err != nil {
			t.Fatalf("failed to get: %v", err)
		}

		assertResponseMap(t, resp, http.StatusOK, func(tb testing.TB, v map[string]interface{}) {
			if v["message"] != "bar" {
				tb.Fatalf("expected this output, got: %#v", v)
			}
		})
	})

	t.Run("empty content type", func(t *testing.T) {
		resp, err := svr.Client().Post(svr.URL, "", bytes.NewBufferString(`{"message": "bar"}`))
		if err != nil {
			t.Fatalf("failed to get: %v", err)
		}

		assertResponseMap(t, resp, http.StatusOK, func(tb testing.TB, v map[string]interface{}) {
			if v["message"] != "empty_input" {
				tb.Fatalf("expected this output, got: %#v", v)
			}
		})
	})

	t.Run("dont overwite if body doesn't specifys", func(t *testing.T) {
		resp, err := svr.Client().Post(svr.URL+"?OverwriteMe=true", "application/json", bytes.NewBufferString(`{"message": "bar"}`))
		if err != nil {
			t.Fatalf("failed to get: %v", err)
		}

		assertResponseMap(t, resp, http.StatusOK, func(tb testing.TB, v map[string]interface{}) {
			if v["overwriteMe"] != true {
				tb.Fatalf("expected query input not to be overwritten, got: %#v", v)
			}

			if v["message"] != "bar" {
				tb.Fatalf("expected this output, got: %#v", v)
			}
		})
	})

	t.Run("do overwite if body specifies", func(t *testing.T) {
		resp, err := svr.Client().Post(svr.URL+"?overwriteMe=true", "application/json", bytes.NewBufferString(`{"message": "bar", "overwriteMe": false}`))
		if err != nil {
			t.Fatalf("failed to get: %v", err)
		}

		assertResponseMap(t, resp, http.StatusOK, func(tb testing.TB, v map[string]interface{}) {
			if _, ok := v["overwriteMe"]; ok {
				tb.Fatalf("expected overwrite to cause empty message field, got: %#v", v)
			}

			if v["message"] != "bar" {
				tb.Fatalf("expected this output, got: %#v", v)
			}
		})
	})
}

func TestGetQueryParams(t *testing.T) {
	var e EchoServer
	e = echo(func(ctx context.Context, i *Input) (*Output, error) {
		return &Output{Message: i.Message}, nil
	})

	echoHandler := NewEchoHandler(e)
	svr := httptest.NewServer(http.HandlerFunc(echoHandler.HandleRepeat))
	defer svr.Close()

	resp, err := svr.Client().Get(svr.URL + "?message=foobar")
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}

	assertResponseMap(t, resp, http.StatusOK, func(tb testing.TB, v map[string]interface{}) {
		if v["message"] != "foobar" {
			tb.Fatalf("expected this output, got: %#v", v)
		}
	})
}

func TestPostForm(t *testing.T) {
	var e EchoServer
	e = echo(func(ctx context.Context, i *Input) (*Output, error) {
		return &Output{Message: i.Message}, nil
	})

	echoHandler := NewEchoHandler(e)
	svr := httptest.NewServer(http.HandlerFunc(echoHandler.HandleRepeat))
	defer svr.Close()

	vals := url.Values{}
	vals.Set("message", "foobar")

	resp, err := svr.Client().PostForm(svr.URL, vals)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}

	assertResponseMap(t, resp, http.StatusOK, func(tb testing.TB, v map[string]interface{}) {
		if v["message"] != "foobar" {
			tb.Fatalf("expected this output, got: %#v", v)
		}
	})
}

func TestBasicClient(t *testing.T) {
	var e EchoServer
	e = echo(func(ctx context.Context, i *Input) (*Output, error) {
		return &Output{Message: i.Message}, nil
	})

	echoHandler := NewEchoHandler(e)
	svr := httptest.NewServer(http.HandlerFunc(echoHandler.HandleRepeat))
	defer svr.Close()

	caller, err := NewEchoCaller(http.DefaultClient, svr.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	input := &Input{Message: "abc"}
	output, err := caller.CallRepeat(context.Background(), http.MethodPost, "/", input)
	if err != nil {
		t.Fatalf("failed to call server: %v", err)
	}

	if output.Message != input.Message {
		t.Fatalf("output message should be equal to input message, in: %#v, out: %#v", input, output)
	}
}
