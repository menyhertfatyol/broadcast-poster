package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {

	lastTuesday := getLastTuesday()
	url := "https://archive.tilos.hu/cache/tilos-" + lastTuesday.Format("20060102") + "-210000-223000.mp3"

	downloadShow(url, lastTuesday.Format("20060102"))

	createSVGLogo()

	exec.Command("inkscape", "mix_image.svg", "--export-png=mix_image.png").Run()

	audioPath, _ := os.Getwd()
	audioPath += "/keddestidrogmusor-" + lastTuesday.Format("20060102") + ".mp3"
	imgPath, _ := os.Getwd()
	imgPath += "/mix_image.png"
	extraParams := map[string]string{
		"name":        "Keddestidrogműsor - " + lastTuesday.Format("2006.01.02") + "-i Adás",
		"tags-0-tag":  "Tilos Radio",
		"tags-2-tag":  "FM 90.3",
		"tags-3-tag":  "Drog",
		"description": "Drogpolitikai magazinműsor kedd esténként",
	}

	request, err := newMixcloudUploadRequest("https://api.mixcloud.com//upload/?access_token="+os.Getenv("ACCESS_TOKEN"), extraParams, "mp3", audioPath, "picture", imgPath)
	check(err)
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	} else {
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()
		fmt.Println(resp.StatusCode)
		fmt.Println(resp.Header)
		fmt.Println(body)
	}

	defer cleanupFiles([]string{audioPath, imgPath, "mix_image.svg"})

}

func cleanupFiles(fileList []string) {
	for _, file := range fileList {
		os.Remove(file)
	}
}

func newMixcloudUploadRequest(uri string, params map[string]string, mp3param, mp3Path, imgParam, imgPath string) (*http.Request, error) {
	file, err := os.Open(mp3Path)
	check(err)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	mp3Part, err := writer.CreateFormFile(mp3param, filepath.Base(mp3Path))
	check(err)

	_, err = io.Copy(mp3Part, file)

	imgFile, err := os.Open(imgPath)
	check(err)
	defer imgFile.Close()

	imgPart, err := writer.CreateFormFile(imgParam, imgPath)
	check(err)

	_, err = io.Copy(imgPart, imgFile)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
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
	check(err)

	defer f.Close()

	_, err = f.WriteString(svg)
	check(err)
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

func getLastTuesday() time.Time {
	currentDate := time.Now()
	for currentDate.Weekday() != time.Tuesday {
		currentDate = currentDate.AddDate(0, 0, -1)
	}
	return currentDate
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
