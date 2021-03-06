/*
Copyright 2011-2013 Frederic Langlet
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
you may obtain a copy of the License at

                http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"kanzi/function"
	"math/rand"
	"os"
	"time"
)

func main() {
	fmt.Printf("TestZRLT\n")
	TestCorrectness()
	TestSpeed()
}

func TestCorrectness() {
	fmt.Printf("Correctness test\n")

	for ii := 0; ii < 20; ii++ {
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		fmt.Printf("\nTest %v\n", ii)
		var arr []int

		if ii == 0 {
			arr2 := [64]int{30, 0, 4, 0, 30, 17, 240, 0, 0, 0, 12, 2, 23, 17, 254, 0, 254, 24, 0, 24, 20, 0, 17, 0, 254, 0, 254, 32, 0, 0, 13, 32, 255, 31, 15, 25, 0, 29, 8, 24, 20, 0, 0, 0, 245, 253, 0, 32, 19, 0, 243, 0, 0, 0, 13, 244, 24, 4, 0, 7, 25, 2, 0, 15}
			arr = arr2[0:len(arr2)]
		} else {
			arr = make([]int, 64)

			for i := range arr {
				val := rnd.Intn(100) - 16

				if val >= 33 {
					val = 0
				}

				arr[i] = val
			}
		}

		size := len(arr)
		input := make([]byte, size)
		output := make([]byte, size*3/2)
		reverse := make([]byte, size)

		for i := range output {
			output[i] = 0x55
		}

		for i := range arr {
			if i == len(arr)/2 {
				input[i] = 255
			} else {
				input[i] = byte(arr[i])
			}
		}

		ZRLT, err := function.NewZRLT()

		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			continue
		}

		fmt.Printf("\nOriginal: ")

		for i := range input {
			if i%100 == 0 {
				fmt.Printf("\n")
			}

			fmt.Printf("%v ", input[i])
		}

		fmt.Printf("\nCoded: ")
		srcIdx, dstIdx, err2 := ZRLT.Forward(input, output)

		if err2 != nil {
			fmt.Printf("\nEncoding error: %v\n", err2)
			continue
		}

		for i := uint(0); i < dstIdx; i++ {
			if i%100 == 0 {
				fmt.Printf("\n")
			}

			fmt.Printf("%v ", output[i])
		}

		fmt.Printf(" (Compression ratio: %v%%)", dstIdx*100/srcIdx)
		ZRLT, err2 = function.NewZRLT()

		if err2 != nil {
			fmt.Printf("\nError: %v\n", err2)
			continue
		}

		_, _, err2 = ZRLT.Inverse(output[0:dstIdx], reverse)

		if err2 != nil {
			fmt.Printf("\nDecoding error: %v\n", err2)
			continue
		}

		fmt.Printf("\nDecoded: ")
		ok := true

		for i := range input {
			if i%100 == 0 {
				fmt.Printf("\n")
			}

			fmt.Printf("%v ", reverse[i])

			if reverse[i] != input[i] {
				ok = false
			}
		}

		if ok == true {
			fmt.Printf("\nIdentical\n")
		} else {
			fmt.Printf("\nDifferent\n")
			os.Exit(1)
		}
	}
}

func TestSpeed() {
	iter := 50000
	size := 50000
	fmt.Printf("\n\nSpeed test\n")
	fmt.Printf("Iterations: %v\n", iter)

	for jj := 0; jj < 3; jj++ {
		input := make([]byte, size)
		output := make([]byte, len(input)*2)
		reverse := make([]byte, len(input))

		// Generate random data with runs
		n := 0
		var compressed uint
		var err error
		delta1 := int64(0)
		delta2 := int64(0)

		for n < len(input) {
			val := byte(rand.Intn(7))
			input[n] = val
			n++
			run := rand.Intn(128)
			run -= 100

			for run > 0 && n < len(input) {
				input[n] = val
				n++
				run--
			}
		}

		for ii := 0; ii < iter; ii++ {
			zrlt, _ := function.NewZRLT()
			before := time.Now()

			if _, compressed, err = zrlt.Forward(input, output); err != nil {
				fmt.Printf("Encoding error: %v\n", err)
				os.Exit(1)
			}

			after := time.Now()
			delta1 += after.Sub(before).Nanoseconds()
		}

		for ii := 0; ii < iter; ii++ {
			zrlt, _ := function.NewZRLT()
			before := time.Now()

			if _, _, err = zrlt.Inverse(output[0:compressed], reverse); err != nil {
				fmt.Printf("Decoding error: %v\n", err)
				os.Exit(1)
			}

			after := time.Now()
			delta2 += after.Sub(before).Nanoseconds()
		}

		idx := -1

		// Sanity check
		for i := range input {
			if input[i] != reverse[i] {
				idx = i
				break
			}
		}

		if idx >= 0 {
			fmt.Printf("Failure at index %v (%v <-> %v)\n", idx, input[idx], reverse[idx])
			os.Exit(1)
		}

		prod := int64(iter) * int64(size)
		fmt.Printf("\nZRLT encoding [ms]: %v", delta1/1000000)
		fmt.Printf("\nThroughput [MB/s]: %d", prod*1000000/delta1*1000/(1024*1024))
		fmt.Printf("\nZRLT decoding [ms]: %v", delta2/1000000)
		fmt.Printf("\nThroughput [MB/s]: %d", prod*1000000/delta2*1000/(1024*1024))
		println()
	}
}
