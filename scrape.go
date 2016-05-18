package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const SpoofedUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_3) " +
	"AppleWebKit/537.75.14 (KHTML, like Gecko) Version/7.0.3 Safari/7046A194A"

func Scrape(inputFile, outputDir string) error {
	var list []*StoryItem

	inputData, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(inputData, &list); err != nil {
		return err
	}

	if err := os.Mkdir(outputDir, 0755); err != nil {
		return err
	}

	for _, post := range list {
		body, err := fetchArticleBody(post.URL)
		if err != nil {
			log.Printf("Error fetching %s: %s", post.URL, err.Error())
		} else {
			fileData := []byte(post.Title + "\n\n" + body)
			postName := strconv.FormatInt(post.ID, 10) + ".txt"
			path := filepath.Join(outputDir, postName)
			if err := ioutil.WriteFile(path, fileData, 0755); err != nil {
				return err
			}
		}
	}

	return nil
}

func fetchArticleBody(urlStr string) (string, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", SpoofedUserAgent)

	cookies, _ := cookiejar.New(nil)
	client := http.Client{Jar: cookies}

	resp, err := client.Do(req)
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
