package main

import (
	"github.com/gocolly/colly"
	"github.com/retailerTool/log"
	"github.com/retailerTool/storage"
	"github.com/retailerTool/utils"
	"strconv"
	"strings"
)

var (
	UserAgents = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/12.246",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:62.0) Gecko/20100101 Firefox/62.0",
	}
)

type Job struct {
	url     string
	jobType string
}

var (
	Sell = Job{
		url:     "https://www.ss.lv/ru/real-estate/flats/riga/all/sell/",
		jobType: "sell",
	}
	Rent = Job{
		url:     "https://www.ss.lv/ru/real-estate/flats/riga/all/hand_over/",
		jobType: "rent",
	}
)

type Crawler struct {
	logger      log.Logger
	flatStorage storage.FlatStorage
	userAgent   string
	collector   *colly.Collector
}

func (c *Crawler) Run(job Job) {
	c.collector = colly.NewCollector(
		colly.UserAgent(c.userAgent),
		colly.AllowedDomains("www.ss.lv"),
	)

	c.collector.OnHTML("tr[id]", func(rowElement *colly.HTMLElement) {
		idStr := rowElement.Attr("id")
		if !strings.HasPrefix(idStr, "tr_") {
			return
		}
		if strings.HasPrefix(idStr, "tr_bnr") {
			return
		}
		flat := storage.Flat{}
		flat.Id, _ = strconv.Atoi(rowElement.Attr("id")[3:])
		flat.Type = job.jobType
		rowElement.ForEach("td", func(i int, cellElement *colly.HTMLElement) {
			switch i {
			case 2:
				flat.Text = utils.FilterChars(cellElement.Text, "[\n]")
				cellElement.ForEach("a[href]", func(i int, element *colly.HTMLElement) {
					flat.Url = element.Request.AbsoluteURL(element.Attr("href"))
				})
			case 3:
				locationText, _ := cellElement.DOM.Html()
				locationText = strings.ReplaceAll(locationText, "<b>", "")
				locationText = strings.ReplaceAll(locationText, "</b>", "")
				locationArr := strings.Split(locationText, "<br/>")
				flat.District, flat.Street = locationArr[0], locationArr[1]
			case 4:
				flat.Rooms, _ = strconv.Atoi(cellElement.Text)
			case 5:
				flat.ApartmentArea, _ = strconv.Atoi(cellElement.Text)
			case 6:
				flat.Floor = cellElement.Text
			case 7:
				flat.HouseType = cellElement.Text
			case 8:
				flat.Price, _ = strconv.Atoi(utils.FilterChars(cellElement.Text, "[^0-9]"))
			case 9:
				flat.Price, _ = strconv.Atoi(utils.FilterChars(cellElement.Text, "[^0-9]"))
			}
		})
		c.flatStorage.Put(flat)
	})

	visitedUrls := make([]string, 50)

	c.collector.OnHTML("a[name]", func(element *colly.HTMLElement) {
		url := element.Request.AbsoluteURL(element.Attr("href"))
		if !utils.IsStringInSlice(visitedUrls, url) {
			visitedUrls = append(visitedUrls, url)
			c.logger.Log("Visit " + url)
			c.collector.Visit(url)
		}
	})

	c.collector.Visit(job.url)
}
