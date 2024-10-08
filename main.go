package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
)

var (
    locationEndpoint = "https://api.tacobell.com/location/v1/"
    storesEndpoint = "https://www.tacobell.com/tacobellwebservices/v4/tacobell/stores"
    foodEndpoint = "https://www.tacobell.com/tacobellwebservices/v4/tacobell/products/menu/"
)

type FoodItem struct {
    Name string `json:"name"`
    Code string `json:"code"`
    Category string `json:"category"`
    Price float64 `json:"price"`
    Image string `json:"image"`
}

type Store struct {
    StoreNumber string `json:"storeNumber"`
    StoreStatus string `json:"storeStatus"`
    FormattedDistance string `json:"formattedDistance"`
}

type Location struct {
    Geometry Geometry `json:"geometry"`
    Success bool `json:"success"`
}

type Geometry struct {
    Lat float64 `json:"lat"`
    Lng float64 `json:"lng"`
}

func main() {
    client := http.Client{}
    cookiejar, _ := cookiejar.New(nil)
    client.Jar = cookiejar

    resp, err := client.Get(locationEndpoint + "pomona")
    if err != nil {
        fmt.Println(err)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err)
    }

    geo := Location{}
    err = json.Unmarshal(body, &geo)
    if err != nil {
        fmt.Println(err)
    }

    resp, err = client.Get(storesEndpoint + "?latitude=" + fmt.Sprintf("%f", geo.Geometry.Lat) + "&longitude=" + fmt.Sprintf("%f", geo.Geometry.Lng))
    if err != nil {
        fmt.Println(err)
    }


    body, err = io.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err)
    }

    var output map[string]interface{}
    err = json.Unmarshal(body, &output)
    if err != nil {
        fmt.Println(err)
    }

    stores := output["nearByStores"].([]interface{})

    storeResults := make(map[string]Store)
    for _, store := range stores {
        storeMap := store.(map[string]interface{})
        storeNum := storeMap["storeNumber"]
        storeStatus := storeMap["storeStatus"]
        formattedDistance := storeMap["formattedDistance"]

        store := Store{
            StoreNumber: storeNum.(string),
            StoreStatus: storeStatus.(string),
            FormattedDistance: formattedDistance.(string),
        }

        storeResults[storeNum.(string)] = store
    }

    var closestStoreNum string
    minDistance := 99999999999.0
    for key, store := range storeResults {
        if store.StoreStatus == "openNow" {
            distance := strings.Split(store.FormattedDistance, " ")[0]
            distanceVal, err := strconv.ParseFloat(distance, 64)
            if err != nil {
                fmt.Println(err)
            }
            if distanceVal < minDistance {
                minDistance = distanceVal
                closestStoreNum = key
            }
        }
    }

    req, err := http.NewRequest("GET", foodEndpoint + closestStoreNum, nil)
    if err != nil {
        fmt.Println(err)
    }

    resp, err = client.Do(req)
    if err != nil {
        fmt.Println(err)
    }

    body, err = io.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err)
    }

    var foodOutput map[string]interface{}
    err = json.Unmarshal(body, &foodOutput)
    if err != nil {
        fmt.Println(err)
    }

    menu := foodOutput["menuProductCategories"].([]interface{})
    for _, category := range menu {
        category := category.(map[string]interface{})
        fmt.Println(category["name"])
        fmt.Println("-------------------------------------------------")
        products := category["products"].([]interface{})
        for _, product := range products {
            product := product.(map[string]interface{})
            name := product["name"]
            priceVal := product["price"].(map[string]interface{})["value"].(float64)
            image := product["images"].([]interface{})[0].(map[string]interface{})["url"]
            code := product["code"]
            foodItem := FoodItem{
                Name: name.(string),
                Code: code.(string),
                Category: category["name"].(string),
                Price: priceVal,
                Image: image.(string),
            }
            fmt.Println("Name: ", foodItem.Name)
            fmt.Println("Code: ", foodItem.Code)
            fmt.Println("Category: ", foodItem.Category)
            fmt.Println("Price: ", foodItem.Price)
            fmt.Println("Image: ", foodItem.Image)
            fmt.Println()
        }
    }



}
