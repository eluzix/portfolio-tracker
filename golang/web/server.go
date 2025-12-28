package web

import (
	"embed"
	"html/template"
	"net/http"
	"tracker/loaders"
	"tracker/market"
	"tracker/portfolio"
	"tracker/storage"
	"tracker/types"
	"tracker/utils"

	"github.com/gin-gonic/gin"
)

//go:embed templates/*
var f embed.FS

func toCurrencyWithRate(val int64, precision int, symbol string, rate float64) string {
	return utils.ToCurrencyString(val, precision, symbol, rate)
}

func collectUniqueTags(accounts *[]types.Account) []string {
	tagSet := make(map[string]bool)
	for _, ac := range *accounts {
		for _, tag := range ac.Tags {
			tagSet[tag] = true
		}
	}
	tags := []string{"All"}
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	return tags
}

func hasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func getFilteredAccountIds(accounts *[]types.Account, tagFilter string) []string {
	var ids []string
	for _, ac := range *accounts {
		if ac.Id == "" {
			continue
		}
		if tagFilter == "All" || hasTag(ac.Tags, tagFilter) {
			ids = append(ids, ac.Id)
		}
	}
	return ids
}

func StartServer() {
	r := gin.Default()
	funcMap := template.FuncMap{
		"toCurrency":         utils.ToCurrencyStringUSD,
		"toCurrencyWithRate": toCurrencyWithRate,
		"toYield":            utils.ToYieldString,
	}
	templ := template.Must(template.New("").Funcs(funcMap).ParseFS(f, "templates/*.html"))
	r.SetHTMLTemplate(templ)

	db, cleanup := storage.OpenDatabase()
	defer cleanup()

	r.GET("/", func(c *gin.Context) {
		// db, cleanup := storage.OpenLocalDatabase(false)

		currency := c.DefaultQuery("currency", "USD")
		currencySymbol := market.CurrencySymbolUSD
		exchangeRate := 1.0
		if currency == "ILS" {
			currencySymbol = market.CurrencySymbolILS
			exchangeRate = loaders.CurrencyExchangeRate(db, "ILS")
		}

		tagFilter := c.DefaultQuery("tag", "All")

		accounts, _ := loaders.UserAccounts(db)

		tags := collectUniqueTags(accounts)

		accountsData := make(map[string]types.AnalyzedPortfolio, len(*accounts))
		for _, ac := range *accounts {
			data, _ := portfolio.LoadAndAnalyze(db, ac)
			accountsData[ac.Id] = data
		}

		var filteredAccounts []types.Account
		for _, ac := range *accounts {
			if tagFilter == "All" || hasTag(ac.Tags, tagFilter) {
				filteredAccounts = append(filteredAccounts, ac)
			}
		}

		filteredIds := getFilteredAccountIds(accounts, tagFilter)
		allPortfolioData, _ := portfolio.LoadAndAnalyzeAccounts(db, filteredIds)

		c.HTML(http.StatusOK, "index.html", gin.H{
			"accounts":         &filteredAccounts,
			"accountsData":     accountsData,
			"allPortfolioData": allPortfolioData,
			"currency":         currency,
			"currencySymbol":   currencySymbol,
			"exchangeRate":     exchangeRate,
			"tags":             tags,
			"tagFilter":        tagFilter,
		})
	})

	r.POST("/updateMarket", func(c *gin.Context) {
		// db, cleanup := storage.OpenLocalDatabase(false)
		// db, cleanup := storage.OpenDatabase()
		// defer cleanup()

		market.UpdateMarketData(db)

		c.String(http.StatusOK, "Market data updated")
	})

	r.Run()
}
