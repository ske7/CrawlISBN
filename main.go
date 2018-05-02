package main

import (
	"fmt"
	"log"
	my "my/crawlisbn/spider"
	"os"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/crackcomm/crawl"
	"github.com/tealeg/xlsx"
	"golang.org/x/net/context"
)

var trim = func(char rune) rune {
	if unicode.IsSpace(char) {
		return -1
	}
	if !unicode.IsDigit(char) {
		return -1
	}
	return char
}

func trimToNum(r int) bool {
	if n := r - '0'; n >= 0 && n <= 9 {
		return true
	}
	return false
}

func saveToExcel(filename string, slice [][]string) error {
	var file *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var err error

	file = xlsx.NewFile()
	sheet, err = file.AddSheet("Sheet1")
	if err != nil {
		fmt.Printf(err.Error())
	}
	row = sheet.AddRow()
	for i := 0; i < len(slice); i++ {
		row.WriteSlice(&slice[i], -1)
		row = sheet.AddRow()
	}
	err = file.Save(filename)
	if err != nil {
		return err
	}
	return nil
}

func readFromSourceExcel(filename string) (res []string, err error) {
	var file *xlsx.File

	file, err = xlsx.OpenFile(filename)
	if err != nil {
		return nil, err
	}

	for _, sheet := range file.Sheets {
		for _, row := range sheet.Rows {
			for _, cell := range row.Cells {
				str, _ := cell.String()
				str = strings.Map(trim, str)
				if str == "ISBN" || str == "" {
					continue
				}
				res = append(res, str)
			}
		}
	}
	return res, nil
}

func sortStrSlice(slice [][]string) {
	sort.SliceStable(slice[:], func(i, j int) bool {
		for x := range my.Outlist[i] {
			if slice[i][x] == slice[j][x] {
				continue
			}
			return slice[i][x] < slice[j][x]
		}
		return false
	})
}

func main() {
	var err error
	dir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Fatal error in getting start directory: %v\n", err))
	}
	lines := make([]string, 0)
	lines, err = readFromSourceExcel(dir + "/ISBN.xlsx")
	if err != nil {
		log.Fatal(err)
	}

	c := crawl.New(
		crawl.WithQueue(crawl.NewQueue(30000)),
		crawl.WithConcurrency(25),
		crawl.WithSpiders(my.Spider),
	)

	Headers := make(map[string]string)
	Headers["Accept-Language"] = "en-US,en;q=0.5"
	Headers["X-FORWARDED-FOR"] = "45.32.242.77"

	my.Outlist = make([][]string, 0, len(lines))
	my.CheckMap = make(map[string]bool)
	my.WG = new(sync.WaitGroup)
	for _, line := range lines {
		my.WG.Add(1)
		line2 := "https://www.bookfinder.com/search/?keywords=" + line + "&lang=&st=sh&ac=qr&submit="

		ctx := crawl.WithProxy(context.Background(), "45.32.242.77", "45.32.242.77")

		if err := c.Schedule(ctx, &crawl.Request{URL: line2,
			Header:    Headers,
			Callbacks: crawl.Callbacks(my.List)}); err != nil {
			log.Fatal(err)
		}
	}

	log.Print("Starting crawl")

	go func() {
		for err := range c.Errors() {
			my.WG.Done()
			log.Printf("Crawl error: %v", err)
		}
	}()

	c.Start()

	my.WG.Wait()
	fmt.Println("")
	fmt.Printf("Done crawl!")

	fmt.Println("")
	fmt.Printf("Saving results...")

	needtocheckslice := make([][]string, 0, len(lines))
	for _, v := range lines {
		_, ok := my.CheckMap[v]
		if !ok {
			needtocheckslice = append(needtocheckslice, []string{v, "", "", "", ""})
		}
	}
	if len(needtocheckslice) > 0 {
		sortStrSlice(needtocheckslice)
		if err = saveToExcel("needtocheckagain.xlsx", needtocheckslice); err != nil {
			log.Fatalf("Failed to write the needtocheckagain xlsx file: %v", err)
		}
	} else {
		if _, err := os.Stat(dir + "/needtocheckagain.xlsx"); !os.IsNotExist(err) {
			err := os.Remove(dir + "/needtocheckagain.xlsx")
			if err != nil {
				log.Fatalf("Cannot delete file needtocheckagain.xlsx")
			}
		}
	}

	sortStrSlice(my.Outlist)
	if err = saveToExcel("outFile.xlsx", my.Outlist); err != nil {
		log.Fatalf("Failed to write the output xlsx file: %v", err)
	}

	fmt.Println("")
	log.Println("Results saved!")
}
