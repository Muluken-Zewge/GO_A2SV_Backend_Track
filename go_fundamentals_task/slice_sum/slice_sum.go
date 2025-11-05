package main

import "fmt"

func sliceSum(intSlice []int) int {
	sum := 0
	for _, num := range intSlice {
		sum += num
	}
	return sum
}

func main() {
	nums := []int{1, 2, 3, 4, 5}
	sum := sliceSum(nums)
	fmt.Println(sum)
}
