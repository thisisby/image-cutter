package main

import (
	"crypto/sha1"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nfnt/resize"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

func main() {
	inputDir := "./images"
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	hash := shortHash(timestamp)
	runFolder := fmt.Sprintf("results/%s_%s", timestamp, hash)

	err := os.MkdirAll(runFolder, os.ModePerm)
	if err != nil {
		panic(err)
	}

	err = filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			return nil
		}

		return processImage(path, runFolder)
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("✅ All images processed. Output folder: %s\n", runFolder)
}

// processImage splits and resizes image into 4 quadrants
func processImage(imagePath, outputDir string) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	midX := width / 2
	midY := height / 2

	quadrants := []image.Rectangle{
		image.Rect(0, 0, midX, midY),          // top-left
		image.Rect(midX, 0, width, midY),      // top-right
		image.Rect(0, midY, midX, height),     // bottom-left
		image.Rect(midX, midY, width, height), // bottom-right
	}

	baseName := strings.TrimSuffix(filepath.Base(imagePath), filepath.Ext(imagePath))

	for i, rect := range quadrants {
		subImg := img.(interface {
			SubImage(r image.Rectangle) image.Image
		}).SubImage(rect)

		// Resize to 120x120
		resized := resize.Resize(120, 120, subImg, resize.Lanczos3)

		outFileName := fmt.Sprintf("%s_part%d.png", baseName, i+1)
		outPath := filepath.Join(outputDir, outFileName)

		outFile, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		if format == "jpeg" || format == "jpg" {
			err = jpeg.Encode(outFile, resized, nil)
		} else {
			err = png.Encode(outFile, resized)
		}
		if err != nil {
			return err
		}
	}

	fmt.Printf("✅ Processed: %s\n", baseName)
	return nil
}

// shortHash returns first 6 characters of a SHA-1 hash
func shortHash(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))[:6]
}
