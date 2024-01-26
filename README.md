# go-baidu-sem
```
package main
import (
    "fmt"
    baiduSem "github.com/havenden/go-baidu-sem/baidu-sem"
)

func main() {
    s := baiduSem.New("BDCCuser", "accessToken")
    // 报告数据类型： 1，账户报告 2，计划报告 3，地域报告
    res := s.GetAllReportData("2024-01-22", "2024-01-22", []int{1, 2, 3})
    fmt.Println(res)
}
```