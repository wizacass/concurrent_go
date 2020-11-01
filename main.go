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

func dataThread(size int, cIn, cOut chan car, getElement, control chan int) {
	arr := make([]car, size)
	counter := 0

	for {
		if counter == 0 {
			c, ok := <-cIn
			if !ok {
				break
			}
			arr[counter] = c
			counter++
		} else if counter >= size {
			<-getElement
			counter--
			cOut <- arr[counter]
		} else {
			select {
			case c := <-cIn:
				arr[counter] = c
				counter++
			case <-getElement:
				counter--
				cOut <- arr[counter]
			}
		}
	}

	close(cOut)

	ctrl := 0
	for i := range getElement {
		ctrl += i
		fmt.Println("Closing worker", ctrl)
	}
}

func workerThread(threshold float64, cIn chan car, cOut chan computedCar, controlIn, controlOut chan int) {
	for {
		controlIn <- 1
		c, ok := <-cIn
		if !ok {
			break
		}

		cc := getComputedCar(c)
		if cc.computedValue < threshold {
			cOut <- cc
		}
	}

	controlOut <- 1
}

func resultThread(cIn chan computedCar, cOut chan []computedCar) {
	var arr []computedCar
	for c := range cIn {
		arr = sortedInsert(arr, c)
	}
	cOut <- arr
}

func sortedInsert(arr []computedCar, c computedCar) []computedCar {
	arr = append(arr, c)

	for i := len(arr) - 1; i > 0; i-- {
		if arr[i].computedValue < arr[i-1].computedValue {
			t := arr[i]
			arr[i] = arr[i-1]
			arr[i-1] = t
		} else {
			break
		}
	}

	return arr
}

func run(dataFile string, resultsFile string) {
	processCount := 4

	dataIn := make(chan car)
	dataOut := make(chan car)
	workerOut := make(chan computedCar)
	resultsOut := make(chan []computedCar)
	dataControl := make(chan int)
	inputControl := make(chan int)
	workerControl := make(chan int, processCount)

	var threshold float64 = 5000
	var cars = read(dataFile)

	fmt.Println("Analyzing file", dataFile)

	go dataThread(len(cars)/2, dataIn, dataOut, inputControl, dataControl)
	for i := 0; i < processCount; i++ {
		go workerThread(threshold, dataOut, workerOut, inputControl, workerControl)
	}
	go resultThread(workerOut, resultsOut)

	for i := 0; i < len(cars); i++ {
		dataIn <- cars[i]
	}
	close(dataIn)

	control := 0
	for i := range workerControl {
		control += i
		if control >= processCount {
			close(workerOut)
			break
		}
	}

	cCars := <-resultsOut
	writeToFile(resultsFile, cCars)

	time.Sleep(100 * time.Millisecond)
}

func writeToFile(filename string, cCars []computedCar) {
	header := "|   Model  |  Price   | P. |  Initial  | Monthly |\n"
	sep := "+----------+----------+----+-----------+---------+\n"

	f, err := os.Create(filename)
	check(err)
	defer f.Close()

	f.WriteString(sep)
	f.WriteString(header)
	f.WriteString(sep)
	if len(cCars) == 0 {
		f.WriteString("|                  No elements!                  |\n")
		f.WriteString(sep)
		return
	}

	for _, c := range cCars {
		row := fmt.Sprintf(
			"| %8s | %8.2f | %2d | %9.2f | %7.2f |\n",
			c.car.Model, c.car.Price, c.car.Period, c.car.InitialPayment, c.computedValue)
		f.WriteString(row)
	}
	f.WriteString(sep)
	f.WriteString(fmt.Sprintf("Total elements: %v\n", len(cCars)))
}

func main() {
	fmt.Println("Hello!")
	defer fmt.Println("Done!")
	template := "data/IFF8-1_PetrauskasV_L2"
	filenames := 3

	for i := 1; i <= filenames; i++ {
		dataFile := fmt.Sprintf("%v_dat_%v.json", template, i)
		resultsFile := fmt.Sprintf("%v_rez_%v.txt", template, i)
		run(dataFile, resultsFile)
		fmt.Println()
	}
}
