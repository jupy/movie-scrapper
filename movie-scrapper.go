package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type Movie struct {
	Type          string
	FileName      string
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

var Translations = map[string]string{
	"Канада":              "Canada",
	"СССР":                "USSR",
	"США":                 "USA",
	"детектив":            "detective",
	"драма":               "drama",
	"комедия":             "comedy",
	"мелодрама":           "melodrama",
	"мультфильм":          "cartoon",
	"мюзикл":              "musical",
	"научная фантастика":  "science fiction",
	"приключение":         "adventures",
	"приключения":         "adventures",
	"семейный":            "family",
	"сказка":              "fairy tale",
	"стимпанк":            "steampunk",
	"фэнтези":             "fantasy",
	"экранизация":         "film adaptation",
	"юридический триллер": "legal thriller",
}

func PrintList(w *bufio.Writer, title string, lst []string) {
	if len(lst) == 0 {
		return
	}

	_, err := fmt.Fprintf(w, title)
	check(err)

	text := ""
	for _, str := range lst {
		if len(text) > 0 {
			text += ", "
		}
		text += "[[" + str + "]]"
	}

	_, err = fmt.Fprintf(w, " %s", text)
	check(err)
	_, err = fmt.Fprintf(w, "\n")
	check(err)
}

func (movie *Movie) PrintMarkdown() {
	f, err := os.Create(movie.FileName)
	check(err)
	defer f.Close()

	w := bufio.NewWriter(f)

	_, err = fmt.Fprintf(w, "---\n")
	check(err)
	_, err = fmt.Fprintf(w, "created: %s\n", time.Now().Format("2006-01-02 15:04"))
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
	_, err = fmt.Fprintf(w, "**type:** #%s\n", movie.Type)
	check(err)
	_, err = fmt.Fprintf(w, "**status:** #inbox\n")
	check(err)
	_, err = fmt.Fprintf(w, "**rate:**\n")
	check(err)

	PrintList(w, "**director:**", movie.Directors)
	PrintList(w, "**producer:**", movie.Producers)
	PrintList(w, "**screenwriter:**", movie.Screenwriters)
	PrintList(w, "**company:**", movie.Companies)
	PrintList(w, "**country:**", movie.Countries)
	PrintList(w, "**tags:**", movie.Genres)

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
	_, err = fmt.Fprintf(w, "## Review\n\n")
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

	movie.Type = "movie"
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
		if movie.Name == "" {
			movie.Name = e.Text
		}
	})

	c.OnHTML(".infobox tbody tr:nth-child(2)", func(e *colly.HTMLElement) {
		s := e.Text
		s = strings.TrimPrefix(s, "англ.")
		s = strings.TrimLeftFunc(s, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})
		if strings.HasPrefix(s, "A") && unicode.IsSpace([]rune(s)[1]) {
			s = s[1:]
			s = strings.TrimLeftFunc(s, func(r rune) bool {
				return !unicode.IsLetter(r) && !unicode.IsNumber(r)
			})
			s = s + ", A"
		}
		if strings.HasPrefix(s, "The") && unicode.IsSpace([]rune(s)[3]) {
			s = s[3:]
			s = strings.TrimLeftFunc(s, func(r rune) bool {
				return !unicode.IsLetter(r) && !unicode.IsNumber(r)
			})
			s = s + ", The"
		}
		if movie.InitName == "" {
			movie.InitName = strings.TrimSpace(s)
		}
	})

	c.OnHTML(".infobox-image a img[srcset]", func(e *colly.HTMLElement) {
		link := e.Attr("srcset")
		decodedLink, _ := url.QueryUnescape(link)
		v := strings.Split(decodedLink, " ")
		if len(movie.PosterUrl) == 0 {
			movie.PosterUrl = "https:" + v[0]
		}
	})

	c.OnHTML(".infobox tbody tr", func(e *colly.HTMLElement) {
		title := e.ChildText("th")
		if title == "Жанр" {
			e.ForEach("td a[href]", func(i int, a *colly.HTMLElement) {
				if strings.HasPrefix(a.Text, "[") {
					return
				}
				if a.Text == "экранизация" {
					return
				}
				trans := Translations[a.Text]
				if len(trans) == 0 {
					fmt.Printf("can't translate: %s\n", a.Text)
					movie.Genres = append(movie.Genres, a.Text)
				} else {
					movie.Genres = append(movie.Genres, trans+"|"+a.Text)
				}
			})
		}
		if title == "Сезонов" {
			movie.Type = "serial"
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
			e.ForEach("td span", func(i int, a *colly.HTMLElement) {
				html, _ := a.DOM.Html()
				movie.Companies = ParseList(html)
			})
		}
		if title == "Страна" {
			e.ForEach("td a", func(i int, a *colly.HTMLElement) {
				if a.Text != "" {
					trans := Translations[a.Text]
					if len(trans) == 0 {
						fmt.Printf("can't translate: %s\n", a.Text)
						movie.Countries = append(movie.Countries, a.Text)
					} else {
						movie.Countries = append(movie.Countries, trans+"|"+a.Text)
					}
				}
			})
		}
		if title == "Год" || title == "Премьера" {
			r, _ := regexp.Compile("[0-9][0-9][0-9][0-9]")
			e.ForEachWithBreak("td", func(i int, a *colly.HTMLElement) bool {
				s := r.FindString(a.Text)
				if s != "" {
					movie.Year = s
					return false
				}
				return true
			})
		}
		if title == "На экранах" {
			r, _ := regexp.Compile("[0-9][0-9][0-9][0-9]")
			e.ForEachWithBreak("td", func(i int, a *colly.HTMLElement) bool {
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

func VisitKinoMail(link string) (string, string) {

	var summary string
	var picture string

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

	c.OnHTML("div.p-movie-info img.p-picture__image[src]", func(e *colly.HTMLElement) {
		picture = e.Attr("src")
	})

	c.Visit(link)
	return summary, picture
}

func SearchGoogle(query string, site string) string {
	/*s := "-w " + site*/
	out, err := exec.Command("googler", "-n", "1", "--np", "-w", site, "--json", query).Output()
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
	var pic string

	w, _ := url.QueryUnescape(wikipedia)
	movie = VisitWikipedia(w)
	movie.WikipediaUrl = w
	movie.KinopoiskUrl = kinopoisk
	movie.MailUrl = mail
	movie.Summary, pic = VisitKinoMail(mail)
	if len(movie.PosterUrl) == 0 {
		movie.PosterUrl = pic
	}

	name := strings.TrimSpace(movie.InitName)
	if name == "" {
		name = movie.Name
	}
	movie.FileName = name + " (" + movie.Year + ").md"
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

	reader := bufio.NewReader(os.Stdin)
	movie = ScrapeMovie(os.Args[1])
	movie.Print()
	fmt.Printf("=======\n")
	fmt.Printf("Save markdown file \"" + movie.FileName + "\"? (yes/no)> [yes]")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSuffix(text, "\n")
	if text == "" || text == "yes" {
		movie.PrintMarkdown()
		fmt.Println("file \"" + movie.FileName + "\" created")
	}
}
