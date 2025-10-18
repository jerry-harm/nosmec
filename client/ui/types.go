package ui

// AppInterface 应用接口，用于视图访问应用功能
type AppInterface interface {
	GetNostrClient() interface{}
	GetConfig() interface{}
}

// BaseView 基础视图接口
type BaseView interface {
	SetSize(width, height int)
}
