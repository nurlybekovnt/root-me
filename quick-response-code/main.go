package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

const (
	challengeURL = "http://challenge01.root-me.org/programmation/ch7/"
)

func main() {
	challenge, err := Fetch()
	if err != nil {
		log.Panicf("failed to fetch challenge: %v", err)
	}

	solution, err := Solve(challenge)
	if err != nil {
		log.Panicf("failed to solve challenge: %v", err)
	}

	result, err := Submit(solution)
	if err != nil {
		log.Panicf("failed to submit solution: %v", err)
	}

	fmt.Println(result)
}

type Challenge struct {
	Image   image.Image
	cookies []*http.Cookie
}

type Solution struct {
	Key     string
	cookies []*http.Cookie
}

var (
	regexpBase64Image = regexp.MustCompile(`base64,([^"]+)`)
	regexpKey         = regexp.MustCompile(`\/.*$`)
)

func Fetch() (Challenge, error) {
	resp, err := http.Get(challengeURL)
	if err != nil {
		return Challenge{}, fmt.Errorf("failed to fetch challenge: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Challenge{}, fmt.Errorf("failed to read response body: %v", err)
	}

	matches := regexpBase64Image.FindStringSubmatch(string(body))
	if len(matches) != 2 {
		return Challenge{}, fmt.Errorf("failed to find base64 encoded image")
	}

	decodedData, err := base64.StdEncoding.DecodeString(matches[1])
	if err != nil {
		return Challenge{}, fmt.Errorf("failed to decode base64 encoded image: %v", err)
	}

	img, _, err := image.Decode(bytes.NewBuffer(decodedData))
	if err != nil {
		return Challenge{}, fmt.Errorf("failed to decode image: %v", err)
	}

	return Challenge{
		Image:   img,
		cookies: resp.Cookies(),
	}, nil
}

func Solve(ch Challenge) (Solution, error) {
	img := FixImage(ch.Image)

	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		log.Panicf("failed to create binary bitmap: %v", err)
	}

	qrReader := qrcode.NewQRCodeReader()
	qrCode, err := qrReader.Decode(bmp, nil)
	if err != nil {
		log.Panicf("failed to read QR: %v", err)
	}

	return Solution{
		Key:     regexpKey.FindString(qrCode.GetText()),
		cookies: ch.cookies,
	}, nil
}

func FixImage(img image.Image) image.Image {
	// Create a new image with the same dimensions as the original image
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	// Copy the original image to the new image
	draw.Draw(newImg, bounds, img, bounds.Min, draw.Src)

	const (
		w  = 9 // Each large pixel in the QR code is composed of 9x9 pixels
		w2 = w * 2
		w5 = w * 5
		w6 = w * 6
		w7 = w * 7
	)

	black := color.Black
	white := color.White
	for _, rect := range []image.Point{
		{18, 18},
		{18, 216},
		{216, 18},
	} {
		draw.Draw(newImg,
			image.Rect(rect.X, rect.Y, rect.X+w7, rect.Y+w7),
			&image.Uniform{black},
			image.Point{},
			draw.Src,
		)
		draw.Draw(newImg,
			image.Rect(rect.X+w, rect.Y+w, rect.X+w6, rect.Y+w6),
			&image.Uniform{white},
			image.Point{},
			draw.Src,
		)
		draw.Draw(newImg,
			image.Rect(rect.X+w2, rect.Y+w2, rect.X+w5, rect.Y+w5),
			&image.Uniform{black},
			image.Point{},
			draw.Src,
		)
	}

	return newImg
}

func Submit(s Solution) (string, error) {
	form := url.Values{
		"metu": {s.Key},
	}

	req, err := http.NewRequest(http.MethodPost, challengeURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	for _, cookie := range s.cookies {
		req.AddCookie(cookie)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to submit solution: %v", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	return string(data), nil
}
