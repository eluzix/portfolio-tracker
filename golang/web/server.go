package web

import (
	"embed"
	"html/template"
	"net/http"
	"sync"
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

func StartServer() {
	r := gin.Default()
	funcMap := template.FuncMap{
		"toCurrency": utils.ToCurrencyStringUSD,
		"toYield":    utils.ToYieldString,
	}
	templ := template.Must(template.New("").Funcs(funcMap).ParseFS(f, "templates/*.html"))
	r.SetHTMLTemplate(templ)

	r.GET("/", func(c *gin.Context) {
		db, cleanup := storage.OpenLocalDatabase(false)
		defer cleanup()

		accounts, _ := loaders.UserAccounts(db)
		ac := types.Account{
			Id:   "",
			Name: "All Portfolio",
		}
		(*accounts) = append((*accounts), ac)

		accountsData := make(map[string]types.AnalyzedPortfolio, len(*accounts))
		var wg sync.WaitGroup
		for i := range *accounts {
			wg.Go(func() {
				ac := (*accounts)[i]
				data, _ := portfolio.LoadAndAnalyze(db, ac)
				accountsData[ac.Id] = data
			})
		}
		wg.Wait()

		c.HTML(http.StatusOK, "index.html", gin.H{
			"accounts":     accounts,
			"accountsData": accountsData,
		})
	})

	r.POST("/updateMarket", func(c *gin.Context) {
		db, cleanup := storage.OpenLocalDatabase(false)
		defer cleanup()

		market.UpdateMarketData(db)

		c.String(http.StatusOK, "Market data updated")
	})

	r.Run()
}
