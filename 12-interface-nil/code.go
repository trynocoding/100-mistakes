package main

import (
	"errors"
	"fmt"
	"reflect"
)

// MultiError 是一个自定义错误类型
type MultiError struct {
	errors []error
}

func (m *MultiError) Error() string {
	if m == nil || len(m.errors) == 0 {
		return ""
	}
	var result string
	for i, err := range m.errors {
		if i > 0 {
			result += "; "
		}
		result += err.Error()
	}
	return result
}

// getMultiErrorReturnNil 返回 nil *MultiError
// 注意：这在接口层面会导致问题
func getMultiErrorReturnNil() *MultiError {
	return nil
}

// getErrorAsInterface 返回 error 接口类型的 nil *MultiError
// 这是陷阱所在！
func getErrorAsInterface() error {
	var me *MultiError
	return me // 返回 nil *MultiError，但接口类型是 error
}

// getRealNilError 返回真正的 nil error
func getRealNilError() error {
	return nil
}

// doSomething 返回 (error, error)，其中第二个是具体类型的 nil
func doSomething() (error, error) {
	var me *MultiError
	return nil, me // err2 是 "nil *MultiError" 作为 error 接口
}

// IsNil 检查接口或指针是否为 "真正的 nil"
func IsNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return rv.IsNil()
	case reflect.Interface:
		if rv.IsNil() {
			return true
		}
		// 非 nil 接口，检查其底层值
		return IsNil(rv.Elem().Interface())
	}
	return false
}

func main() {
	fmt.Println("=== 示例 1: 返回 *MultiError 类型 ===")
	err1 := getMultiErrorReturnNil()
	fmt.Printf("err1 == nil: %v (正确，因为是具体类型)\n", err1 == nil)
	fmt.Printf("err1 type: %T\n", err1)
	fmt.Printf("err1 value: %v\n\n", err1)

	fmt.Println("=== 示例 2: 返回 error 接口，但底层是 nil *MultiError ===")
	err2 := getErrorAsInterface()
	fmt.Printf("err2 == nil: %v (错误！应该是 true)\n", err2 == nil)
	fmt.Printf("err2 type: %T\n", err2)
	fmt.Printf("err2 value: %v\n\n", err2)

	fmt.Println("=== 示例 3: 返回真正的 nil error ===")
	err3 := getRealNilError()
	fmt.Printf("err3 == nil: %v (正确)\n", err3 == nil)
	fmt.Printf("err3 type: %T\n", err3)
	fmt.Printf("err3 value: %v\n\n", err3)

	fmt.Println("=== 示例 4: 多重返回值中的 nil *MultiError ===")
	_, err4 := doSomething()
	fmt.Printf("err4 == nil: %v (错误！应该是 true)\n", err4 == nil)
	fmt.Printf("err4 type: %T\n", err4)
	fmt.Printf("err4 value: %v\n\n", err4)

	fmt.Println("=== 使用 IsNil 辅助函数 ===")
	fmt.Printf("IsNil(err2): %v (正确识别为 nil)\n", IsNil(err2))
	fmt.Printf("IsNil(err4): %v (正确识别为 nil)\n", IsNil(err4))
	fmt.Printf("IsNil(err3): %v\n\n", IsNil(err3))

	fmt.Println("=== 验证 MultiError.Error() 在 nil 时的行为 ===")
	var nilMe *MultiError
	fmt.Printf("nilMultiError.Error(): %q\n", nilMe.Error())
	fmt.Printf("nilMultiError == nil: %v\n\n", nilMe == nil)

	// 创建一个非 nil 的 MultiError
	me := &MultiError{errors: []error{errors.New("error1"), errors.New("error2")}}
	fmt.Printf("non-nil MultiError.Error(): %q\n", me.Error())
}
