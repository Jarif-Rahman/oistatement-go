# oistatement-go

Based on Rezwan Arefin's [oistatement](https://github.com/RezwanArefin01/oistatement). A tool to generate statements for OI-like contests.

## Installation
- Install [cpdf](https://github.com/coherentgraphics/cpdf-binaries) (If you are using arch, you can install it from [aur](https://aur.archlinux.org/packages/cpdf)).
- Make sure you have common fonts like Arial and Noto Sans installed.
- Install any chromium-based browser (for example: Chromium, Google Chrome, Brave).
- Download `oistatement-go` binary from the [releases](https://github.com/Jarif-Rahman/oistatement-go/releases/latest) page. This is the only file you need to run the program. You can start using oistatement-go by running: `./oistatement-go`. For ease of use, it is recommended that you copy this binary to a folder that is in your PATH (for example: `~/.local/bin`).

## Usage 
You can convert a markdown file into pdf statement using:
```bash
oistatement-go statement.md
```
This command expects a file named `banner.svg` or `banner.png` to be in the same folder as `statement.md`. If no banner file is found, an empty banner.svg file will be created. The command also expects a configuration file named `name-of-the-markdown-file.json` to be present in the folder. If this file is missing, a new file with default configuration will be created.

This tool is mainly used by BdOI. If you are looking for BdOI banner, you can find it [here](https://raw.githubusercontent.com/RezwanArefin01/oistatement/refs/heads/master/static/img/current_banner.svg).