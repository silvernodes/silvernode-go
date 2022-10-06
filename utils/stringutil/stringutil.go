package stringutil

import (
	"strconv"
	"strings"

	"github.com/silvernodes/silvernode-go/utils/errutil"
)

func Snake(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:]))
}

func Camel(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if k == false && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || k == false) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:])
}

func UcFirst(str string) string {
	var upperStr string
	cp := []rune(str)
	for i := 0; i < len(cp); i++ {
		if i == 0 {
			if cp[i] >= 97 && cp[i] <= 122 {
				cp[i] -= 32
				upperStr += string(cp[i])
			} else {
				return str
			}
		} else {
			upperStr += string(cp[i])
		}
	}
	return upperStr
}

func SubString(str string, index int, length int) (string, bool) {
	runes := []rune(str)
	if index >= len(runes) {
		return "", false
	}
	if length == 0 {
		return "", true
	} else if length < 0 {
		return string(runes[index:]), true
	} else {
		if index+length > len(runes) {
			return "", false
		}
		return string(runes[index:length]), true
	}
}

func SplitFloat32(str string, sep string) ([]float32, error) {
	strs := strings.Split(str, sep)
	vals := make([]float32, 0, len(strs))
	for _, str := range strs {
		val64, err := strconv.ParseFloat(str, 32)
		if err != nil {
			return nil, errutil.Extend("string转float32出错:"+str, err)
		}
		vals = append(vals, float32(val64))
	}
	return vals, nil
}

func SplitFloat64(str string, sep string) ([]float64, error) {
	strs := strings.Split(str, sep)
	vals := make([]float64, 0, len(strs))
	for _, str := range strs {
		val64, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return nil, errutil.Extend("string转float64出错:"+str, err)
		}
		vals = append(vals, val64)
	}
	return vals, nil
}

func SplitInt(str string, sep string) ([]int, error) {
	strs := strings.Split(str, sep)
	vals := make([]int, 0, len(strs))
	for _, str := range strs {
		val, err := strconv.Atoi(str)
		if err != nil {
			return nil, errutil.Extend("string转int出错:"+str, err)
		}
		vals = append(vals, val)
	}
	return vals, nil
}

func SplitInt64(str string, sep string) ([]int64, error) {
	strs := strings.Split(str, sep)
	vals := make([]int64, 0, len(strs))
	for _, str := range strs {
		val64, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return nil, errutil.Extend("string转int64出错:"+str, err)
		}
		vals = append(vals, val64)
	}
	return vals, nil
}
