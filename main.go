package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	tuesday := getLastTuesday()
	url := "https://archive.tilos.hu/download/tilos-" + tuesday + "-210000-220000.mp3"

	downloadShow(url, tuesday)

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
