package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

type versionInfo struct {
	Version     string
	ChecksumUrl string
	AssetUrl    string
	AssetSize   int64
}

func (v versionInfo) getAssetPath(prefix string) string {
	return path.Join(buildTempDir, prefix+v.Version+"_"+assetFilename)
}

func (v versionInfo) checkAssetExist() bool {
	_, err := os.Stat(v.getAssetPath(""))
	return !os.IsNotExist(err)
}

func (v versionInfo) downloadAsset() (err error) {
	file, err := os.OpenFile(v.getAssetPath(""), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	resp, err := http.Get(v.AssetUrl)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(file, resp.Body)
	return err
}

func (v versionInfo) downloadChecksum() (checksum string, err error) {
	resp, err := http.Get(v.ChecksumUrl)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	r := bufio.NewReader(resp.Body)
	var line []byte
	for err == nil {
		line, _, err = r.ReadLine()
		lineStr := string(line)
		if strings.Contains(lineStr, assetFilename) {
			checksum = strings.ToLower(strings.TrimSpace(strings.Split(lineStr, " ")[0]))
			return
		}
	}
	return "", fmt.Errorf("checksum for %s not found", assetFilename)
}

func (v versionInfo) calcAssetChecksum() (checksum string, err error) {
	file, err := os.Open(v.getAssetPath(""))
	if err != nil {
		return
	}
	defer file.Close()
	sha := sha256.New()
	_, err = io.Copy(sha, file)
	if err != nil {
		return
	}
	checksum = hex.EncodeToString(sha.Sum(nil))
	return
}
