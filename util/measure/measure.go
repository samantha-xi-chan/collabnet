package measure

import (
	"time"
)

func MeasureExecutionTime(f func()) time.Duration {
	start := time.Now()
	f()
	end := time.Now()
	return end.Sub(start)
}

func MeasureExecutionTimeWithArgs(f func(...interface{}) interface{}, args ...interface{}) (interface{}, time.Duration) {
	start := time.Now()
	result := f(args...)
	end := time.Now()
	return result, end.Sub(start)
}

//func main() {
//	a := 42
//	b := "Hello, World!"
//
//	result, executionTime := MeasureExecutionTime(FunctionToMeasure, a, b)
//	fmt.Printf("执行时间: %v\n", executionTime)
//	fmt.Printf("函数结果: %v\n", result)
//}
