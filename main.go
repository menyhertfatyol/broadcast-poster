package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	tuesday := getLastTuesday()
	url := "https://archive.tilos.hu/download/tilos-" + tuesday + "-210000-220000.mp3"

	downloadShow(url, tuesday)

	createSVGLogo()

}

func createSVGLogo() {
	topColorValues := []int{178, 226, 218, 150}
	bottomColorValues := []int{50, 83, 98, 0}

	topColors := pickRandomValuesFromAry(topColorValues, 3)
	bottomColors := pickRandomValuesFromAry(bottomColorValues, 3)

	topColorsString := (strings.Trim(strings.Join(strings.Fields(fmt.Sprint(topColors)), ","), "[]"))
	bottomColorsString := (strings.Trim(strings.Join(strings.Fields(fmt.Sprint(bottomColors)), ","), "[]"))

	svg := `<svg height="640" width="640">
  <defs>
    <linearGradient id="grad1" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" style="stop-color:rgb(` + topColorsString + `);stop-opacity:1" />
      <stop offset="100%" style="stop-color:rgb(` + bottomColorsString + `);stop-opacity:1" />
    </linearGradient>
  </defs>
  <ellipse cx="320" cy="320" rx="300" ry="300" fill="url(#grad1)" />
  <text fill="#ffffff" font-size="80" font-family="Verdana" x="60" y="340">keddestidrog</text>
  Sorry, your browser does not support inline SVG.
</svg>`

	f, err := os.Create("mix_image.svg")
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err = f.WriteString(svg)
	if err != nil {
		log.Fatal(err)
	}
}

func pickRandomValuesFromAry(listOfNumbers []int, numbersToPick int) []int {
	var numbersPicked []int
	rand.Seed(time.Now().Unix())

	for index := 0; index < numbersToPick; index++ {
		numbersPicked = append(numbersPicked, listOfNumbers[rand.Intn(len(listOfNumbers))])
	}
	return numbersPicked
}

func downloadShow(url, dateOfShow string) {

	response, err := http.Get(url)
	check(err)
	defer response.Body.Close()

	file, err := os.Create("keddestidrogmusor-" + dateOfShow + ".mp3")
	check(err)

	_, err = io.Copy(file, response.Body)
	check(err)
	file.Close()
}

func getLastTuesday() string {
	currentDate := time.Now()
	for currentDate.Weekday() != time.Tuesday {
		currentDate = currentDate.AddDate(0, 0, -1)
	}
	return currentDate.Format("20060102")
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
