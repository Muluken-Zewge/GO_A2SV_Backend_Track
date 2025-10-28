package main

import "fmt"

func palindromCheck(s string) bool {
	// two pointer approach
	left := 0
	right := len(s) - 1

	for left <= right {
		if s[left] != s[right] {
			return false
		}
		left++
		right--
	}

	return true
}

func main() {
	word := "abcba"
	fmt.Println(palindromCheck(word))
}
