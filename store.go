package main

import "time"

// Product represents a product in a store.
type Product struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ThumbnailURL string `json:"thumbnailURL"`
	ImageURL     string `json:"imageURL"`
	Price        string `json:"price"`
	URL          string `json:"url"`

	// One of: babies, beauty, books, fashion-babies, fashion-girls,
	// fashion-boys, fashion-women, fashion-men, jewelry, pets, toys-and-games,
	// watches
	Category string    `json:"category"`
	AddedOn  time.Time `json:"addedOn"`
}

// Store holds products that users can view and buy.
type Store interface {
	GetRandomProduct(category string) (Product, error)
	GetProduct(id string) (Product, error)
	GetSimilarProducts(id string) ([]Product, error)
}
