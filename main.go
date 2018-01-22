package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

	url := "https://tilos.hu/api/v1/episode/keddidrog/" + lastTuesday.Format("2006") + "/" + lastTuesday.Format("01") + "/" + lastTuesday.Format("02")

	var htmlBody string

	if isValidURL(url) {
		htmlBody = getURLBodyString(url)
	} else {
		log.Fatal("Invalid episode url")
	}

	jsonBytes := []byte(htmlBody)

	var episodeJSON interface{}

	if err := json.Unmarshal(jsonBytes, &episodeJSON); err != nil {
		log.Fatal(err)
	}

	name := episodeJSON.(map[string]interface{})["text"].(map[string]string)["title"]
	description := episodeJSON.(map[string]interface{})["text"].(map[string]string)["content"]
	mp3 := strings.TrimRight(episodeJSON.(map[string]string)["m3uUrl"], ".m3u") + ".mp3"

	if isValidURL(mp3) {
		downloadShow(mp3, lastTuesday.Format("20060102"))
	} else {
		log.Fatal("Invalid mp3 url")
	}

	createSVGLogo()

	exec.Command("inkscape", "mix_image.svg", "--export-png=mix_image.png").Run()

	audioPath, _ := os.Getwd()
	audioPath += "/keddestidrogmusor-" + lastTuesday.Format("20060102") + ".mp3"
	imgPath, _ := os.Getwd()
	imgPath += "/mix_image.png"
	extraParams := map[string]string{
		"name":        lastTuesday.Format("2006.01.02") + " - " + name,
		"tags-0-tag":  "Tilos Radio",
		"tags-2-tag":  "FM 90.3",
		"tags-3-tag":  "Drog",
		"description": description,
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

func newMixcloudUploadRequest(uri string, params map[string]string, mp3Param, mp3Path, imgParam, imgPath string) (*http.Request, error) {
	file, err := os.Open(mp3Path)
	check(err)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	mp3Part, err := writer.CreateFormFile(mp3Param, filepath.Base(mp3Path))
	check(err)

	_, err = io.Copy(mp3Part, file)
	check(err)

	imgFile, err := os.Open(imgPath)
	check(err)
	defer imgFile.Close()

	imgPart, err := writer.CreateFormFile(imgParam, imgPath)
	check(err)

	_, err = io.Copy(imgPart, imgFile)
	check(err)

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

func getURLBodyString(url string) string {
	resp, err := http.Get(url)
	check(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	check(err)
	return string(body)
}

func isValidURL(url string) bool {
	_, err := http.Get(url)
	return err == nil
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
