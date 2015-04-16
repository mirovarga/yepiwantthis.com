package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	amazon "github.com/dominicphillips/amazing"
	"github.com/kennygrant/sanitize"
)

const (
	productsFile = "amazonstore.json"
)

// AmazonStore provides products from amazon.com. It uses the Product Advertising
// API to fetch the products.
//
// The products are stored in the amazonstore.json file. It tries to rebuild
// the file every 4 hours.
type AmazonStore struct {
	Products             map[string]Product  `json:"products"`
	ProductIDs           []string            `json:"productIDs"`
	ProductIDsByCategory map[string][]string `json:"productIDsByCategory"`
}

// NewAmazonStore creates a new AmazonStore.
//
// It reads the products from the amazonstore.json file. If the file is missing
// it tries to create it by fetching products from the Product Advertising API.
func NewAmazonStore() AmazonStore {
	store := AmazonStore{
		Products:             make(map[string]Product),
		ProductIDsByCategory: make(map[string][]string),
	}
	store.loadFromProductsFile()

	return store
}

// GetRandomProduct returns a random product from the store. If the category
// doesn't equal "" it limits the lookup to the category.
func (s AmazonStore) GetRandomProduct(category string) (Product, error) {
	if category == "" {
		id := s.ProductIDs[rand.Intn(len(s.ProductIDs))]
		return s.Products[id], nil
	}

	products := s.ProductIDsByCategory[category]
	if len(products) == 0 {
		return Product{}, fmt.Errorf("No products found [category: %s]", category)
	}
	return s.Products[products[rand.Intn(len(products))]], nil
}

// GetProduct returns the product with the provided ID from the store.
func (s AmazonStore) GetProduct(id string) (Product, error) {
	if product, found := s.Products[id]; found {
		return product, nil
	}

	return Product{}, fmt.Errorf("Product not found [ID: %s]", id)
}

// GetSimilarProducts returns up to 10 products from the same category as
// the product with the provided ID.
func (s AmazonStore) GetSimilarProducts(id string) ([]Product, error) {
	if product, found := s.Products[id]; found {
		var similar []Product

		similarProductIDs := make(map[string]string)
		productIDs := s.ProductIDsByCategory[product.Category]
		for range productIDs {
			productID := productIDs[rand.Intn(len(productIDs))]
			if productID != id && similarProductIDs[productID] == "" {
				similarProductIDs[productID] = productID
				similar = append(similar, s.Products[productID])
				if len(similar) == 10 {
					break
				}
			}
		}

		return similar, nil
	}

	return []Product{}, fmt.Errorf("Product not found [ID: %s]", id)
}

func (s *AmazonStore) loadFromProductsFile() {
	bytes, err := ioutil.ReadFile(productsFile)
	if err != nil {
		log.Printf("ERROR %s", err.Error())
		return
	}

	json.Unmarshal(bytes, s)
}

const (
	locale      = "US"
	associateID = ""
	apiKey      = ""
	apiSecret   = ""
)

func (s AmazonStore) build() error {
	categories := map[string]string{
		"babies":         "Baby",
		"beauty":         "Beauty",
		"books":          "Books",
		"fashion-babies": "FashionBaby",
		"fashion-boys":   "FashionBoys",
		"fashion-girls":  "FashionGirls",
		"fashion-men":    "FashionMen",
		"fashion-women":  "FashionWomen",
		"jewelry":        "Jewelry",
		"pets":           "PetSupplies",
		"toys-and-games": "Toys",
		"watches":        "Watches",
	}

	log.Printf("INFO  Building products started")
	client, _ := amazon.NewAmazing(locale, associateID, apiKey, apiSecret)

	for key, value := range categories {
		for i := 1; i <= 10; i++ {
			err := s.buildCategory(client, key, value, i)
			if err != nil {
				log.Printf("ERROR Building products finished [error: %s]", err.Error())
				return err
			}
		}
	}

	json, _ := json.Marshal(s)
	ioutil.WriteFile(productsFile, json, 0600)

	log.Printf("INFO  Building products finished [products: %d]", len(s.ProductIDs))
	return nil
}

func (s *AmazonStore) buildCategory(client *amazon.Amazing, category,
	amazonCategory string, page int) error {
	products, err := fetchProducts(client, category, amazonCategory, page)
	if err != nil {
		return err
	}

	for _, product := range products {
		s.Products[product.ID] = product
		s.ProductIDs = append(s.ProductIDs, product.ID)
		s.ProductIDsByCategory[category] = append(s.ProductIDsByCategory[category], product.ID)
	}
	return nil
}

func fetchProducts(client *amazon.Amazing, category, amazonCategory string,
	page int) ([]Product, error) {
	sort := "salesrank"
	if strings.Contains(amazonCategory, "Fashion") {
		sort = "reviewrank"
	}

	params := url.Values{
		"Operation":     []string{"ItemSearch"},
		"SearchIndex":   []string{amazonCategory},
		"Keywords":      []string{"%20"},
		"Condition":     []string{"New"},
		"Availability":  []string{"Available"},
		"MerchantId":    []string{"Amazon"},
		"Sort":          []string{sort},
		"ItemPage":      []string{strconv.Itoa(page)},
		"ResponseGroup": []string{"ItemAttributes,EditorialReview,Images"},
	}

	res, err := client.ItemSearch(params)
	if err != nil {
		return []Product{}, err
	}

	var products []Product
	for _, item := range res.AmazonItems.Items {
		product, err := toProduct(item)
		if err != nil {
			continue
		}

		product.Category = category
		products = append(products, product)
	}
	return products, nil
}

func toProduct(item amazon.AmazonItem) (Product, error) {
	if len(item.EditorialReviews) == 0 {
		return Product{}, fmt.Errorf("Invalid product")
	}

	product := Product{
		ID:           item.ASIN,
		Title:        strings.TrimSpace(item.ItemAttributes.Title),
		Description:  strings.TrimSpace(sanitize.HTML(item.EditorialReviews[0].Content)),
		ThumbnailURL: item.MediumImage.URL,
		ImageURL:     item.LargeImage.URL,
		Price:        item.ItemAttributes.ListPrice.FormattedPrice,
		URL:          item.DetailPageURL,
		AddedOn:      time.Now(),
	}

	if product.Title == "" || product.Description == "" ||
		product.ThumbnailURL == "" || product.ImageURL == "" ||
		product.Price == "" {
		return Product{}, fmt.Errorf("Invalid product")
	}
	return product, nil
}
