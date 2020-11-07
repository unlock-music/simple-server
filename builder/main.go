package main

import (
	"archive/tar"
	"compress/gzip"
	"github.com/tidwall/gjson"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"
)

const buildTempDir = "./build"
const checkVersionUrl = "https://api.github.com/repos/ix64/unlock-music/releases/latest"
const assetFilename = "legacy.tar.gz"
const checksumFilename = "sha256sum.txt"

func main() {
	if err := checkTempDirExist(); err != nil {
		log.Fatal(err)
	}
	log.Println("gathering version info: " + checkVersionUrl)
	v, err := getLatestVersionInfo()
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < 3; i++ {

		if !v.checkAssetExist() {
			log.Printf("downloading %s to %s\n", v.AssetUrl, v.getAssetPath(""))
			if err := v.downloadAsset(); err != nil {
				log.Fatal(err)
			}
		}
		log.Printf("gathering checksum info: %s\n", v.ChecksumUrl)
		expect, err := v.downloadChecksum()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("checksum of %s should be: %s\n", assetFilename, expect)

		actual, err := v.calcAssetChecksum()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("checksum of %s is: %s\n", assetFilename, actual)

		if expect != actual {
			newFilename := v.getAssetPath("unexpected-" + strconv.FormatInt(time.Now().Unix(), 10) + "-")
			if err := os.Rename(v.getAssetPath(""), newFilename); err != nil {
				log.Fatal(err)
			}
		} else {
			if err := unArchive(v.getAssetPath(""), path.Join(buildTempDir, "for-build")); err != nil {
				log.Fatal(err)
			}
			return
		}
	}
	log.Fatal("failed for 3 times")
}

func checkTempDirExist() error {
	_, err := os.Stat(buildTempDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(buildTempDir, 0755)
	}
	return err
}

func getLatestVersionInfo() (info *versionInfo, err error) {
	resp, err := http.Get(checkVersionUrl)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	asset := gjson.GetBytes(respBody, "assets.#(name="+assetFilename+")")
	checksum := gjson.GetBytes(respBody, "assets.#(name="+checksumFilename+")")
	return &versionInfo{
		Version:     gjson.GetBytes(respBody, "tag_name").String(),
		ChecksumUrl: checksum.Get("browser_download_url").String(),
		AssetUrl:    asset.Get("browser_download_url").String(),
		AssetSize:   asset.Get("size").Int(),
	}, nil
}
func unArchive(source string, destination string) error {
	src, err := os.Open(source)
	if err != nil {
		return nil
	}
	defer src.Close()
	uncompressed, err := gzip.NewReader(src)
	if err != nil {
		return nil
	}
	arc := tar.NewReader(uncompressed)
	for {
		var f *tar.Header
		f, err = arc.Next()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		if f.FileInfo().IsDir() {
			err = os.MkdirAll(path.Join(destination, f.Name), 0755)
			if err != nil {
				return err
			}
		} else {
			dst, err := os.OpenFile(path.Join(destination, f.Name), os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				return err
			}
			_, err = io.CopyN(dst, arc, f.Size)
			dst.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
