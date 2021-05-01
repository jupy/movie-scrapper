package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
)

func main() {

	c := colly.NewCollector(
		colly.AllowedDomains("ru.wikipedia.org"),
	)

	c.OnHTML("h1", func(e *colly.HTMLElement) {
		fmt.Printf("Title: %q\n", e.Text)
	})

	c.OnHTML(".infobox tbody tr:nth-child(2)", func(e *colly.HTMLElement) {
		fmt.Printf("Original Title: %q\n", e.Text)
	})

	c.OnHTML(".infobox-image a img[srcset]", func(e *colly.HTMLElement) {
		link := e.Attr("srcset")
		decodedLink, _ := url.QueryUnescape(link)
		v := strings.Split(decodedLink, " ")
		fmt.Printf("Picture: https:%s\n", v[0])
	})

	c.OnHTML(".infobox tbody tr", func(e *colly.HTMLElement) {
		title := e.ChildText("th")
		if title == "Жанр" {
			e.ForEach("td a[href]", func(i int, a *colly.HTMLElement) {
				fmt.Printf("Genre: %s\n", a.Text)
			})
		}
		if title == "Режиссёр" {
			e.ForEach("td span", func(i int, a *colly.HTMLElement) {
				fmt.Printf("Director: %s\n", a.Text)
			})
		}
		if title == "Продюсер" {
			e.ForEach("td span > a", func(i int, a *colly.HTMLElement) {
				fmt.Printf("Producer: %s\n", a.Text)
			})
		}
		if strings.Contains(title, "Автор") && strings.Contains(title, "сценария") {
			e.ForEach("td span > a", func(i int, a *colly.HTMLElement) {
				fmt.Printf("Written by: %s\n", a.Text)
			})
		}
		if title == "Кинокомпания" {
			e.ForEach("td span * ", func(i int, a *colly.HTMLElement) {
				if a.Text != "" {
					fmt.Printf("Company: %s\n", a.Text)
				}
			})
		}
		if title == "Страна" {
			e.ForEach("td span.country-name", func(i int, a *colly.HTMLElement) {
				if a.Text != "" {
					fmt.Printf("Country: %s\n", a.Text)
				}
			})
			e.ForEach("td span a.mw-redirect", func(i int, a *colly.HTMLElement) {
				if a.Text != "" {
					fmt.Printf("Country: %s\n", a.Text)
				}
			})
		}
		if title == "Год" {
			r, _ := regexp.Compile("[0-9][0-9][0-9][0-9]")
			e.ForEachWithBreak("td a", func(i int, a *colly.HTMLElement) bool {
				s := r.FindString(a.Text)
				if s != "" {
					fmt.Printf("Year: %s\n", s)
					return false
				}
				return true
			})
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit("https://ru.wikipedia.org/wiki/Девять_ярдов")
	c.Visit("https://ru.wikipedia.org/wiki/Крепкий_орешек_(фильм,_1988)")
}
