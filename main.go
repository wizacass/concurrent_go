package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
)

type car struct {
	Model          string
	Price          float64
	Period         int
	InitialPayment float64
}

type computedCar struct {
	car           car
	computedValue float64
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

func getComputedValue(c car) float64 {
	contractPrice := 300
	interestRate := 0.03
	realPrice := c.Price - c.InitialPayment
	fullPrice := realPrice*math.Pow((1+interestRate), float64(c.Period)) + float64(contractPrice)
	monthlyPayment := fullPrice / float64(c.Period)

	return math.Round(monthlyPayment*100) / 100
}

func getComputedCar(c car) computedCar {
	value := getComputedValue(c)
	compCar := computedCar{c, value}
	return compCar
}

func main() {
	filename := "data/IFF8-1_PetrauskasV_L1_dat_1.json"
	var threshold float64 = 5000

	var cars = read(filename)
	fmt.Println("Total cars: ", len(cars))
	for i := 0; i < len(cars); i++ {
		cc := getComputedCar(cars[i])
		ok := ""
		if cc.computedValue < threshold {
			ok = "OK"
		}
		outString := fmt.Sprintf("%6s %9.2f ", cc.car.Model, cc.computedValue)
		fmt.Println(outString, ok)
	}
}
