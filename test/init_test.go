package test

import (
	"fmt"
	nested "github.com/aohanhongzhi/nested-logrus-formatter"
	"testing"
)

// 这个包下面所有单元测试都会执行这个数据库初始化
func TestMain(m *testing.M) {
	nested.LogInit()
	fmt.Println("begin")
	m.Run()
	fmt.Println("end")
}
