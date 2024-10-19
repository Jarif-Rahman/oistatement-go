package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const usage_message string =`Usage:
  oistatement-go [OPTIONS] <markdown-file>

Statement generator for OI-like contests

Options:
  -h, -help
	Show help
`

const default_json_config string = `{
    "task_name": "task-name",
    "language": "language", 
    "contest": "contest"
}
`

//go:embed static
var static_files embed.FS

//go:embed template.html
var template_html string

func extract_static_files(dir []fs.DirEntry, path string, root string){
	os.Mkdir(filepath.Join(root, path), 0755)
	
	for _, file := range dir {
		if file.IsDir() {
			file_dir, _ := static_files.ReadDir(filepath.Join(path, file.Name()))
			extract_static_files(file_dir, filepath.Join(path, file.Name()), root)
		} else {
			content, _ := static_files.ReadFile(filepath.Join(path, file.Name()))
			os.WriteFile(filepath.Join(root, path, file.Name()), content, 0644)
		}
	}
}

var browser_list = [...]string{"chromium", "google-chrome", "brave"}

type JsonConfig struct {
	TaskName	string		`json:"task_name"`
	Language	string
	Contest		string
	Banner		string
	StaticDir 	string
	Content		string
}


func main() {
	var (
		browser string
		banner string
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage_message)
        flag.PrintDefaults()
	}

	flag.StringVar(&browser, "browser", "", "Chromium-based browser used for rendering pdf")
	flag.StringVar(&banner, "banner", "", "Banner of your olympiad")

	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	file := flag.Arg(0)
	fileInfo, err := os.Stat(file)
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: File %s does not exist\n", file)
		os.Exit(1)
	}
	if fileInfo.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is a directory, not a file\n", file)
		os.Exit(1)
	}

	filePath := filepath.Dir(file)
	fileExt := filepath.Ext(file)
	fileBasename := strings.TrimSuffix(filepath.Base(file), fileExt)

	if fileExt != ".md" {
		fmt.Fprint(os.Stderr, "Error: File extension must be .md\n")
		os.Exit(1)
	}

	jsonConfigFile := filepath.Join(filePath, fileBasename + ".json")
	jsonConfigFileInfo, err := os.Stat(jsonConfigFile)
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: %s does not exist\n", jsonConfigFile)
		fmt.Fprintf(os.Stderr, "Creating file %s. Please Modify it.\n", jsonConfigFile)
		err = os.WriteFile(jsonConfigFile, []byte(default_json_config), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Unable to create JSON config file %s\n", jsonConfigFile)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if jsonConfigFileInfo.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is a directory, not a file\n", jsonConfigFile)
		os.Exit(1)
	}

	var jsonConfig JsonConfig

	jsonConfigData, err := os.ReadFile(jsonConfigFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to read file %s\n", jsonConfigFile)
		os.Exit(1)
	}

	err = json.Unmarshal(jsonConfigData, &jsonConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to parse file %s\n", jsonConfigFile)
		os.Exit(1)
	}
	if jsonConfig.TaskName == "" || jsonConfig.Language == "" || jsonConfig.Contest == "" {
		fmt.Fprintf(os.Stderr, "Error: File %s is not valid\n", jsonConfigFile)
		os.Exit(1)
	}
	
	var banner_options = [...]string{
		filepath.Join(filePath, "banner.svg"),
		filepath.Join(filePath, "banner.png"),
	}

	if banner == "" {
		for _, b := range banner_options {
			if b != "XXX" {
				_, err := os.Stat(b)
				if err == nil {
					banner = b
					break
				}
			}
		}
		if banner == "" {
			fmt.Fprint(os.Stderr, "Error: No banner found.\nLooked for banner in following locations:\n")
			for _, b := range banner_options {
				if b != "XXX" {
					fmt.Fprintf(os.Stderr, "  %s\n", b)
				}
			}
			fmt.Fprintf(os.Stderr, "Please specify your banner using the -banner flag.\n")
			os.Exit(1)
		}
	}
	
	bannerInfo, err := os.Stat(banner)
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Banner %s does not exist\n", banner)
		os.Exit(1)
	}
	if bannerInfo.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is a directory, not a file\n", banner)
		os.Exit(1)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to read file %s\n", file)
		os.Exit(1)
	}

	jsonConfig.Banner, err = filepath.Abs(banner)
	jsonConfig.StaticDir = "static"
	jsonConfig.Content = string(content)
	
	if browser == "" {
		for _, b := range browser_list {
			_, err := exec.LookPath(b)
			if err == nil {
				browser = b
				break
			}
		}
		if browser == "" {
			fmt.Fprint(os.Stderr, "Error: No chromium-based browser found. Please specify your browser using the -browser flag.\n")
			os.Exit(1)
		}
	} else {
		_, err := exec.LookPath(browser)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Browser %s not found\n", browser)
			os.Exit(1)
		}
	}
	fmt.Fprintf(os.Stderr, "Using browser: %s\n", browser)

	fmt.Fprintf(os.Stderr, "Extracting files\n")
	tempDir, err := os.MkdirTemp("", "oistatement-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to create temporary directory\n")
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)
	tempFile, err := os.Create(filepath.Join(tempDir, "template.html"))

	md_template := template.Must(template.New("html-template").Parse(template_html))
	md_template.Execute(tempFile, jsonConfig)
	pdf_file := filepath.Join(filePath, fileBasename + ".pdf")

	static_files_dir, _ := static_files.ReadDir("static")
	extract_static_files(static_files_dir, "static", tempDir)

	cmd := exec.Command(
		browser,
		"-headless",
		"--disable-cpu",
		fmt.Sprintf("--print-to-pdf=%s", pdf_file),
		"--disable-extensions",
		"--no-pdf-header-footer",
		"--disable-popup-blocking", 
		"--run-all-compositor-stages-before-draw",
		"--disable-checker-imaging", 
		"--virtual-time-budget=10000",
		tempFile.Name(),
	)
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Error occurred while printing to pdf\n")
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	cmd2 := exec.Command(
		"cpdf", 
		"-add-text",
		fmt.Sprintf("%s (%%Page of %%EndPage)", jsonConfig.TaskName),
		"-font", "Arial", 
		"-color", "0.4 0.4 0.4", 
		"-font-size", "10", 
		"-bottomright", ".62in",
		pdf_file,
		"-o", pdf_file,
	)
	err = cmd2.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Error occurred while adding footer to pdf\n")
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "PDF file generated successfully\n")
}