package main

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"log"
	"net/http"
)

const fileId = "1bzZqgCtkbLaWchoro5dnN8Jz7gPceTmGelR0ov97Tjc"

func main() {
	apiHtmlResponse, err := http.Get("https://confluence.hflabs.ru/pages/viewpage.action?pageId=1181220999")
	if err != nil {
		fmt.Println("Не удалось получить страницу")
	}
	content, err := goquery.NewDocumentFromReader(apiHtmlResponse.Body)
	if err != nil {
		fmt.Println("Не удалось cоздать документ из ответа")
	}
	apiCodes := content.Find("tr")

	// Записываем в слайс данные из таблицы, которую парсим
	parsedValues := make([][]interface{}, 0, 20)

	for i := range apiCodes.Nodes {
		apiCode := apiCodes.Eq(i)
		code := apiCode.Find("td").Eq(0).Text()
		description := apiCode.Find("td").Eq(1).Text()
		if code == "" || description == "" {
			continue
		}
		parsedValues = append(parsedValues, []interface{}{code, description})
	}

	client := getClient()
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))

	if err != nil {
		log.Fatalf("Не удалось подключиться к API: %v", err)
	}
	clearRequest := &sheets.ClearValuesRequest{}
	_, err = srv.Spreadsheets.Values.Clear(fileId, "1", clearRequest).Do()
	if err != nil {
		log.Fatalf("Не удалось очистить sheet: %v", err)
	}
	updatedRows := &sheets.ValueRange{
		Values: parsedValues,
	}

	appendResponse, err := srv.Spreadsheets.Values.Append(fileId, "1", updatedRows).
		ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Context(ctx).Do()

	if err != nil || appendResponse.HTTPStatusCode != 200 {
		log.Fatalf("%e", err)
		return
	}
}
