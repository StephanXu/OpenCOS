package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"flag"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"os"

	"xxtuitui.com/filehasher/quickxorhash"
)

func listFile(path string) ([]string, error) {
	dir, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal("read dir failed")
		return nil, err
	}
	var res []string
	for _, fi := range dir {
		fullName := path + string(os.PathSeparator) + fi.Name()
		if fi.IsDir() {
			fileList, err := listFile(fullName)
			if err != nil {
				log.Fatal("list file error", err)
				return nil, err
			}
			res = append(res, fileList...)
			continue
		}
		res = append(res, fullName)
	}
	return res, nil
}

type HasherConfig struct {
	Hasher hash.Hash
	Name   string
}

type FileHash struct {
	Filename string            `json:"filename"`
	Hashes   map[string]string `json:"hashes"`
}

func MultipleHash(filename string) (map[string]string, error) {
	configs := [...]HasherConfig{
		{Hasher: md5.New(), Name: "md5"},
		{Hasher: sha1.New(), Name: "sha1"},
		{Hasher: sha256.New(), Name: "sha256"},
		{Hasher: sha512.New(), Name: "sha512"},
		{Hasher: quickxorhash.New(), Name: "quickxorhash"},
	}
	var writers []io.Writer
	for _, c := range configs {
		writers = append(writers, c.Hasher)
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	io.Copy(io.MultiWriter(writers...), f)

	res := make(map[string]string)
	for _, h := range configs {
		res[h.Name] = base64.StdEncoding.EncodeToString(h.Hasher.Sum(nil))
	}
	return res, nil
}

func saveResult(obj interface{}, filename string) error {
	fileContent, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		log.Printf("Error encoding result file %v\n", err)
		return err
	}
	ioutil.WriteFile(filename, fileContent, 0666)
	return nil
}

func main() {
	var hashIndexFileName string
	var dirBase string
	var Rebuild bool
	flag.StringVar(&hashIndexFileName, "i", "localhash.json", "Hash index file destination")
	flag.StringVar(&dirBase, "d", "", "Directory of files")
	flag.BoolVar(&Rebuild, "rebuild", false, "Force rebuilding hash index")
	flag.Parse()

	var res []FileHash
	restoredFiles := make(map[string]FileHash)
	if !Rebuild {
		if b, err := os.ReadFile(hashIndexFileName); err == nil {
			var restoredList []FileHash
			if err := json.Unmarshal(b, &restoredList); err == nil {
				for _, item := range restoredList {
					restoredFiles[item.Filename] = item
				}
				log.Printf("Restored %d hash of local files.\n", len(restoredFiles))
			} else {
				log.Printf("Restore hashes failed: %v\n", err)
			}
		} else {
			log.Printf("Can't find cache of hashes.\n")
		}
	} else {
		log.Printf("Force rebuilding hash index\n")
	}

	fileList, err := listFile(dirBase)
	if err != nil {
		log.Printf("List file failed: %v\n", err)
		return
	}
	for _, filename := range fileList {
		if restored, is_restored := restoredFiles[filename]; is_restored {
			res = append(res, restored)
			log.Printf("Restored: %v\n", filename)
			continue
		}
		hash, err := MultipleHash(filename)
		if err != nil {
			log.Printf("Hash file %s failed: %v", filename, err)
			return
		}
		res = append(res, FileHash{Filename: filename, Hashes: hash})
		log.Printf("Produced: %v\n", filename)
	}
	saveResult(res, hashIndexFileName)
}
