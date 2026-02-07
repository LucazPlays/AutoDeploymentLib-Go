package autodeployment

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type ReleaseInfo struct {
	LastModifiedEpochMs int64  `json:"lastModifiedEpochMs"`
	DownloadURL         string `json:"downloadUrl"`
	SHA256              string `json:"sha256"`
}

type Updater struct {
	apiRoot          string
	updateInterval   time.Duration
	projectUUID      string
	projectKey       string
	running          bool
	stopChan         chan struct{}
	serverTimeOffset int64
}

func New(uuid, key string) *Updater {
	return &Updater{
		apiRoot:        "https://api.insights-api.top/deployment/",
		updateInterval: 30 * time.Second,
		projectUUID:    uuid,
		projectKey:     key,
		stopChan:       make(chan struct{}),
	}
}

func (u *Updater) SetAPIRoot(apiRoot string) {
	apiRoot = strings.TrimSuffix(apiRoot, "/api")
	apiRoot = strings.TrimSuffix(apiRoot, "/")
	u.apiRoot = apiRoot
}

func (u *Updater) SetUpdateInterval(interval time.Duration) {
	u.updateInterval = interval
}

func (u *Updater) Start() error {
	if u.projectUUID == "" || u.projectKey == "" {
		return fmt.Errorf("missing project UUID or key")
	}

	u.SyncTime()
	u.running = true
	go u.loop()
	return nil
}

func (u *Updater) Stop() {
	u.running = false
	close(u.stopChan)
}

func (u *Updater) SyncTime() {
	serverTime := u.GetServerTime()
	if serverTime > 0 {
		u.serverTimeOffset = serverTime - time.Now().UnixMilli()
	}
}

func (u *Updater) GetServerTime() int64 {
	resp, err := http.Get(u.apiRoot + "/api/public/time")
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	var result struct {
		CurrentEpochMs int64 `json:"currentEpochMs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0
	}
	return result.CurrentEpochMs
}

func (u *Updater) GetLocalTime() int64 {
	return time.Now().UnixMilli()
}

func (u *Updater) GetTimeDiff() int64 {
	return u.serverTimeOffset
}

func (u *Updater) GetAdjustedLocalTime() int64 {
	return time.Now().UnixMilli() + u.serverTimeOffset
}

type TimeInfo struct {
	ServerTime        int64 `json:"serverTime"`
	LocalTime         int64 `json:"localTime"`
	AdjustedLocalTime int64 `json:"adjustedLocalTime"`
	TimeDiff          int64 `json:"timeDiff"`
}

func (u *Updater) GetTimeInfo() TimeInfo {
	return TimeInfo{
		ServerTime:        u.GetServerTime(),
		LocalTime:         u.GetLocalTime(),
		AdjustedLocalTime: u.GetAdjustedLocalTime(),
		TimeDiff:          u.GetTimeDiff(),
	}
}

func (u *Updater) loop() {
	ticker := time.NewTicker(u.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			u.checkAndUpdate()
		case <-u.stopChan:
			return
		}
	}
}

func (u *Updater) checkAndUpdate() {
	selfPath, err := os.Executable()
	if err != nil {
		return
	}

	release, err := u.fetchReleaseInfo()
	if err != nil {
		return
	}

	if release.SHA256 == "" {
		return
	}

	info, statErr := os.Stat(selfPath)
	if statErr != nil {
		return
	}

	localMtime := info.ModTime().UnixMilli()
	adjustedMtime := localMtime + u.serverTimeOffset

	if adjustedMtime >= release.LastModifiedEpochMs {
		return
	}

	downloadURL := u.resolveURL(release.DownloadURL)
	tmpPath := selfPath + ".download"
	os.Remove(tmpPath)

	if err := u.download(downloadURL, tmpPath); err != nil {
		os.Remove(tmpPath)
		return
	}

	sha256Hash, err := calculateSHA256(tmpPath)
	if err != nil {
		os.Remove(tmpPath)
		return
	}

	if !strings.EqualFold(release.SHA256, sha256Hash) {
		os.Remove(tmpPath)
		return
	}

	if !u.verify(sha256Hash) {
		os.Remove(tmpPath)
		return
	}

	backupPath := selfPath + ".bak"
	os.Remove(backupPath)
	os.Rename(selfPath, backupPath)
	os.Rename(tmpPath, selfPath)

	os.Chmod(selfPath, 0755)

	mtime := time.UnixMilli(release.LastModifiedEpochMs)
	if err := os.Chtimes(selfPath, mtime, mtime); err != nil {
	}

	os.Exit(0)
}

func (u *Updater) fetchReleaseInfo() (*ReleaseInfo, error) {
	reqURL := fmt.Sprintf("%s/api/public/projects/%s/release?key=%s",
		u.apiRoot, url.PathEscape(u.projectUUID), url.PathEscape(u.projectKey))

	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}

	var info ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	if info.LastModifiedEpochMs <= 0 || info.DownloadURL == "" {
		return nil, fmt.Errorf("invalid release info")
	}

	return &info, nil
}

func (u *Updater) verify(sha256 string) bool {
	reqURL := fmt.Sprintf("%s/api/public/projects/%s/verify?key=%s&sha256=%s",
		u.apiRoot, url.PathEscape(u.projectUUID), url.PathEscape(u.projectKey), url.PathEscape(sha256))

	resp, err := http.Get(reqURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return strings.Contains(string(body), `"ok":true`) || strings.Contains(string(body), `"ok": true`)
}

func (u *Updater) download(downloadURL, destPath string) error {
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "AutoDeploymentUpdater/1.0")
	req.Header.Set("X-Project-Key", u.projectKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("http %d", resp.StatusCode)
	}

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	io.Copy(file, resp.Body)
	return nil
}

func (u *Updater) resolveURL(maybeRelative string) string {
	if strings.HasPrefix(maybeRelative, "http://") || strings.HasPrefix(maybeRelative, "https://") {
		return maybeRelative
	}
	if !strings.HasPrefix(maybeRelative, "/") {
		maybeRelative = "/" + maybeRelative
	}
	return u.apiRoot + maybeRelative
}

func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	io.Copy(hash, file)
	return hex.EncodeToString(hash.Sum(nil)), nil
}
