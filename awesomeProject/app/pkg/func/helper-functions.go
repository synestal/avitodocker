package _func

import "regexp"

func IsNumeric(s string) bool {
	matched, _ := regexp.MatchString("^[0-9]+$", s)
	return matched
}

func AllNumeric(slice []string) bool {
	for _, str := range slice {
		if !IsNumeric(str) {
			return false
		}
	}
	return true
}

func Intersection(slice1, slice2 []int) []int {
	set := make(map[int]bool)
	for _, num := range slice1 {
		set[num] = true
	}
	var result []int
	for _, num := range slice2 {
		if set[num] {
			result = append(result, num)
		}
	}

	return result
}
