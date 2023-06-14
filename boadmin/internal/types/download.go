package types

import "time"

type DownloadCenterX struct {
	DownloadCenter
	UserName  string   `json:"userName"`
	CreatedAt JsonTime `json:"createdAt"`
}

type DownloadCenterQueryAllResponseX struct {
	List     []DownloadCenterX `json:"list"`
	PageNum  int               `json:"pageNum" gorm:"-"`
	PageSize int               `json:"pageSize" gorm:"-"`
	RowCount int64             `json:"rowCount"`
}

type DownloadReportCreate struct {
	ID           int64  `json:"id"`
	MerchantCode string `json:"merchantCode"`
	UserId       int64  `json:"userId"`
	IsAdmin      string `json:"isAdmin"`
	Status       string `json:"status"`
	ReqParam     string `json:"reqParam"`
	Type         string `json:"type"`
	FilePath     string `json:"filePath"`
	FileName     string `json:"fileName"`
	MissionName  string `json:"missionName"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type DownloadReportUpdate struct {
	DownloadCenter
	UpdatedAt time.Time
}

type CreateDownloadTask struct {
	Prefix       string `json:"Prefix"`
	Infix        string `json:"infix"`
	Suffix       string `json:"suffix"`
	IsAdmin      bool   `json:"isAdmin"`
	StartAt      string `json:"startAt, optional"`
	EndAt        string `json:"endAt, optional"`
	CurrencyCode string `json:"currencyCode"`
	MerchantCode string `json:"merchantCode"`
	UserId       int64  `json:"userId"`
	ReqParam     string `json:"reqParam"`
	Type         string `json:"type"`
}
