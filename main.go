package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	id       string
	title    string
	company  string
	location string
	summary  string
}

var baseURL string = "https://www.indeed.com/jobs?q=python&l=Los+Angeles%2C+CA"

func main() {
	page := 10
	var jobs []extractedJob
	for i := 0; i < page; i++ {
		extractedJobs := getPages(i)
		jobs = append(jobs, extractedJobs...)
	}
	writeJobs(jobs)
}

func writeJobs(jobs []extractedJob) {
	c := make(chan string)
	file, err := os.Create("jobs.csv")
	go checkErr(err, c)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"ID", "Title", "Company", "Location", "Summary"}

	wErr := w.Write(headers)
	go checkErr(wErr, c)
	fmt.Println(<-c)
	fmt.Println(<-c)

	for _, job := range jobs {
		jobSlice := []string{job.id, job.title, job.company, job.location, job.summary}
		jwErr := w.Write(jobSlice)
		go checkErr(jwErr, c)
		fmt.Println(<-c)
	}
	fmt.Println("Done, Extracted : " + strconv.Itoa(len(jobs)) + " jobs")
}

func getPageURL(pageNum int) string {
	pageURL := baseURL + "&start=" + strconv.Itoa(pageNum*10)
	return pageURL
}

func getPages(page int) []extractedJob {
	c := make(chan string)
	c2 := make(chan extractedJob)

	var jobs []extractedJob
	pageURL := getPageURL(page)

	res, err := http.Get(pageURL)
	checkErrorAndCode(err, res, c)

	defer res.Body.Close() // Need to close res.Body

	doc, err := goquery.NewDocumentFromReader(res.Body)
	searchCards := doc.Find(".jobsearch-SerpJobCard")
	searchCards.Each(func(i int, s *goquery.Selection) {
		go extractJob(s, c2)
	})

	for i := 0; i < searchCards.Length(); i++ {
		job := <-c2
		jobs = append(jobs, job)
	}
	return jobs
}

func extractJob(s *goquery.Selection, c2 chan extractedJob) {
	id, _ := s.Attr("data-jk")
	title := cleanString(s.Find(".title>a").Text())
	company := cleanString(s.Find(".sjcl").Find(".company").Text())
	location := cleanString(s.Find(".sjcl").Find(".location").Text())
	summary := cleanString(s.Find(".summary").Find("li").Text())
	c2 <- extractedJob{id, title, company, location, summary}
}

func checkErrorAndCode(err error, res *http.Response, c chan string) {
	go checkErr(err, c)
	go checkCode(res, c)
	fmt.Println(<-c)
	fmt.Println(<-c)
}

func checkErr(err error, c chan string) {
	if err != nil {
		c <- err.Error()
	}
	c <- "No Error Detected"
}

func checkCode(res *http.Response, c chan string) {
	if res.StatusCode != 200 {
		c <- "Failed To Connect : " + strconv.Itoa(res.StatusCode)
	}
	c <- "Successfully Connected With Status Code : " + strconv.Itoa(res.StatusCode)
}

func cleanString(s string) string {
	spaceCleaned := strings.Fields(strings.TrimSpace(s))
	return strings.Join(spaceCleaned, " ")
}
