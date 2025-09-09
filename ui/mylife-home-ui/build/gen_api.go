package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gzuidhof/tygo/tygo"
)

const inputPath = "mylife-home-ui/pkg/web/api"
const outputPath = "webapp/src/app/api"

func main() {
	// List files to generate
	files, err := listGoFiles(inputPath)
	if err != nil {
		fmt.Printf("Error listing files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d .go files in %s:\n", len(files), inputPath)
	for _, file := range files {
		fmt.Printf("  %s\n", file)
	}
}

func listGoFiles(dirPath string) ([]string, error) {
	var goFiles []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			// Convert to relative path from the input directory
			relPath, err := filepath.Rel(dirPath, path)
			if err != nil {
				return err
			}
			goFiles = append(goFiles, relPath)
		}

		return nil
	})

	return goFiles, err
}

func generate(inputFile, outputFile string) error {
	config := &tygo.Config{
		Packages: []*tygo.PackageConfig{
			&tygo.PackageConfig{
				Path:       inputPath,
				OutputPath: path.Join(outputPath, outputFile),
				IncludeFiles: []string{
					inputFile,
				},
			},
		},
	}

	gen := tygo.New(config)
	return gen.Generate()

}
