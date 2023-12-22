package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gocolly/colly/v2"
)

// Parity struct represents a currency exchange parity.
type Parity struct {
	From     string
	To       string
	Value    float64
	DateTime time.Time
}

// NewParity creates a new Parity instance with the given currency pair.
func NewParity(from, to string) Parity {
	return Parity{From: from, To: to}
}

// Info returns a formatted string representation of the parity.
func (p Parity) Info() string {
	return p.From + "/" + p.To + "," + fmt.Sprint(p.Value) + "," + p.DateTime.Format("2006-01-02 15:04:05")
}

// Get fetches the currency parity from Wise.
func (p *Parity) Get() {
	url := fmt.Sprintf("https://wise.com/us/currency-converter/%s-to-%s-rate", p.From, p.To)

	// Create a new collector instance
	c := colly.NewCollector()
	c.SetRequestTimeout(120 * time.Second)

	// Define the HTML element and its handler to extract the parity value
	c.OnHTML("span.text-success", func(e *colly.HTMLElement) {
		p.Value, _ = strconv.ParseFloat(e.Text, 64)
		p.DateTime = time.Now()
	})

	// Log events during the scraping process
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting page...", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Got a response from:", r.Request.URL)
	})

	c.OnError(func(r *colly.Response, e error) {
		fmt.Println("Got this error:", e, "from", r.Request.URL)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished!", r.Request.URL)
	})

	// Visit the URL to initiate the scraping process
	c.Visit(url)
}

// WriteFile appends parity information to a CSV file.
func WriteFile(filePath string, parities []Parity) {
	dataFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dataFile.Close()

	// Iterate through parities and write to the file
	for _, p := range parities {
		p.Get()
		_, err := fmt.Fprintln(dataFile, p.Info())
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

// CreateFile creates a new CSV file if it doesn't exist.
func CreateFile(filePath string) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		newFile, err := os.Create(filePath)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer newFile.Close()
		_, err = fmt.Fprintln(newFile, "parities,parity_value,parity_date_time")
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func main() {
	// Initial currency pairs to track
	parities := []Parity{
		NewParity("USD", "TRY"),
		NewParity("EUR", "TRY"),
		NewParity("GBP", "TRY"),
		NewParity("CHF", "TRY"),
		NewParity("JPY", "TRY"),
	}

	// File path to store parity information
	filePath := "exchange.csv"

	// Update interval for fetching and updating currency rates
	interval := time.Minute
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Create the CSV file if it doesn't exist
	CreateFile(filePath)

	// Write initial parity information to the file
	WriteFile(filePath, parities)

	// Periodically update and write parity information
	for range ticker.C {
		WriteFile(filePath, parities)
	}
}
