package main

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
	googlesearch "github.com/rocketlaunchr/google-search"
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
			e.ForEach("td span > a", func(i int, a *colly.HTMLElement) {
				movie.Producers = append(movie.Producers, a.Text)
			})
		}
		if strings.Contains(title, "Автор") && strings.Contains(title, "сценария") {
			e.ForEach("td span > a", func(i int, a *colly.HTMLElement) {
				movie.Screenwriters = append(movie.Screenwriters, a.Text)
			})
		}
		if title == "Кинокомпания" {
			e.ForEach("td span * ", func(i int, a *colly.HTMLElement) {
				if a.Text != "" {
					movie.Companies = append(movie.Companies, a.Text)
				}
			})
		}
		if title == "Страна" {
			e.ForEach("td span.country-name", func(i int, a *colly.HTMLElement) {
				if a.Text != "" {
					movie.Countries = append(movie.Countries, a.Text)
				}
			})
			e.ForEach("td span a.mw-redirect", func(i int, a *colly.HTMLElement) {
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

func ScrapeMovie(query string) Movie {
	var movie Movie
	var wikipedia string
	var kinopoisk string
	var mail string

	wikipedia = SearchGoogle(query, "ru.wikipedia.org")
	kinopoisk = SearchGoogle(query, "kinopoisk.ru")
	mail = SearchGoogle(query, "kino.mail.ru")

	movie = VisitWikipedia(wikipedia)
	movie.WikipediaUrl = wikipedia
	movie.KinopoiskUrl = kinopoisk
	movie.MailUrl = mail
	movie.Summary = VisitKinoMail(mail)
	return movie
}

func main() {
	movie := ScrapeMovie("Белоснежка")
	movie.Print()

	/* 	fmt.Println("---")

	   	movie = ScrapeMovie("Крепкий орешек")
	   	movie.Print() */
}
