package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/common/log"
)

const imageQuality = 75

var (
	error404              = errors.New("page not found")
	errorInvalidDimension = errors.New("width and height values must be positive")
)

func sendError(w http.ResponseWriter, err error) {
	response, _ := json.Marshal(map[string]string{"error": err.Error()})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)
	w.Write(response)
}

func main() {

	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	log.Infoln("Starting http server on port:", port)

	r := mux.NewRouter()
	r.HandleFunc("/", serveMainPageRequest)
	r.HandleFunc("/{width:[0-9]+}", serveImageRequest).Methods("GET")
	r.HandleFunc("/{width:[0-9]+}.jpg", serveImageRequest).Methods("GET")
	r.HandleFunc("/{width:[0-9]+}/{height:[0-9]+}", serveImageRequest).Methods("GET")
	r.HandleFunc("/{width:[0-9]+}x{height:[0-9]+}", serveImageRequest).Methods("GET")
	r.HandleFunc("/{width:[0-9]+}/{height:[0-9]+}.jpg", serveImageRequest).Methods("GET")
	r.HandleFunc("/{width:[0-9]+}x{height:[0-9]+}.jpg", serveImageRequest).Methods("GET")

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		sendError(w, error404)
	})

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	log.Fatalln(srv.ListenAndServe())
}
func serveMainPageRequest(w http.ResponseWriter, _ *http.Request) {
	var html = []byte(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
<!-- Bootstrap CSS -->
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/css/bootstrap.min.css" integrity="sha384-Gn5384xqQ1aoWXA+058RXPxPg6fy4IWvTNh0E263XmFcJlSAwiGgFAW/dAiS6JXm" crossorigin="anonymous">
<title>Placer!</title>
</head>
<body>
<div class="container">
<div class="row">
<div class="col-sm">
<h1>Placer!</h1>
<p>Provide image dimension (in px) as part of our URL to generate placeholder image.<br />Example:</p>
<pre>
&lt;img src="/WIDTHxHEIGHT.jpg" /&gt;
&lt;img src="/220x64.jpg" /&gt;
</pre>
<p>will generate:</p>
<a href="/220x64.jpg"><img src="/220x64.jpg" /></a>
</div>
</div>
</div>
</body>
</html>`)
	w.Write(html)
}

func getInputValue(requestVars map[string]string, keyName string) (value int, err error) {
	var userValue bool

	_, userValue = requestVars[keyName]
	if userValue {
		value, err = strconv.Atoi(requestVars[keyName])
		if err != nil {
			return 0, err
		}
	}

	return value, nil
}

func generateImage(width, height int) (*image.RGBA, error) {

	if width < 0 || height < 0 {
		return nil, errorInvalidDimension
	}

	if width == 0 && height == 0 {
		return nil, errorInvalidDimension
	}

	if width == 0 {
		width = height
	} else if height == 0 {
		height = width
	}

	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(canvas, canvas.Bounds(), image.NewUniform(color.RGBA{0xCE, 0xCE, 0xCE, 0xFF}), image.ZP, draw.Src)

	return canvas, nil
}

func serveImageRequest(w http.ResponseWriter, r *http.Request) {
	requestVars := mux.Vars(r)

	width, err := getInputValue(requestVars, "width")
	if err != nil {
		sendError(w, err)
		return
	}

	height, err := getInputValue(requestVars, "height")
	if err != nil {
		sendError(w, err)
		return
	}

	img, err := generateImage(width, height)
	if err != nil {
		sendError(w, err)
		return
	}

	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, img, &jpeg.Options{Quality: imageQuality}); err != nil {
		sendError(w, err)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	w.Write(buffer.Bytes())
}
