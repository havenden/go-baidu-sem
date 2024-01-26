package request

type BaiduHeader struct {
	UserName    string `json:"userName"`
	AccessToken string `json:"accessToken"`
}
type BaiduBody struct{}
type BaiduRequest struct {
	BaiduHeader BaiduHeader `json:"header"`
	BaiduBody   BaiduBody   `json:"body"`
}
type BaiduReportRequest struct {
	BaiduHeader BaiduHeader            `json:"header"`
	BaiduBody   BaiduReportRequestBody `json:"body"`
}
type BaiduReportRequestBody struct {
	UserIds    []int64                  `json:"userIds"`
	ReportType int                      `json:"reportType"`
	StartDate  string                   `json:"startDate"`
	EndDate    string                   `json:"endDate"`
	TimeUnit   string                   `json:"timeUnit"`
	Columns    []string                 `json:"columns"`
	StartRow   int                      `json:"startRow"`
	RowCount   int                      `json:"rowCount"`
	NeedSum    bool                     `json:"needSum"`
	Sorts      []BaiduReportRequestSort `json:"sorts"`
	Filters    []struct {
		Column   string `json:"column"`
		Operator string `json:"operator"`
		Values   []int  `json:"values"`
	} `json:"filters"`
}
type BaiduReportRequestSort struct {
	Column   string `json:"column"`
	SortRule string `json:"sortRule"`
}
