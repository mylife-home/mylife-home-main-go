package main

import (
	"os"
	"path"
	"strings"

	"github.com/gzuidhof/tygo/tygo"
)

const inputPath = "mylife-home-ui/pkg/web/api"
const outputPath = "webapp/src/app/api"

func main() {
	files, err := listFiles(inputPath)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		outputFile := strings.TrimSuffix(file, ".go") + ".ts"
		println("Generating", outputFile)

		err := generate(file, outputFile)
		if err != nil {
			panic(err)
		}
	}
}

func listFiles(dirPath string) ([]string, error) {
	var files []string

	// Read directory contents (only root level, no subdirectories)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		// Only process files (not directories) that end with .go
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			files = append(files, entry.Name())
		}
	}

	return files, nil
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
