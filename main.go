package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

type StockData struct {
	Symbol        string  `json:"01. symbol"`
	Price         float64 `json:"05. price,string"`
	Change        float64 `json:"09. change,string"`
	ChangePercent string  `json:"10. change percent"`
}

type AlphaVantageResponse struct {
	GlobalQuote StockData `json:"Global Quote"`
}

func loadEnv() error {
	return godotenv.Load()
}

func getStockData(symbol, apiKey string) (*StockData, error) {
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s", symbol, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response AlphaVantageResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	changePercent := strings.TrimSuffix(response.GlobalQuote.ChangePercent, "%")
	var changePercentFloat float64
	fmt.Sscanf(changePercent, "%f", &changePercentFloat)
	response.GlobalQuote.ChangePercent = fmt.Sprintf("%.2f%%", changePercentFloat)

	return &response.GlobalQuote, nil
}

func displayStockData(data *StockData) {
	yellow := color.New(color.FgYellow).SprintFunc()

	var changeText string
	if data.Change < 0 {
		changeText = color.RedString(data.ChangePercent)
	} else {
		changeText = color.GreenString(data.ChangePercent)
	}

	fmt.Printf("%s: $%.2f (%s)\n",
		yellow(data.Symbol),
		data.Price,
		changeText)
}

func main() {
	if err := loadEnv(); err != nil {
		fmt.Println("Error loading .env file:", err)
		os.Exit(1)
	}

	apiKey := os.Getenv("ALPHA_VANTAGE_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: ALPHA_VANTAGE_API_KEY not set in .env file")
		os.Exit(1)
	}

	symbols := strings.Split(os.Getenv("STOCK_SYMBOLS"), ",")
	if len(symbols) == 0 {
		symbols = []string{"AAPL", "MSFT", "GOOGL"}
	}

	updateInterval := 60
	if interval := os.Getenv("UPDATE_INTERVAL"); interval != "" {
		if _, err := fmt.Sscanf(interval, "%d", &updateInterval); err != nil {
			fmt.Println("Error parsing UPDATE_INTERVAL:", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Starting Stock Market Tracker (Update interval: %d seconds)\n", updateInterval)
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println()

	for {
		for _, symbol := range symbols {
			data, err := getStockData(symbol, apiKey)
			if err != nil {
				fmt.Printf("Error fetching data for %s: %v\n", symbol, err)
				continue
			}
			displayStockData(data)
		}
		fmt.Println()
		time.Sleep(time.Duration(updateInterval) * time.Second)
	}
}
