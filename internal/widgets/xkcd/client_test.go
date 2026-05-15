package xkcd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func TestClient_GetComic(t *testing.T) {
	response := map[string]any{
		"num":        614,
		"title":      "Woodpecker",
		"safe_title": "Woodpecker",
		"img":        "https://imgs.xkcd.com/comics/woodpecker.png",
		"alt":        "If you don't have an emergency, feel free to make one up.",
		"day":        "9",
		"month":      "7",
		"year":       "2009",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/614/info.0.json" {
			http.Error(w, "not found", 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	cacheDir := t.TempDir()
	client := NewClient(cacheDir)
	client.baseURL = srv.URL

	comic, err := client.GetComic(614)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comic.Num != 614 {
		t.Errorf("num = %d, want 614", comic.Num)
	}
	if comic.Title != "Woodpecker" {
		t.Errorf("title = %q, want Woodpecker", comic.Title)
	}
	if comic.Img != "https://imgs.xkcd.com/comics/woodpecker.png" {
		t.Errorf("img = %q, want woodpecker URL", comic.Img)
	}
	if comic.Alt != "If you don't have an emergency, feel free to make one up." {
		t.Errorf("alt = %q", comic.Alt)
	}
}

func TestClient_GetComic_CachesOnDisk(t *testing.T) {
	callCount := 0
	response := map[string]any{
		"num": 614, "title": "Woodpecker", "safe_title": "Woodpecker",
		"img": "https://imgs.xkcd.com/comics/woodpecker.png", "alt": "alt text",
		"day": "9", "month": "7", "year": "2009",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	cacheDir := t.TempDir()
	client := NewClient(cacheDir)
	client.baseURL = srv.URL

	_, _ = client.GetComic(614)
	_, _ = client.GetComic(614)

	if callCount != 1 {
		t.Errorf("expected 1 API call (cached on disk), got %d", callCount)
	}

	cachePath := filepath.Join(cacheDir, "614.json")
	if _, err := readCacheFile(cachePath); err != nil {
		t.Errorf("expected cache file at %s: %v", cachePath, err)
	}
}

func TestClient_GetLatest(t *testing.T) {
	response := map[string]any{
		"num": 3228, "title": "Day Counter", "safe_title": "Day Counter",
		"img": "https://imgs.xkcd.com/comics/day_counter.png",
		"alt": "It has been ...", "day": "3", "month": "4", "year": "2026",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/info.0.json" {
			http.Error(w, "not found", 404)
			return
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	cacheDir := t.TempDir()
	client := NewClient(cacheDir)
	client.baseURL = srv.URL

	comic, err := client.GetLatest()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comic.Num != 3228 {
		t.Errorf("num = %d, want 3228", comic.Num)
	}
	if comic.Title != "Day Counter" {
		t.Errorf("title = %q, want Day Counter", comic.Title)
	}
}

func TestClient_GetDailyRandom_Deterministic(t *testing.T) {
	latestResponse := map[string]any{
		"num": 100, "title": "Latest", "safe_title": "Latest",
		"img": "https://imgs.xkcd.com/comics/latest.png", "alt": "alt",
		"day": "1", "month": "1", "year": "2026",
	}
	comicResponse := map[string]any{
		"num": 42, "title": "Geico", "safe_title": "Geico",
		"img": "https://imgs.xkcd.com/comics/geico.jpg", "alt": "alt text",
		"day": "1", "month": "1", "year": "2006",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/info.0.json" {
			_ = json.NewEncoder(w).Encode(latestResponse)
			return
		}
		_ = json.NewEncoder(w).Encode(comicResponse)
	}))
	defer srv.Close()

	fixedTime := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	cacheDir := t.TempDir()
	client := NewClient(cacheDir)
	client.baseURL = srv.URL
	client.nowFn = func() time.Time { return fixedTime }

	comic1, err := client.GetDailyRandom()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cacheDir2 := t.TempDir()
	client2 := NewClient(cacheDir2)
	client2.baseURL = srv.URL
	client2.nowFn = func() time.Time { return fixedTime }

	comic2, err := client2.GetDailyRandom()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if comic1.Num != comic2.Num {
		t.Errorf("expected same comic for same day, got %d and %d", comic1.Num, comic2.Num)
	}
}

func TestClient_GetDailyRandom_DifferentDays(t *testing.T) {
	day1 := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)

	num1 := dailyComicNum(day1, 3000)
	num2 := dailyComicNum(day2, 3000)

	if num1 == num2 {
		t.Errorf("expected different comic numbers for different days, both got %d", num1)
	}
	if num1 < 1 || num1 > 3000 {
		t.Errorf("comic number %d out of range [1, 3000]", num1)
	}
	if num2 < 1 || num2 > 3000 {
		t.Errorf("comic number %d out of range [1, 3000]", num2)
	}
}

func TestClient_GetComic_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", 500)
	}))
	defer srv.Close()

	client := NewClient(t.TempDir())
	client.baseURL = srv.URL

	_, err := client.GetComic(614)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestClient_GetLatest_CachesAsLatest(t *testing.T) {
	callCount := 0
	response := map[string]any{
		"num": 3228, "title": "Day Counter", "safe_title": "Day Counter",
		"img": "https://imgs.xkcd.com/comics/day_counter.png",
		"alt": "alt", "day": "3", "month": "4", "year": "2026",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	client := NewClient(t.TempDir())
	client.baseURL = srv.URL

	_, _ = client.GetLatest()
	_, _ = client.GetLatest()

	if callCount != 1 {
		t.Errorf("expected 1 API call (cached), got %d", callCount)
	}
}
