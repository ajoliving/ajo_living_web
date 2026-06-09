/*
 * Property address import command.
 * 1. Parse public legacy xls files through a header-aware Python helper.
 * 2. Import all English and Chinese building name columns into searchable records.
 * 3. Preserve addresses, district labels, block candidates, and completion year.
 */
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/database"
	"ajoliving_web/http_service/internal/logger"
	"ajoliving_web/http_service/internal/service"
)

// 1. main imports all file paths passed on the command line.
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/import-property-addresses <address.xls> [address.xls...]")
		os.Exit(1)
	}

	cfg := config.Load()
	log := logger.New(cfg)
	db, err := database.Open(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect database: %v\n", err)
		os.Exit(1)
	}
	if err := database.Migrate(db); err != nil {
		fmt.Fprintf(os.Stderr, "migrate database: %v\n", err)
		os.Exit(1)
	}

	runtime := &service.Runtime{Config: cfg, DB: db, Logger: log, Now: time.Now}
	propertyService := service.NewPropertyService(runtime)
	total := 0
	for _, path := range os.Args[1:] {
		rows, err := parseAddressFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse %s: %v\n", path, err)
			os.Exit(1)
		}
		imported, err := propertyService.ImportPropertyAddressRows(context.Background(), rows, path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "import %s: %v\n", path, err)
			os.Exit(1)
		}
		total += imported
		fmt.Printf("%s imported %d rows\n", filepath.Base(path), imported)
	}
	fmt.Printf("total imported %d rows\n", total)
}

// 2. parseAddressFile delegates legacy xls reading to Python xlrd.
func parseAddressFile(path string) ([]service.PropertyAddressSuggestion, error) {
	command := exec.Command("python3", "-c", propertyAddressParserScript, path)
	output, err := command.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, err
	}

	var rows []service.PropertyAddressSuggestion
	if err := json.Unmarshal(output, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

const propertyAddressParserScript = `
import json
import os
import re
import sys

try:
    import xlrd
except Exception as exc:
    raise SystemExit("xlrd is required: python3 -m pip install --user xlrd==2.0.1") from exc

path = sys.argv[1]
book = xlrd.open_workbook(path)
rows = []
seen = set()

def clean(value):
    if value is None:
        return ""
    if isinstance(value, float):
        if value.is_integer():
            return str(int(value))
        return str(value)
    return re.sub(r"\s+", " ", str(value)).strip()

def district_from_address(value):
    text = clean(value)
    if not text:
        return ""
    chunks = re.split(r"[,，]", text)
    tail = clean(chunks[-1])
    if tail and not re.search(r"\d", tail):
        return tail
    for chunk in reversed(chunks):
        chunk = clean(chunk)
        if chunk and not re.search(r"\d", chunk):
            return chunk
    return ""

def district_code(source_path, district_label):
    name = os.path.basename(source_path).lower()
    label = district_label.lower()
    if "nt" in name:
        return "new_territories"
    if any(item in label for item in ["kowloon", "tsim sha tsui", "mong kok", "kwun tong", "sham shui po"]):
        return "kowloon"
    return "hong_kong_island"

def display_name(zh, en, district):
    names = []
    if zh:
        names.append(zh)
    if en and en != zh:
        names.append(en)
    base = " ".join(names) if names else en or zh
    return f"{base} ({district})" if district else base

def pick_header(sheet):
    for row_index in range(min(sheet.nrows, 6)):
        values = [clean(sheet.cell_value(row_index, column)) for column in range(sheet.ncols)]
        if any("Building Name (1)" in value for value in values):
            return row_index
    return -1

def col(headers, label):
    for index, value in enumerate(headers):
        if label in value:
            return index
    return -1

for sheet in book.sheets():
    header_row = pick_header(sheet)
    if header_row < 0:
        continue
    headers = [clean(sheet.cell_value(header_row, column)) for column in range(sheet.ncols)]
    indexes = {
        "en1": col(headers, "Building Name (1)"),
        "zh1": col(headers, "樓宇名稱 (一)"),
        "en2": col(headers, "Building Name (2)"),
        "zh2": col(headers, "樓宇名稱 (二)"),
        "en3": col(headers, "Building Name (3)"),
        "zh3": col(headers, "樓宇名稱 (三)"),
        "addr_en1": col(headers, "Address (1)"),
        "addr_zh1": col(headers, "樓宇地址(一)"),
        "addr_en2": col(headers, "Address (2)"),
        "addr_zh2": col(headers, "樓宇地址(二)"),
        "addr_en3": col(headers, "Address (3)"),
        "addr_zh3": col(headers, "樓宇地址(三)"),
        "year": col(headers, "Year of Completion"),
        "remark": col(headers, "Remark"),
    }
    for row_index in range(header_row + 1, sheet.nrows):
        def cell(key):
            index = indexes.get(key, -1)
            return clean(sheet.cell_value(row_index, index)) if index >= 0 else ""

        en_names = [cell("en1"), cell("en2"), cell("en3")]
        zh_names = [cell("zh1"), cell("zh2"), cell("zh3")]
        addr_en = " ".join(item for item in [cell("addr_en1"), cell("addr_en2"), cell("addr_en3")] if item)
        addr_zh = " ".join(item for item in [cell("addr_zh1"), cell("addr_zh2"), cell("addr_zh3")] if item)
        year_text = cell("year")
        year = int(year_text) if year_text.isdigit() else 0
        district = district_from_address(addr_zh) or district_from_address(addr_en)
        base_en = en_names[0]
        base_zh = zh_names[0]
        if not base_en and not base_zh:
            continue
        address = addr_zh or addr_en
        if not address:
            continue
        block_names = []
        for value in [en_names[1], zh_names[1], en_names[2], zh_names[2]]:
            if value and value not in block_names:
                block_names.append(value)
        aliases = []
        for en, zh in [(en_names[0], zh_names[0]), (en_names[1], zh_names[1]), (en_names[2], zh_names[2])]:
            if en or zh:
                aliases.append((en, zh))
        if not aliases:
            aliases.append((base_en, base_zh))
        for en, zh in aliases:
            estate_en = en or base_en
            estate_zh = zh or base_zh
            estate_name = estate_zh or estate_en
            key = (os.path.basename(path), estate_name, address)
            if not estate_name or key in seen:
                continue
            seen.add(key)
            rows.append({
                "estate_name": estate_name,
                "estate_name_en": estate_en,
                "display_name": display_name(estate_zh, estate_en, district),
                "address_text": address,
                "address_text_en": addr_en,
                "district_code": district_code(path, district),
                "district_label": district,
                "region_code": "new_territories" if "nt" in os.path.basename(path).lower() else "hk_kowloon",
                "block_names": block_names,
                "completion_year": year,
                "remark": cell("remark"),
            })

print(json.dumps(rows, ensure_ascii=False))
`
