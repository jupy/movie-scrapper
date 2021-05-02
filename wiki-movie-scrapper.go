package main

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	googlesearch "github.com/rocketlaunchr/google-search"
	"golang.org/x/time/rate"
)

type Movie struct {
	Name          string
	InitName      string
	PosterUrl     string
	Year          string
	Genres        []string
	Director      string
	Producers     []string
	Screenwriters []string
	Countries     []string
	Companies     []string
	Summary       string
	WikipediaUrl  string
	KinopoiskUrl  string
	MailUrl       string
}

func (movie *Movie) Print() {
	fmt.Printf("Name:           %s\n", movie.Name)
	fmt.Printf("Original Title: %s\n", movie.InitName)
	fmt.Printf("Picture:        %s\n", movie.PosterUrl)
	for _, a := range movie.Genres {
		fmt.Printf("Genre:          %s\n", a)
	}
	fmt.Printf("Director:       %s\n", movie.Director)
	for _, p := range movie.Producers {
		fmt.Printf("Producer:       %s\n", p)
	}
	for _, s := range movie.Screenwriters {
		fmt.Printf("Screenwriter:   %s\n", s)
	}
	for _, company := range movie.Companies {
		fmt.Printf("Company:        %s\n", company)
	}
	for _, c := range movie.Countries {
		fmt.Printf("Country:        %s\n", c)
	}
	fmt.Printf("Wikipedia:      %s\n", movie.WikipediaUrl)
	fmt.Printf("Kinopoisk:      %s\n", movie.KinopoiskUrl)
	fmt.Printf("Mail:           %s\n", movie.MailUrl)
	fmt.Printf("Summary:\n")
	fmt.Printf("%s\n", movie.Summary)
}

func VisitWikipedia(link string) Movie {

	var movie Movie

	movie.WikipediaUrl = link

	c := colly.NewCollector(
		colly.AllowedDomains("ru.wikipedia.org"),
	)

	c.Limit(&colly.LimitRule{
		// Filter domains affected by this rule
		DomainGlob: "ru.wikipedia.org/*",
		// Set a delay between requests to these domains
		Delay: 1 * time.Second,
		// Add an additional random delay
		RandomDelay: 1 * time.Second,
	})

	c.OnHTML(".infobox tbody tr:nth-child(1)", func(e *colly.HTMLElement) {
		movie.Name = e.Text
	})

	c.OnHTML(".infobox tbody tr:nth-child(2)", func(e *colly.HTMLElement) {
		movie.InitName = e.Text
	})

	c.OnHTML(".infobox-image a img[srcset]", func(e *colly.HTMLElement) {
		link := e.Attr("srcset")
		decodedLink, _ := url.QueryUnescape(link)
		v := strings.Split(decodedLink, " ")
		movie.PosterUrl = "https:" + v[0]
		//fmt.Printf("Picture: https:%s\n", v[0])
	})

	c.OnHTML(".infobox tbody tr", func(e *colly.HTMLElement) {
		title := e.ChildText("th")
		if title == "Жанр" {
			e.ForEach("td a[href]", func(i int, a *colly.HTMLElement) {
				movie.Genres = append(movie.Genres, a.Text)
			})
		}
		if title == "Режиссёр" {
			e.ForEach("td span", func(i int, a *colly.HTMLElement) {
				movie.Director = a.Text
			})
		}
		if title == "Продюсер" {
			e.ForEach("td span", func(i int, a *colly.HTMLElement) {
				movie.Producers = append(movie.Producers, a.Text)
				html, _ := a.DOM.Html()
				lines := strings.Split(html, "<br>")
				for _, line := range lines {
					dom, err := goquery.NewDocument(line)
					if err != nil {
						fmt.Printf("something went wrong: %v\n", err)
					} else {
						dom.Each(func(i int, sel *goquery.Selection) {
							fmt.Printf("> Item: %s\n", sel.Text)
						})
					}
				}

				/* 				span := a.DOM.Eq(1)
				   				fmt.Printf("> html: %s\n", html)
				   				span.Each(func(i int, sel *goquery.Selection) {
				   					fmt.Printf("> i: %d\n", i)
				   					for _, n := range sel.Nodes {
				   						fmt.Printf("> Item (%s): %s\n", n.Type, n.Data)
				   					}
				   				}) */
			})
		}
		if strings.Contains(title, "Автор") && strings.Contains(title, "сценария") {
			e.ForEach("td span > a", func(i int, a *colly.HTMLElement) {
				movie.Screenwriters = append(movie.Screenwriters, a.Text)
			})
		}
		if title == "Кинокомпания" || title == "Студия" {
			e.ForEach("td span * ", func(i int, a *colly.HTMLElement) {
				if a.Text != "" {
					movie.Companies = append(movie.Companies, a.Text)
				}
			})
		}
		if title == "Страна" {
			e.ForEach("td a", func(i int, a *colly.HTMLElement) {
				if a.Text != "" {
					movie.Countries = append(movie.Countries, a.Text)
				}
			})
		}
		if title == "Год" {
			r, _ := regexp.Compile("[0-9][0-9][0-9][0-9]")
			e.ForEachWithBreak("td a", func(i int, a *colly.HTMLElement) bool {
				s := r.FindString(a.Text)
				if s != "" {
					movie.Year = s
					return false
				}
				return true
			})
		}
	})

	c.Visit(movie.WikipediaUrl)
	return movie
}

