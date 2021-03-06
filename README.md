# yepiwantthis.com

An Amazon affiliate website written in Go using
the [Amazon product advertising API](https://affiliate-program.amazon.com/gp/advertising/api/detail/main.html).
It was available at *yepiwantthis.com* for a short time.

The website got rejected by Amazon so I decided to host the code on GitHub to
help others use the API.


## What it does

Just run it and see for yourself :)

```
$ go get github.com/dominicphillips/amazing \
  github.com/kennygrant/sanitize \
  github.com/gorilla/mux \
  github.com/gosimple/slug
```
```
$ go run website.go amazonstore.go store.go
```

Now open [localhost:3000](http://localhost:3000) in your browser.


## How it works

The products are stored in the `amazonstore.json` file and loaded into memory on
start. The file was generated by running the `amazonstore_test.go` test (it
doesn't work anymore because I removed the API credentials). The file was last
generated on March 27, 2015.

The website uses just one [html/template](http://golang.org/pkg/html/template/)
template.


> There's also a
[blog post](http://www.mirovarga.com/my-amazon-affiliate-experiment-a-sequel)
on my blog.
