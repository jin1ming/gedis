package server

// redis协议将传输的数据结构分为5种最小单元类型，单元结束时统一机上回车换行符\r\n
const (
	SingleLineReplay	=	"+"		// 单行字符开头
	MultiLineReplay		=	"$"		// 多行字符开头，后跟字符串长度
	IntReplay			=	":"		// 整数开头


)