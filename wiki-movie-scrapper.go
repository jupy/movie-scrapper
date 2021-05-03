package main

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"

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
						continue;
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
				/*movie.Director = a.Text*/
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
				/*movie.Screenwriters = append(movie.Screenwriters, a.Text)*/
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

var GoogleDomains = map[string]string{
	"us":  "https://www.google.com/search?q=",
	"ru":  "https://www.google.ru/search?q=",
}

func Query(searchTerm string, countryCode string, languageCode string, limit int, start int) string {
	searchTerm = strings.Trim(searchTerm, " ")
	searchTerm = strings.Replace(searchTerm, " ", "+", -1)
	countryCode = strings.ToLower(countryCode)

	var url string

	if googleBase, found := GoogleDomains[countryCode]; found {
		if start == 0 {
			url = fmt.Sprintf("%s%s&hl=%s", googleBase, searchTerm, languageCode)
		} else {
			url = fmt.Sprintf("%s%s&hl=%s&start=%d", googleBase, searchTerm, languageCode, start)
		}
	} else {
		if start == 0 {
			url = fmt.Sprintf("%s%s&hl=%s", GoogleDomains["us"], searchTerm, languageCode)
		} else {
			url = fmt.Sprintf("%s%s&hl=%s&start=%d", GoogleDomains["us"], searchTerm, languageCode, start)
		}
	}

	if limit != 0 {
		url = fmt.Sprintf("%s&num=%d", url, limit)
	}

	return url
}

func SearchGoogleInner(ctx context.Context, searchTerm string, opts ...googlesearch.SearchOptions) ([]googlesearch.Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := googlesearch.RateLimit.Wait(ctx); err != nil {
		return nil, err
	}

	//appe := appengine.NewContext(r)
	c := colly.NewCollector(colly.MaxDepth(1))
	//c.Appengine(appe)
	c.Limit(&colly.LimitRule{
		// Set a delay between requests to these domains
		Delay: 1 * time.Second,
		// Add an additional random delay
		RandomDelay: 1 * time.Second,
	})

	if len(opts) == 0 {
		opts = append(opts, googlesearch.SearchOptions{})
	}

	if opts[0].UserAgent == "" {
		c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36"
	} else {
		c.UserAgent = opts[0].UserAgent
	}

	var lc string
	if opts[0].LanguageCode == "" {
		lc = "en"
	} else {
		lc = opts[0].LanguageCode
	}

	results := []googlesearch.Result{}
	var rErr error
	rank := 1

	c.OnRequest(func(r *colly.Request) {
		if err := ctx.Err(); err != nil {
			r.Abort()
			rErr = err
			return
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		rErr = err
	})

	// https://www.w3schools.com/cssref/css_selectors.asp
	c.OnHTML("div.g", func(e *colly.HTMLElement) {

		sel := e.DOM

		linkHref, _ := sel.Find("a").Attr("href")
		linkText := strings.TrimSpace(linkHref)
		titleText := strings.TrimSpace(sel.Find("div > div > a > h3").Text())

		descText := strings.TrimSpace(sel.Find("div > div > div > span > span").Text())

		if linkText != "" && linkText != "#" {
			result := googlesearch.Result{
				Rank:        rank,
				URL:         linkText,
				Title:       titleText,
				Description: descText,
			}
			results = append(results, result)
			rank += 1
		}
	})

	limit := opts[0].Limit
	if opts[0].OverLimit {
		limit = int(float64(opts[0].Limit) * 1.5)
	}

	query := Query(searchTerm, opts[0].CountryCode, lc, limit, opts[0].Start)
	c.Visit(query)

	if rErr != nil {
		if strings.Contains(rErr.Error(), "Too Many Requests") {
			return nil, googlesearch.ErrBlocked
		}
		return nil, rErr
	}

	// Reduce results to max limit
	if opts[0].Limit != 0 && len(results) > opts[0].Limit {
		return results[:opts[0].Limit], nil
	}

	return results, nil
}

func SearchGoogle(query string, site string) string {

	ctx := context.Background()
	q := query + " site:" + site
	opts := googlesearch.SearchOptions{
		Limit:       1,
		OverLimit:   false,
		CountryCode: "ru",
		UserAgent:   "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.105 Safari/537.36",
	}

	fmt.Printf("%v\n", Query(q, opts.CountryCode, "en", 1, opts.Start))

	googlesearch.RateLimit = rate.NewLimiter(1, 3)
	/*links, err := googlesearch.Search(ctx, q, opts)*/
	links, err := SearchGoogleInner(ctx, q, opts)
	if err != nil {
		fmt.Printf("something went wrong: %v\n", err)
		return ""
	}

	if len(links) < 1 {
		fmt.Printf("there is no any search reults: %v, %v\n", links, err)
		return ""
	}

	link, _ := url.QueryUnescape(links[0].URL)
	return link
}

func ScrapeMovieInner(wikipedia string, kinopoisk string, mail string) Movie {
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
	time.Sleep(2 * time.Second)
	kinopoisk := SearchGoogle(query, "kinopoisk.ru")
	time.Sleep(2 * time.Second)
	mail := SearchGoogle(query, "kino.mail.ru")
	time.Sleep(2 * time.Second)

	return ScrapeMovieInner(wikipedia, kinopoisk, mail)
}

type SearchWikiResult struct {
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

	//c.OnHTML("li.mw-search-result > div > a[href]", func(e *colly.HTMLElement) {
	//c.OnHTML("div.mw-search-result-heading", func(e *colly.HTMLElement) {
	c.OnHTML(".jump-to-nav", func(e *colly.HTMLElement) {
		fmt.Printf("!!! \n")
/*		var r SearchWikiResult
		r.Title = e.Text
		r.Url = e.Attr("href")
		res = append(res, r)*/
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
}

func main() {
	var movie Movie
	var w string

	w = "https://ru.wikipedia.org/wiki/%D0%91%D0%B5%D0%BB%D0%BE%D1%81%D0%BD%D0%B5%D0%B6%D0%BA%D0%B0_%D0%B8_%D1%81%D0%B5%D0%BC%D1%8C_%D0%B3%D0%BD%D0%BE%D0%BC%D0%BE%D0%B2_(%D0%BC%D1%83%D0%BB%D1%8C%D1%82%D1%84%D0%B8%D0%BB%D1%8C%D0%BC)"
	movie = ScrapeMovieInner(w, "", "")
	movie.Print()

/* 	fmt.Println("---")
	w = "https://ru.wikipedia.org/wiki/%D0%9A%D1%80%D0%B5%D0%BF%D0%BA%D0%B8%D0%B9_%D0%BE%D1%80%D0%B5%D1%88%D0%B5%D0%BA_(%D1%84%D0%B8%D0%BB%D1%8C%D0%BC,_1988)"
	movie = ScrapeMovieInner(w, "", "")
	movie.Print()

 	fmt.Println("---")
	w = "https://ru.wikipedia.org/wiki/%D0%94%D0%B5%D0%B2%D1%8F%D1%82%D1%8C_%D1%8F%D1%80%D0%B4%D0%BE%D0%B2"
	movie = ScrapeMovieInner(w, "", "")
	movie.Print()*/

/*   	movie := ScrapeMovie("9 ярдов")
   	movie.Print()

 	fmt.Println("---")

   	movie = ScrapeMovie("Крепкий орешек")
   	movie.Print()*/

/*   	movie := ScrapeMovie("Белоснежка (1937)")
   	movie.Print()*/
//	SearchWiki("Белоснежка")
/*	movie = ScrapeMovieInner(w, "", "")
	movie.Print()*/
}
