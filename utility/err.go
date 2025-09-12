package utility

import "fmt"

func PanicErr(err error) {
	if err != nil {
		panic(fmt.Errorf("发生 panic 错误: %s", err.Error()))
	}
}