func VisitKinoMail(link string) string {

	var summary string

	c := colly.NewCollector(
		colly.AllowedDomains("kino.mail.ru"),
	)

	c.Limit(&colly.LimitRule{
		// Filter domains affected by this rule
		DomainGlob: "kino.mail.ru/*",
		// Set a delay between requests to these domains
		Delay: 1 * time.Second,
		// Add an additional random delay
		RandomDelay: 1 * time.Second,
	})

	c.OnHTML("div.p-movie-info__content p", func(e *colly.HTMLElement) {
		summary = e.Text
	})

	c.Visit(link)
	return summary
}

func SearchGoogle(query string, site string) string {

	ctx := context.Background()
	q := query + "site:" + site
	opts := googlesearch.SearchOptions{
		Limit:     1,
		OverLimit: false,
		UserAgent: "Chrome/61.0.3163.100",
	}

	googlesearch.RateLimit = rate.NewLimiter(1, 3)
	links, err := googlesearch.Search(ctx, q, opts)
	if err != nil {
		fmt.Printf("something went wrong: %v\n", err)
		return ""
	}

	if len(links) < 1 {
		return ""
	}

	link, _ := url.QueryUnescape(links[0].URL)
	return link
}

func ScrapeMovieInner(query string, wikipedia string, kinopoisk string, mail string) Movie {
	var movie Movie

	movie = VisitWikipedia(wikipedia)
	movie.WikipediaUrl = wikipedia
	movie.KinopoiskUrl = kinopoisk
	movie.MailUrl = mail
	movie.Summary = VisitKinoMail(mail)
	return movie
}

func ScrapeMovie(query string) Movie {
	wikipedia := SearchGoogle(query, "ru.wikipedia.org")
	kinopoisk := SearchGoogle(query, "kinopoisk.ru")
	mail := SearchGoogle(query, "kino.mail.ru")

	return ScrapeMovieInner(query, wikipedia, kinopoisk, mail)
}

func main() {
	//w := "https://ru.wikipedia.org/wiki/%D0%91%D0%B5%D0%BB%D0%BE%D1%81%D0%BD%D0%B5%D0%B6%D0%BA%D0%B0_%D0%B8_%D1%81%D0%B5%D0%BC%D1%8C_%D0%B3%D0%BD%D0%BE%D0%BC%D0%BE%D0%B2_(%D0%BC%D1%83%D0%BB%D1%8C%D1%82%D1%84%D0%B8%D0%BB%D1%8C%D0%BC)"
	//w := "https://ru.wikipedia.org/wiki/%D0%9A%D1%80%D0%B5%D0%BF%D0%BA%D0%B8%D0%B9_%D0%BE%D1%80%D0%B5%D1%88%D0%B5%D0%BA_(%D1%84%D0%B8%D0%BB%D1%8C%D0%BC,_1988)"
	w := "https://ru.wikipedia.org/wiki/%D0%94%D0%B5%D0%B2%D1%8F%D1%82%D1%8C_%D1%8F%D1%80%D0%B4%D0%BE%D0%B2"
	movie := ScrapeMovieInner("Белоснежка", w, "", "")
	movie.Print()

	/* 	fmt.Println("---")

	   	movie = ScrapeMovie("Крепкий орешек")
	   	movie.Print() */
}
