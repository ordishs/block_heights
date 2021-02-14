package fx

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

var rates map[uint32]float64

func init() {

	rates := make(map[uint32]float64)

	csvfile, err := os.Open("./fx/bsv_prices_daily.csv")
	if err != nil {
		log.Fatal(err)
	}

	defer csvfile.Close()

	r := csv.NewReader(csvfile)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		datenum, err := strconv.ParseUint(strings.ReplaceAll(record[0], "-", ""), 10, 32)
		if err != nil {
			log.Fatal(err)
		}

		rate, err := strconv.ParseFloat(record[1], 64)

		rates[uint32(datenum)] = rate
	}
}

func GetRate(datenum uint32) float64 {
	return rates[datenum]
}
