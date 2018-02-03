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

	var jsonBytes []byte

	URLResponse, URLerr := http.Get(url)
	if URLerr != nil {
		log.Fatal(URLerr)
	} else if URLResponse.StatusCode == 200 {
		jsonBytes = getHTMLBodyBytes(URLResponse)
	} else {
		fmt.Println("The following url seems to be broken:", url)
	}

	var episodeJSON episode

	if err := json.Unmarshal(jsonBytes, &episodeJSON); err != nil {
		log.Fatal(err)
	}

	mp3Response, err := http.Get(episodeJSON.mp3Url())
	if err != nil {
		log.Fatal(err)
	} else if mp3Response.StatusCode == 200 {
		downloadShow(mp3Response, lastTuesday.Format("20060102"))
	} else {
		fmt.Println("The following url seems to be broken:", url)
	}

	createSVGLogo()

	exec.Command("inkscape", "mix_image.svg", "--export-png=mix_image.png").Run()

	audioPath, _ := os.Getwd()
	audioPath += "/keddestidrogmusor-" + lastTuesday.Format("20060102") + ".mp3"
	imgPath, _ := os.Getwd()
	imgPath += "/mix_image.png"
	extraParams := map[string]string{
		"name":        lastTuesday.Format("2006.01.02") + " - " + episodeJSON.Text.Title,
		"tags-0-tag":  "Tilos Radio",
		"tags-2-tag":  "FM 90.3",
		"tags-3-tag":  "Drog",
		"description": episodeJSON.Text.Content,
	}

	request, err := newMixcloudUploadRequest("https://api.mixcloud.com//upload/?access_token="+os.Getenv("MIXCLOUD_ACCESS_TOKEN"), extraParams, "mp3", audioPath, "picture", imgPath)
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

type episode struct {
	Text struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	} `json:"text"`
	M3uURL string `json:"m3uUrl"`
}

func (e episode) mp3Url() string {
	return strings.Split(e.M3uURL, ".m3u")[0] + ".mp3"
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

func downloadShow(r *http.Response, dateOfShow string) {

	defer r.Body.Close()

	file, fileerr := os.Create("keddestidrogmusor-" + dateOfShow + ".mp3")
	check(fileerr)

	_, err := io.Copy(file, r.Body)
	check(err)
	file.Close()
}

func getHTMLBodyBytes(r *http.Response) []byte {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	check(err)
	return []byte(body)
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
