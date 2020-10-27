package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"time"
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

func dataThread(size int, c chan car) {
	fmt.Println("Data process!")
	// arr := make([]car, size)
	counter := 0
	for i := range c {
		counter++
		fmt.Println(i)
	}
	fmt.Println("Total:", counter)
}

func workerThread(threshold float64, cIn chan car, cOut chan computedCar, control chan int) {
	fmt.Println("Worker process!")
	defer fmt.Println("Worker done!")
	for c := range cIn {
		cc := getComputedCar(c)
		if cc.computedValue < threshold {
			cOut <- cc
		}
	}
	control <- 1
}

func resultThread(cIn chan computedCar, cOut chan []computedCar) {
	fmt.Println("Result process!")
	defer fmt.Println("Results done!")
	var arr []computedCar
	for c := range cIn {
		arr = append(arr, c)
	}
	cOut <- arr
}

func main() {
	fmt.Println("Hello!")
	defer fmt.Println("Done!")
	filename := "data/IFF8-1_PetrauskasV_L1_dat_1.json"
	processCount := 4
	ccIn := make(chan computedCar)
	ccOut := make(chan []computedCar)
	controlChan := make(chan int, processCount)

	var threshold float64 = 5000
	var cars = read(filename)
	cIn := make(chan car)

	for i := 0; i < processCount; i++ {
		go workerThread(threshold, cIn, ccIn, controlChan)
	}
	go resultThread(ccIn, ccOut)

	for i := 0; i < len(cars); i++ {
		cIn <- cars[i]
	}
	close(cIn)

	control := 0
	for i := 0; i < processCount; i++ {
		signal := <-controlChan
		control = control + signal
		// fmt.Println(control)
	}
	close(ccIn)

	cCars := <-ccOut
	fmt.Println("Computed cars:", len(cCars))
	// for i := 0; i < len(cCars); i++ {
	// 	cc := cCars[i]
	// 	ok := ""
	// 	if cc.computedValue < threshold {
	// 		ok = "OK"
	// 	}
	// 	outString := fmt.Sprintf("%6s %9.2f ", cc.car.Model, cc.computedValue)
	// 	fmt.Println(outString, ok)
	// }
	time.Sleep(100 * time.Millisecond)
}
