package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gosimple/slug"
)

var store = NewAmazonStore()

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/robots.txt", func(res http.ResponseWriter, req *http.Request) {
		log.Printf("INFO  robots.txt requested [user agent: %s]", req.UserAgent())
		res.WriteHeader(404)
	})

	router.HandleFunc("/favicon.ico", func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res, req, "./public/images/favicon.ico")
	})

	router.PathPrefix("/public/").Handler(http.StripPrefix("/public/",
		http.FileServer(http.Dir("public/"))))

	router.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		product, err := store.GetRandomProduct("")

		if err != nil {
			log.Printf("FATAL %s", err.Error())
			res.WriteHeader(500)
		}
		render(product, res, req)
	})

	router.HandleFunc("/{category}", func(res http.ResponseWriter, req *http.Request) {
		category := mux.Vars(req)["category"]
		product, err := store.GetRandomProduct(category)

		if err != nil {
			log.Printf("ERROR %s", err.Error())
			router.NotFoundHandler.ServeHTTP(res, req)
		}
		render(product, res, req)
	})

	router.HandleFunc("/{id}/{slug}", func(res http.ResponseWriter, req *http.Request) {
		id := mux.Vars(req)["id"]
		product, err := store.GetProduct(id)

		if err != nil {
			log.Printf("ERROR %s", err.Error())
			router.NotFoundHandler.ServeHTTP(res, req)
		}
		render(product, res, req)
	})

	router.NotFoundHandler = http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		http.Redirect(res, req, "/", 301)
	})

	http.ListenAndServe(":3000", router)
}

var tmplFuncs = template.FuncMap{
	"Truncate": func(toLen int, str string) string {
		if len(str) > toLen {
			return fmt.Sprintf("%s ...", str[:toLen])
		}
		return str
	},
	"LocalURL": func(product Product) string {
		return "/" + product.ID + "/" + slug.Make(product.Title)
	},
	"Subslice": func(from, to int, products []Product) []Product {
		if from >= 0 && from < len(products) && to <= len(products) {
			return products[from:to]
		}
		return []Product{}
	},
}
var tmplSource, _ = ioutil.ReadFile("public/index.html")
var tmpl, _ = template.New("").Funcs(tmplFuncs).Parse(string(tmplSource))

type tmplData struct {
	Product Product
	Similar []Product
}

func render(product Product, res http.ResponseWriter, req *http.Request) {
	similar, _ := store.GetSimilarProducts(product.ID)

	res.Header().Set("Content-Type", "text/html")
	tmpl.Execute(res, tmplData{product, similar})
}
