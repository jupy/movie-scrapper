package main

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"
	"bufio"
	"os"
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

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func (movie *Movie) PrintMarkdown() {
	name := movie.InitName + " (" + movie.Year + ").md"
	f, err := os.Create(name)
    check(err)
    defer f.Close()

    w := bufio.NewWriter(f)

    _, err = fmt.Fprintf(w, "---\n")
    check(err)
    _, err = fmt.Fprintf(w, "created: %s\n", time.Now().Format("2006.01.02 15:04"))
    check(err)
    _, err = fmt.Fprintf(w, "alias: \"%s (%s)\"\n", movie.Name, movie.Year)
    check(err)
    _, err = fmt.Fprintf(w, "---\n\n")
    check(err)


    _, err = fmt.Fprintf(w, "<div style=\"float:right; padding: 10px\"><img width=200px src=\"%s\"/></div>\n\n", movie.PosterUrl)
    check(err)
    _, err = fmt.Fprintf(w, "![[movie-beginning.png|50]]\n")
    check(err)
    _, err = fmt.Fprintf(w, "# %s\n", movie.Name)
    check(err)

	_, err = fmt.Fprintf(w, "**original name:** %s\n", movie.InitName)
	check(err)
	_, err = fmt.Fprintf(w, "**year:** #y%s\n", movie.Year)
	check(err)
	_, err = fmt.Fprintf(w, "**type:** #movie\n")
	check(err)
	_, err = fmt.Fprintf(w, "**status:** #inbox\n")
	check(err)

	_, err = fmt.Fprintf(w, "**director:**")
	check(err)
	for _, director := range movie.Directors {
		_, err = fmt.Fprintf(w, " [[%s]]", director)
		check(err)
	}
	_, err = fmt.Fprintf(w, "\n")
	check(err)

	_, err = fmt.Fprintf(w, "**producer:**")
	check(err)
	for _, producer := range movie.Producers {
		_, err = fmt.Fprintf(w, " [[%s]]", producer)
		check(err)
	}
	_, err = fmt.Fprintf(w, "\n")
	check(err)

	_, err = fmt.Fprintf(w, "**screenwriter:**")
	check(err)
	for _, screenwriter := range movie.Screenwriters {
		_, err = fmt.Fprintf(w, " [[%s]]", screenwriter)
		check(err)
	}
	_, err = fmt.Fprintf(w, "\n")
	check(err)

	_, err = fmt.Fprintf(w, "**rate:**\n")
	check(err)

	_, err = fmt.Fprintf(w, "**company:**")
	check(err)
	for _, company := range movie.Companies {
		_, err = fmt.Fprintf(w, " [[%s]]", company)
		check(err)
	}
	_, err = fmt.Fprintf(w, "\n")
	check(err)

	_, err = fmt.Fprintf(w, "**country:**")
	check(err)
	for _, country := range movie.Countries {
		_, err = fmt.Fprintf(w, " [[%s]]", country)
		check(err)
	}
	_, err = fmt.Fprintf(w, "\n")
	check(err)

	_, err = fmt.Fprintf(w, "**tags:**")
	check(err)
	for _, genre := range movie.Genres {
		_, err = fmt.Fprintf(w, " [[%s]]", genre)
		check(err)
	}
	_, err = fmt.Fprintf(w, "\n")
	check(err)

	_, err = fmt.Fprintf(w, "**[wikipedia](%s)**\n", movie.WikipediaUrl)
		check(err)
	_, err = fmt.Fprintf(w, "**[kinopoisk](%s)**\n", movie.KinopoiskUrl)
		check(err)
	_, err = fmt.Fprintf(w, "**[kino.mail](%s)**\n", movie.MailUrl)
		check(err)

    _, err = fmt.Fprintf(w, "\n---\n\n")
    check(err)

	_, err = fmt.Fprintf(w, "## Summary\n")
	check(err)
	_, err = fmt.Fprintf(w, "%s\n\n", movie.Summary)
	check(err)
	_, err = fmt.Fprintf(w, "## Main ideas\n\n")
	check(err)
	_, err = fmt.Fprintf(w, "## What attracted attention\n\n")
	check(err)
	_, err = fmt.Fprintf(w, "## Who might be interested\n\n")
	check(err)

	_, err = fmt.Fprintf(w, "## Links\n\n")
	check(err)


    w.Flush()
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
			e.ForEach("td span", func(i int, a *colly.HTMLElement) {
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
		if title == "Год" || title == "Премьера" {
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

func main() {
	var movie Movie

/*   	movie = ScrapeMovie("9 ярдов")
  	movie.Print()

	fmt.Println("---")

  	movie = ScrapeMovie("Крепкий орешек 1988")
  	movie.Print()

	fmt.Println("---")*/

  	movie = ScrapeMovie(os.Args[1])
  	movie.PrintMarkdown()
}
