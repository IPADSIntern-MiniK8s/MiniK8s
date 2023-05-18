package utils

/*
单位约定：
cpu  ： k8s的1000 = cpu的一个核

	如果一台服务器cpu是4核 那么 k8s单位表示就是 4* 1000

内存 : k8s的8320MI = 8320 * 1024 * 1024 字节

	1MI = 1024*1024 字节

	同理 1024MI /1024 = 1G
*/
type Quantity int
