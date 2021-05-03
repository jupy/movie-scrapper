package main

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"
	"os/exec"
	"encoding/json"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type Movie struct {
	Name          string
	InitName      string
	PosterUrl     string
	Year          string
	Genres        []string
	Directors     []string
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
	for _, d := range movie.Directors {
		fmt.Printf("Director:       %s\n", d)
	}
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

func firstRune(str string) (r rune) {
	for _, r = range str {
		return
	}
	return
}

func ParseList(html string) []string {
	var list []string
	lines := strings.Split(html, "<br/>")
	for _, line := range lines {
		dom, err := goquery.NewDocumentFromReader(strings.NewReader(line))
		if err == nil {
			dom.Each(func(i int, sel *goquery.Selection) {
				div := regexp.MustCompile(`[;,:\[\]]`)
				for _, item := range div.Split(sel.Text(), -1) {
					item = strings.Trim(item, " \t")
					if len(item) == 0 {
						continue
					} else if unicode.IsUpper(firstRune(item)) {
						list = append(list, item)
					}
				}
			})
		}
	}
	//fmt.Printf("item: %v\n", list)
	return list
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
				html, _ := a.DOM.Html()
				movie.Directors = ParseList(html)
			})
		}
		if title == "Продюсер" {
			e.ForEach("td span", func(i int, a *colly.HTMLElement) {
				html, _ := a.DOM.Html()
				movie.Producers = ParseList(html)
			})
		}
		if strings.Contains(title, "Автор") && strings.Contains(title, "сценария") {
			e.ForEach("td span > a", func(i int, a *colly.HTMLElement) {
				html, _ := a.DOM.Html()
				movie.Screenwriters = ParseList(html)
			})
		}
		if title == "Кинокомпания" || title == "Студия" {
			e.ForEach("td span * ", func(i int, a *colly.HTMLElement) {
				html, _ := a.DOM.Html()
				movie.Companies = ParseList(html)
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
	s := "-w " + site
	out, err := exec.Command("googler", "-n 1", "--np", "--json", s, query).Output()
	if err != nil {
		log.Fatal(err)
	    panic(err)
	}

	var dat []map[string]interface{}

	if err := json.Unmarshal(out, &dat); err != nil {
	    panic(err)
	}

	var str string
	for _, d := range dat {
		str, _ = url.QueryUnescape(d["url"].(string))
	}
	return str
}

func ScrapeMovieInner(wikipedia string, kinopoisk string, mail string) Movie {
	var movie Movie

	w, _ := url.QueryUnescape(wikipedia)
	movie = VisitWikipedia(w)
	movie.WikipediaUrl = w
	movie.KinopoiskUrl = kinopoisk
	movie.MailUrl = mail
	movie.Summary = VisitKinoMail(mail)
	return movie
}

func ScrapeMovie(query string) Movie {
	wikipedia := SearchGoogle(query, "ru.wikipedia.org")
	time.Sleep(1 * time.Second)
	kinopoisk := SearchGoogle(query, "kinopoisk.ru")
	time.Sleep(1 * time.Second)
	mail := SearchGoogle(query, "kino.mail.ru")
	time.Sleep(1 * time.Second)

	return ScrapeMovieInner(wikipedia, kinopoisk, mail)
}

/*type SearchWikiResult struct {
	Url     string
	Title   string
	Summary string
}

func SearchWiki(query string) string {
	var res []SearchWikiResult

	c := colly.NewCollector(
		colly.AllowedDomains("en.wikipedia.org"),
	)

	c.Limit(&colly.LimitRule{
		// Filter domains affected by this rule
		DomainGlob: "en.wikipedia.org/*",
		// Set a delay between requests to these domains
		Delay: 1 * time.Second,
		// Add an additional random delay
		RandomDelay: 1 * time.Second,
	})

	c.OnHTML(".jump-to-nav", func(e *colly.HTMLElement) {
		fmt.Printf("!!! \n")
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Printf("Something went wrong: %v\n", err)
	})

	q := "https://en.wikipedia.org/w/index.php?search=" + query + "&ns0=1"
	fmt.Printf("query: %s\n", q)
	c.Visit(q)
	fmt.Printf("finish\n")

	for _, r := range res {
		fmt.Printf("Title: %s\n", r.Title)
		fmt.Printf("Url:   %s\n", r.Url)
	}

	return q
}*/

func main() {
	var movie Movie

   	movie = ScrapeMovie("9 ярдов")
  	movie.Print()

	fmt.Println("---")

  	movie = ScrapeMovie("Крепкий орешек 1988")
  	movie.Print()

	fmt.Println("---")

  	movie = ScrapeMovie("Белоснежка (1937)")
  	movie.Print()
}
