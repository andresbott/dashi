package xkcd

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const defaultBaseURL = "https://xkcd.com"
const defaultCacheTTL = 1 * time.Hour

type Comic struct {
	Num       int    `json:"num"`
	Title     string `json:"title"`
	SafeTitle string `json:"safe_title"`
	Img       string `json:"img"`
	Alt       string `json:"alt"`
	Day       string `json:"day"`
	Month     string `json:"month"`
	Year      string `json:"year"`
}

type cacheEntry struct {
	Comic   Comic     `json:"comic"`
	Fetched time.Time `json:"fetched"`
}

type Client struct {
	httpClient *http.Client
	baseURL    string
	cacheDir   string
	cacheTTL   time.Duration
	nowFn      func() time.Time
}

func NewClient(cacheDir string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    defaultBaseURL,
		cacheDir:   cacheDir,
		cacheTTL:   defaultCacheTTL,
		nowFn:      time.Now,
	}
}

// SetBaseURL overrides the base URL (for testing).
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

func (c *Client) GetComic(num int) (Comic, error) {
	cacheFile := filepath.Join(c.cacheDir, fmt.Sprintf("%d.json", num))

	if comic, err := c.readCache(cacheFile); err == nil {
		return comic, nil
	}

	url := fmt.Sprintf("%s/%d/info.0.json", c.baseURL, num)
	comic, err := c.fetch(url)
	if err != nil {
		return Comic{}, fmt.Errorf("fetching comic %d: %w", num, err)
	}

	c.writeCache(cacheFile, comic)
	return comic, nil
}

func (c *Client) GetLatest() (Comic, error) {
	cacheFile := filepath.Join(c.cacheDir, "latest.json")

	if comic, err := c.readCache(cacheFile); err == nil {
		return comic, nil
	}

	url := fmt.Sprintf("%s/info.0.json", c.baseURL)
	comic, err := c.fetch(url)
	if err != nil {
		return Comic{}, fmt.Errorf("fetching latest comic: %w", err)
	}

	c.writeCache(cacheFile, comic)
	return comic, nil
}

func (c *Client) GetDailyRandom() (Comic, error) {
	latest, err := c.GetLatest()
	if err != nil {
		return Comic{}, fmt.Errorf("fetching latest for daily random: %w", err)
	}

	num := dailyComicNum(c.nowFn(), latest.Num)
	return c.GetComic(num)
}

func (c *Client) GetRandom() (Comic, error) {
	latest, err := c.GetLatest()
	if err != nil {
		return Comic{}, fmt.Errorf("fetching latest for random: %w", err)
	}
	if latest.Num <= 1 {
		return latest, nil
	}
	num := rand.IntN(latest.Num) + 1 //nolint:gosec // not security-sensitive
	return c.GetComic(num)
}

func dailyComicNum(t time.Time, maxNum int) int {
	if maxNum <= 1 {
		return 1
	}
	dateStr := t.Format("2006-01-02")
	hash := sha256.Sum256([]byte(dateStr))
	n := binary.BigEndian.Uint64(hash[:8])
	return int(n%uint64(maxNum)) + 1 //nolint:gosec // maxNum is a comic number, always fits in int
}

func (c *Client) fetch(url string) (Comic, error) {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return Comic{}, fmt.Errorf("xkcd API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return Comic{}, fmt.Errorf("xkcd API returned status %d", resp.StatusCode)
	}

	var comic Comic
	if err := json.NewDecoder(resp.Body).Decode(&comic); err != nil {
		return Comic{}, fmt.Errorf("decoding xkcd response: %w", err)
	}
	return comic, nil
}

func (c *Client) readCache(path string) (Comic, error) {
	entry, err := readCacheFile(path)
	if err != nil {
		return Comic{}, err
	}
	if c.nowFn().Sub(entry.Fetched) > c.cacheTTL {
		return Comic{}, fmt.Errorf("cache expired")
	}
	return entry.Comic, nil
}

func readCacheFile(path string) (cacheEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return cacheEntry{}, err
	}
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return cacheEntry{}, err
	}
	return entry, nil
}

func (c *Client) writeCache(path string, comic Comic) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return
	}
	entry := cacheEntry{Comic: comic, Fetched: c.nowFn()}
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	tmpFile := path + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0o600); err != nil {
		return
	}
	_ = os.Rename(tmpFile, path)
}
