package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"log"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func main() {
	exts := flag.String("exts", ".png,.jpeg,.jpg", "Image extensions of files to optimize")
	concurrency := flag.Int("concurrency", 5, "Number of files to optimize concurrently")
	outputDir := flag.String("out", "output", "Output directory for optimized files")
	width := flag.Float64("width", 1500, "Desired width of the optimized image, height will be scaled accordingly to maintain aspect ratio, default is 1500 pixels")

	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		fmt.Println("Please provide a directory to optimize images from")
		os.Exit(1)
	}
	dir := args[0]

	extsList := strings.Split(*exts, ",")

	files, err := recursivelyTraverse(dir, extsList)

	if err != nil {
		log.Fatalf("Error while traversing directory: %v", err)
	}

	err = optimize(*files, *concurrency, dir, *outputDir, float32(*width))
	if err != nil {
		log.Fatalf("Error while optimizing files: %v", err)
	}
}

func recursivelyTraverse(dir string, exts []string) (*[]string, error) {
	file, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !file.IsDir() {
		return &[]string{dir}, nil
	}

	var files []string

	filepath.WalkDir(dir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("Error while traversing directory: %v", err)
			return err
		}
		for _, ext := range exts {
			if strings.HasSuffix(strings.ToLower(path), ext) {
				files = append(files, path)
			}
		}
		return nil
	})
	return &files, nil
}

func optimize(files []string, concurrency int, inputDir, outputDir string, width float32) error {
	sem := make(chan struct{}, concurrency)
	var wg = sync.WaitGroup{}
	for _, file := range files {
		fmt.Println(file)
		wg.Add(1)
		go func(file string, sem *chan struct{}) {
			fmt.Println("Optimizing file: ", file)
			*sem <- struct{}{}
			err := optimizeFile(file, inputDir, outputDir, width)
			if err != nil {
				log.Printf("Error while optimizing file %s: %v", file, err)
			}
			wg.Done()
			<-*sem
		}(file, &sem)
	}

	wg.Wait()

	return nil
}

func optimizeFile(file string, inputDir, outputDir string, width float32) error {
	ext := filepath.Ext(file)
	outputFile := strings.Replace(strings.Replace(file, inputDir, outputDir, 1), ext, "_optimized.webp", 1)
	fmt.Println("Optimizing file: ", file, outputFile)
	return ffmpeg.Input(file).Output(outputFile, ffmpeg.KwArgs{"compression_level": "6", "vf": fmt.Sprintf("scale=%f:-1", width)}).Run()
}
