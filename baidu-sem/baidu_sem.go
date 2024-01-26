package baiduSem

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/havenden/go-baidu-sem/baidu-sem/request"
	"github.com/havenden/go-baidu-sem/baidu-sem/response"
	"io"
	"math"
	"net/http"
	"sync"
)

var (
	GetUserListByMccidUrl = "https://api.baidu.com/json/feed/v1/MccFeedService/getUserListByMccid"
	GetReportDataUrl      = "https://api.baidu.com/json/sms/service/OpenApiReportService/getReportData"
)

type BaiduSem struct {
	userName    string
	accessToken string
}

func New(userName, accessToken string) *BaiduSem {
	return &BaiduSem{
		userName:    userName,
		accessToken: accessToken,
	}
}

// GetUserListByMccid 获取账户列表
func (baidu *BaiduSem) GetUserListByMccid() ([]response.BaiduAccountInfo, error) {
	var payload request.BaiduRequest
	payload.BaiduHeader.UserName = baidu.userName
	payload.BaiduHeader.AccessToken = baidu.accessToken
	bytesData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	// 发送api请求
	resp, err := http.Post(GetUserListByMccidUrl, "application/json", bytes.NewReader(bytesData))
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var baiduAccounts response.BaiduAccountListResponse
	err = json.Unmarshal(resBody, &baiduAccounts)
	if err != nil {
		return nil, err
	}
	if baiduAccounts.BaiduHeader.Status != 0 {
		return nil, errors.New(baiduAccounts.BaiduHeader.Failures[0].Message)
	}
	return baiduAccounts.BaiduAccountListBody.Data, nil
}

// GetBaiduAccountIds 获取百度所有账户的id 数组
func (baidu *BaiduSem) GetBaiduAccountIds() []int64 {
	accounts, _ := baidu.GetUserListByMccid()
	var baiduAccountIds []int64
	for _, account := range accounts {
		baiduAccountIds = append(baiduAccountIds, account.Userid)
	}
	return baiduAccountIds
}

// GetAllAccountReport
// 一站式多渠道报告
// reportType 报告类型，唯一标识一个报告
// timeUnit 时间单位 (HOUR: 小时;HOUR: 小时,WEEK: 周,MONTH: 月,SUMMARY: 时间段汇总)
// startDate 数据的起始日期，格式 2020-05-28
// endDate 数据的结束日期，格式 2020-05-28
// todo 其他参数暂未加
func (baidu *BaiduSem) GetAllAccountReport(reportType int, startDate, endDate, timeUnit string, city bool) (*[]response.CostInfo, error) {
	var ResData = make([]response.CostInfo, 0)
	responseData := &ResData
	var payload request.BaiduReportRequest
	payload.BaiduHeader.UserName = baidu.userName
	payload.BaiduHeader.AccessToken = baidu.accessToken
	payload.BaiduBody.ReportType = reportType
	payload.BaiduBody.StartDate = startDate
	payload.BaiduBody.EndDate = endDate
	payload.BaiduBody.TimeUnit = timeUnit
	payload.BaiduBody.StartRow = 0      // 从第几行开始获取结果
	payload.BaiduBody.RowCount = 100000 //要获取多少行，和startRow配合使用，用于分页获取数据。
	payload.BaiduBody.UserIds = baidu.GetBaiduAccountIds()
	var sort request.BaiduReportRequestSort
	sort.Column = "date"
	sort.SortRule = "ASC"
	payload.BaiduBody.Sorts = append(payload.BaiduBody.Sorts, sort)
	payload.BaiduBody.Columns = []string{"date", "userName", "impression", "click", "cost"}
	if city { //分地域
		payload.BaiduBody.Columns = append(payload.BaiduBody.Columns, "provinceCityName")
	}
	if reportType == 2290316 { // 分计划
		payload.BaiduBody.Columns = append(payload.BaiduBody.Columns, "campaignName")
	}
	bytesData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// 发送api请求
	resp, err := http.Post(GetReportDataUrl, "application/json", bytes.NewReader(bytesData))
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var baiduReportInfo response.BaiduReportResponse
	err = json.Unmarshal(resBody, &baiduReportInfo)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if baiduReportInfo.BaiduReportHeader.Status != 0 {
		return nil, errors.New(baiduReportInfo.BaiduReportHeader.Failures[0].Message)
	}
	for _, rs := range baiduReportInfo.BaiduReportBody.Data {
		for _, r := range rs.Rows {
			var item response.CostInfo
			item.Product = r.Product
			item.Name = r.UserName
			item.Date = r.Date
			item.Hour = r.Hour
			item.Plan = r.CampaignName
			item.City = r.ProvinceCityName
			item.Impression = r.Impression
			item.Click = r.Click
			item.Cost = r.Cost
			*responseData = append(*responseData, item)
		}
	}
	//totalRowCount:所有符合条件的数据总行数 rowCount:当前返回的数据行数
	for _, datum := range baiduReportInfo.BaiduReportBody.Data {
		totalRowCount := datum.TotalRowCount
		if totalRowCount > 100000 { //分页获取
			pages := int(math.Ceil(float64(totalRowCount / 100000)))
			for i := 1; i <= pages; i++ {
				startRow := i * 100000
				payload.BaiduBody.StartRow = startRow
				// 发送api请求
				resp, err := http.Post(GetReportDataUrl, "application/json", bytes.NewReader(bytesData))
				if err != nil {
					return nil, err
				}
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					if err != nil {
						panic(err)
					}
				}(resp.Body)
				resBody, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, err
				}
				var baiduReportInfo response.BaiduReportResponse
				err = json.Unmarshal(resBody, &baiduReportInfo)
				if err != nil {
					return nil, err
				}
				if baiduReportInfo.BaiduReportHeader.Status != 0 {
					return nil, errors.New(baiduReportInfo.BaiduReportHeader.Failures[0].Message)
				}
				for _, rs := range baiduReportInfo.BaiduReportBody.Data {
					for _, r := range rs.Rows {
						var item response.CostInfo
						item.Product = r.Product
						item.Name = r.UserName
						item.Date = r.Date
						item.Hour = r.Hour
						item.Plan = r.CampaignName
						item.City = r.ProvinceCityName
						item.Impression = r.Impression
						item.Click = r.Click
						item.Cost = r.Cost
						*responseData = append(*responseData, item)
					}
				}
			}
		}
	}
	return responseData, nil

}

