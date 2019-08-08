package main

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"sync"

	"github.com/antchfx/htmlquery"
)

var (
	reQuote = regexp.MustCompile("^(\\d+. )|(\\n)|(Click to tweet)")
	reKey   = regexp.MustCompile("[^a-zA-Z0-9]+")
	threads = 8
)

type entry struct {
	k interface{}
	v interface{}
}

// QuotesDB contains all quotes
type QuotesDB struct {
	db     map[string][]string
	themes map[string]string
}

func newQuotesDB(data map[string][]string) QuotesDB {
	db := make(map[string][]string)
	titles := make(map[string]string)

	for k, v := range data {
		key := reKey.ReplaceAllString(strings.ToLower(k), "")
		titles[key] = k
		db[key] = v
	}

	return QuotesDB{db, titles}
}

// GetThemes returns list of themes
func (q *QuotesDB) GetThemes() map[string]string {
	return q.themes
}

// GetRandomQuoteByTheme returns random quote from certain theme
func (q *QuotesDB) GetRandomQuoteByTheme(theme string) (string, error) {
	var res string
	if val, ok := q.db[theme]; ok {
		res = val[rand.Intn(len(val))]
	} else {
		return "", fmt.Errorf("GetRandomQuoteByTheme: theme %s not found", theme)
	}
	return res, nil
}

// GetRandomQuote returns random quote
func (q *QuotesDB) GetRandomQuote() string {
	res, _ := q.GetRandomQuoteByTheme(q.getRandomTheme())
	return res
}

func (q *QuotesDB) getRandomTheme() string {
	var res string
	for k := range q.db {
		res = k
		break
	}
	return res
}

// Parse returns QuotesDB with all data
func Parse() (QuotesDB, error) {

	doc, err := htmlquery.LoadURL("http://wisdomquotes.com/")
	if err != nil {
		return QuotesDB{}, fmt.Errorf("Parse: %v", err)
	}

	jobs := make(chan entry)
	results := make(chan entry)
	result := make(chan map[string][]string, 1)
	var wg sync.WaitGroup

	go handler(results, result)

	for i := 0; i < threads; i++ {
		go worker(i, &wg, jobs, results)
	}

	for _, n := range htmlquery.Find(doc, "//div[@class='homepagelinks']/p/a") {
		jobs <- entry{htmlquery.InnerText(n), htmlquery.SelectAttr(n, "href")}
		wg.Add(1)
	}
	close(jobs)

	wg.Wait()

	close(results)
	return newQuotesDB(<-result), nil
}

func worker(id int, wg *sync.WaitGroup, jobs <-chan entry, results chan<- entry) {
	for job := range jobs {
		doc, err := htmlquery.LoadURL(job.v.(string))
		if err != nil {
			log.Printf("worker #%d: processing %s: error: %v\n", id, job.k, err)
			wg.Done()
			continue
		}
		var quotes []string
		for _, n := range htmlquery.Find(doc, "//div/blockquote/p") {
			quotes = append(quotes, strings.Trim(reQuote.ReplaceAllString(htmlquery.InnerText(n), ""), " "))
		}
		results <- entry{job.k, quotes}
		log.Printf("worker %d: processing %s: done\n", id, job.k)
		wg.Done()
	}
}

func handler(results <-chan entry, result chan<- map[string][]string) {
	data := make(map[string][]string)
	for result := range results {
		data[result.k.(string)] = result.v.([]string)
	}
	result <- data
	close(result)
}
