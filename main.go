package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	yaml "gopkg.in/yaml.v2"
)

const (
	listFileName    = "list.json"
	listURL         = "http://archeagedatabase.net/query.php?a=recipes&l=kr"
	itemQueryFormat = "http://archeagedatabase.net/tip.php?id=recipe--%v&l=kr&nf=on"

	recipeDetailSelector    = `body > div > div > table > tbody > tr:nth-child(3) > td > table > tbody > tr > td:nth-child(2)`
	recipeMeterialsSelector = `body > div > div > table > tbody > tr:nth-child(5) > td > div.reward_counter_big`
	recipeRewardSelector    = `body > div > div > table > tbody > tr:nth-child(7) > td > div.reward_counter_big`
)

var (
	nameRe     = regexp.MustCompile(`<span class="item_title">([^<]*)`)
	laborRe    = regexp.MustCompile(`필요 노동력: ([^<]*)`)
	ptimeRe    = regexp.MustCompile(`Production time: ([^<]*)`)
	quantityRe = regexp.MustCompile(`(.*) x (.*)`)

	errSomethingWrong = errors.New("something wrong")
)

type List struct {
	Data [][]interface{}
}

type Recipes []*Recipe

type Recipe struct {
	Name           string
	Labor          int
	ProductionTime string
	Meterials      Meterials
	Reward         string
	Quantity       int
}

func (r Recipe) String() string {
	return fmt.Sprintf("Name:%v, Labor:%v, Ptime:%v, Reward:%v X %v\n%v", r.Name, r.Labor, r.ProductionTime, r.Reward, r.Quantity, r.Meterials.String())
}

type Meterials []Meterial

func (m Meterials) String() string {
	s := ""
	for _, v := range m {
		s += fmt.Sprintf("%v X %v\n", v.Name, v.Quantity)
	}
	return s
}

type Meterial struct {
	Name     string
	Quantity int
}

func init() {
	if _, err := os.Stat(listFileName); os.IsNotExist(err) {
		resp, err := http.Get(listURL)
		if err != nil {
			panic(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile(listFileName, body[3:], os.ModePerm); err != nil {
			panic(err)
		}
	}
}

func main() {
	data, err := ioutil.ReadFile(listFileName)
	if err != nil {
		panic(err)
	}
	list := List{}
	if err := json.Unmarshal(data, &list); err != nil {
		panic(err)
	}

	recipes := make(Recipes, len(list.Data))
	ch := make(chan *Recipe)
	for i, item := range list.Data {
		url := fmt.Sprintf(itemQueryFormat, item[0])
		go request(i, url, ch)
	}
	for i := 0; i < 9064; i++ {
		recipe := <-ch
		fmt.Printf("%v/%v\n", i, len(list.Data))
		recipes[i] = recipe
	}
	// close(ch)

	marshaled, _ := yaml.Marshal(recipes)
	err = ioutil.WriteFile("dictionary.yaml", marshaled, os.ModePerm)
}

func request(i int, url string, ch chan<- *Recipe) {
	// TODO.
	// need more effective error handler
	for i := 0; i < 10; i++ {
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		doc, err := goquery.NewDocumentFromResponse(resp)
		if err != nil {
			continue
		}
		name, labor, ptime, quantity, reward, err := parseRecipeDetail(doc)
		if err != nil {
			continue
		}
		recipe := &Recipe{
			Name:           name,
			Labor:          labor,
			ProductionTime: ptime,
			Quantity:       quantity,
			Reward:         reward,
			Meterials:      parseRecipeMeterials(doc),
		}
		ch <- recipe
		return
	}
	ch <- &Recipe{}
}

func parseRecipeDetail(doc *goquery.Document) (name string, labor int, ptime string, quantity int, reward string, err error) {
	html, err := doc.Find(recipeDetailSelector).Html()
	if err != nil {
		return
	}

	names := nameRe.FindStringSubmatch(html)
	if len(names) != 2 {
		err = errSomethingWrong
		return
	}
	name = names[1]

	labors := laborRe.FindStringSubmatch(html)
	if len(labors) != 2 {
		err = errSomethingWrong
		return
	}
	labor, _ = strconv.Atoi(labors[1])

	ptimes := ptimeRe.FindStringSubmatch(html)
	if len(ptimes) != 2 {
		err = errSomethingWrong
		return
	}
	ptime = ptimes[1]

	rewardsAndQuantity := quantityRe.FindStringSubmatch(doc.Find(recipeRewardSelector).Text())
	if len(rewardsAndQuantity) != 3 {
		err = errSomethingWrong
		return
	}
	reward = rewardsAndQuantity[1]
	quantity, _ = strconv.Atoi(rewardsAndQuantity[2])
	return
}

func parseRecipeMeterials(doc *goquery.Document) (m []Meterial) {
	doc.Find(recipeMeterialsSelector).Each(func(i int, s *goquery.Selection) {
		matches := quantityRe.FindStringSubmatch(s.Text())
		if len(matches) == 3 {
			quantity, _ := strconv.Atoi(matches[2])
			m = append(m, Meterial{Name: matches[1], Quantity: quantity})
		}
	})
	return
}