// GetAccountReport  账户粒度消费报告
func (baidu *BaiduSem) GetAccountReport(startDate, endDate, timeUnit string) (*[]response.CostInfo, error) {
	reportData, err := baidu.GetAllAccountReport(2208157, startDate, endDate, timeUnit, false)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	return reportData, nil
}

// GetCampaignReport  计划粒度报告
func (baidu *BaiduSem) GetCampaignReport(startDate, endDate, timeUnit string) (*[]response.CostInfo, error) {
	reportData, err := baidu.GetAllAccountReport(2290316, startDate, endDate, timeUnit, false)
	if err != nil {
		return nil, err
	}
	return reportData, nil
}

// GetAreaReport 地域粒度报告
func (baidu *BaiduSem) GetAreaReport(startDate, endDate, timeUnit string) (*[]response.CostInfo, error) {
	reportData, err := baidu.GetAllAccountReport(2208157, startDate, endDate, timeUnit, true)
	if err != nil {
		return nil, err
	}
	return reportData, nil
}

// GetCampaignAreaReport 数据盘点，分计划和地域
func (baidu *BaiduSem) GetCampaignAreaReport(startDate, endDate, timeUnit string) (*[]response.CostInfo, error) {
	reportData, err := baidu.GetAllAccountReport(2290316, startDate, endDate, timeUnit, true)
	if err != nil {
		return nil, err
	}
	return reportData, nil
}

// GetAllReportData 报告数据类型： 1，账户报告 2，计划报告 3，地域报告
func (baidu *BaiduSem) GetAllReportData(startDate, endDate string, dataType []int) (map[int]*[]response.CostInfo, map[int]error) {
	var resData sync.Map
	var errs = make(map[int]error)
	var wg sync.WaitGroup
	var mutex sync.Mutex
	for _, t := range dataType {
		wg.Add(1)
		go func(t int) {
			defer wg.Done()
			var res *[]response.CostInfo
			var err error
			switch t {
			case 1: // 账户报告
				res, err = baidu.GetAccountReport(startDate, endDate, "DAY")
			case 2: // 计划报告
				res, err = baidu.GetCampaignReport(startDate, endDate, "DAY")
			case 3: // 地域报告
				res, err = baidu.GetAreaReport(startDate, endDate, "DAY")
			}
			if err != nil {
				mutex.Lock()
				errs[t] = err
				mutex.Unlock()
				return
			}
			mutex.Lock()
			defer mutex.Unlock()
			// 使用 sync.Map 安全地写入 resData
			resData.Store(t, res)
		}(t)
	}

	wg.Wait()
	// 从 sync.Map 中读取结果
	result := make(map[int]*[]response.CostInfo)
	resData.Range(func(key, value interface{}) bool {
		result[key.(int)] = value.(*[]response.CostInfo)
		return true
	})
	return result, errs
}
