package main

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	SpoofedUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_3) " +
		"AppleWebKit/537.75.14 (KHTML, like Gecko) Version/7.0.3 Safari/7046A194A"

	SimultaneousReqCount = 10
	RequestTimeout       = time.Second * 20
)

var scrapeClient http.Client

func Scrape(inputFile, outputDir string) error {
	cookies, _ := cookiejar.New(nil)
	scrapeClient = http.Client{
		Jar: cookies,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: RequestTimeout,
	}

	var list []*StoryItem

	inputData, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(inputData, &list); err != nil {
		return err
	}

	os.Mkdir(outputDir, 0755)

	postChan := make(chan *StoryItem)
	var wg sync.WaitGroup
	for i := 0; i < SimultaneousReqCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for post := range postChan {
				postName := strconv.FormatInt(post.ID, 10) + ".txt"
				postPath := filepath.Join(outputDir, postName)

				if _, err := os.Stat(postPath); err == nil || post.URL == "" {
					continue
				}

				body, err := fetchArticleBody(post.URL)
				if err != nil {
					log.Printf("Error fetching %s: %s", post.URL, err.Error())
				} else {
					fileData := []byte(body)
					ioutil.WriteFile(postPath, fileData, 0755)
				}
			}
		}()
	}

	for _, post := range list {
		postChan <- post
	}
	close(postChan)

	wg.Wait()

	return nil
}

func fetchArticleBody(urlStr string) (string, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", SpoofedUserAgent)
	req.Close = true

	resp, err := scrapeClient.Do(req)
	if err != nil {
		return "", err
	}

	root, err := html.Parse(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}

	paragraphs := scrape.FindAll(root, scrape.ByTag(atom.P))

	var result string

	for i, p := range paragraphs {
		if i != 0 {
			result += "\n\n"
		}
		result += nodeText(p)
	}

	return result, nil
}

func nodeText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var childStrs []string
	child := n.FirstChild
	for child != nil {
		childStrs = append(childStrs, nodeText(child))
		child = child.NextSibling
	}
	return strings.Join(childStrs, "")
}
