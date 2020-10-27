package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type car struct {
	Model          string
	Price          float32
	Period         int
	InitialPayment float32
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func read(filename string) []car {
	jsonFile, err := os.Open(filename)
	check(err)
	defer jsonFile.Close()

	var cars []car
	byteValue, err := ioutil.ReadAll(jsonFile)
	check(err)

	json.Unmarshal(byteValue, &cars)

	return cars
}

func main() {
	filename := "data/IFF8-1_PetrauskasV_L1_dat_1.json"

	fmt.Println("Hello, world!")

	var cars = read(filename)
	fmt.Println("Total cars: ", len(cars))
}
