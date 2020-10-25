package ip_location

import (
	"encoding/binary"
	"errors"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// DefaultDict 默认字典
	DefaultDict = "./qqwry.dat"
	// IndexLen 索引长度
	IndexLen = 7
	// RedirectMode1 国家的类型, 指向另一个指向
	RedirectMode1 = 0x01
	// RedirectMode2 国家的类型, 指向一个指向
	RedirectMode2 = 0x02
)

type IPDict struct {
	fileData    []byte //文件数据
	offset      uint32 //当前下标定位
	firstOffset uint32 //第一条IP记录的偏移地址
	lastOffset  uint32 //最后一条IP记录的偏移地址
	totalIPNum  uint32 //IP记录的总条数（不包含版本信息记录）
}

type IPLocation struct {
	IP      string `json:"ip"`
	BeginIP string `json:"begin_ip"`
	EndIP   string `json:"end_ip"`
	Country string `json:"country"`
	Area    string `json:"area"`
}

func NewIPDict() *IPDict {
	return &IPDict{}
}

func (q *IPDict) Load(fileName string) error {
	filePath, err := dictPath(fileName)
	if err != nil {
		return err
	}
	dictFile, err := os.OpenFile(filePath, os.O_RDONLY, 0400)
	if err != nil {
		return err
	}
	defer dictFile.Close()
	q.fileData, err = ioutil.ReadAll(dictFile)
	if err != nil {
		return err
	}
	buf := q.readBuf(8)
	q.firstOffset = binary.LittleEndian.Uint32(buf[:4])
	q.lastOffset = binary.LittleEndian.Uint32(buf[4:])
	q.totalIPNum = (q.lastOffset - q.firstOffset) / IndexLen
	return nil
}

func (q *IPDict) FindIP(ip string) (*IPLocation, error) {
	if false == checkIPv4(ip) {
		return nil, errors.New("IP format error")
	}
	res := IPLocation{IP: ip}
	if nil == q.fileData {
		err := q.Load(DefaultDict)
		if nil != err {
			return nil, err
		}
	}
	q.seekOffset(0)
	index := q.findIndex(ip)
	if index <= 0 {
		return nil, errors.New("IP not fount")
	}
	q.seekOffset(index)
	res.BeginIP = long2ip(q.getIPLong4()) //endIPOffset
	endIPOffset := q.getRedirectOffset()
	q.seekOffset(endIPOffset)
	res.EndIP = long2ip(q.getIPLong4()) //endIPOffset
	mode := q.readMode()                // 标志字节
	var country, area []byte
	enc := simplifiedchinese.GBK.NewDecoder()
	switch mode {
	case RedirectMode1:                        // 标志字节为1，表示国家和区域信息都被同时重定向
		countryOffset := q.getRedirectOffset() // 重定向地址
		q.seekOffset(countryOffset)
		mode2 := q.readMode() // 标志字节
		switch mode2 {
		case RedirectMode2:                     // 标志字节为2，表示国家信息又被重定向
			q.seekOffset(q.getRedirectOffset()) // 重定向地址
			country = q.readString(0)
			q.seekOffset(countryOffset + 4) // 重定向地址
			area = q.readArea()
		default: // 否则，表示国家信息没有被重定向
			country = q.readString(mode2)
			area = q.readArea()
		}
	case RedirectMode2:                     // 标志字节为2，表示国家信息被重定向
		q.seekOffset(q.getRedirectOffset()) // 重定向地址
		country = q.readString(0)
		q.seekOffset(endIPOffset + 8)
		area = q.readArea()
	default:
		country = q.readString(mode)
		area = q.readArea()
	}
	countryUTF8, _ := enc.String(string(country))
	if strings.Trim(countryUTF8, " ") == "CZ88.NET" {
		res.Country = ""
	} else {
		res.Country = countryUTF8
	}
	areaUTF8, _ := enc.String(string(area))
	if strings.Trim(areaUTF8, " ") == "CZ88.NET" {
		res.Area = ""
	} else {
		res.Area = areaUTF8
	}
	return &res, nil
}

func (q *IPDict) findIndex(ip string) uint32 {
	if false == checkIPv4(ip) {
		return 0
	}
	if nil == q.fileData {
		err := q.Load(DefaultDict)
		if nil != err {
			return 0
		}
	}
	uIP := ip2long(ip)        // 将输入的IP地址转化为可比较的IP地址
	min := uint32(0)          // 搜索的下边界
	max := q.totalIPNum       // 搜索的上边界
	findIndex := q.lastOffset // 如果没有找到就返回最后一条IP记录（IPDict.Dat的版本信息）
	for min <= max {
		// 当上边界小于下边界时，查找失败
		mid := (min + max) / 2 // 计算近似中间记录
		q.seekOffset(q.firstOffset + mid*IndexLen)
		cBeginIP := q.getIPLong4() // 获取中间记录的开始IP地址
		if uIP < cBeginIP { // 用户的IP小于中间记录的开始IP地址时
			max = mid - 1 // 将搜索的上边界修改为中间记录减一
		} else {
			q.seekOffset(q.getRedirectOffset())
			cEndIP := q.getIPLong4() // 获取中间记录的开始IP地址
			if uIP > cEndIP { // 用户的IP大于中间记录的结束IP地址时
				min = mid + 1 // 将搜索的下边界修改为中间记录加一
			} else {
				// 用户的IP在中间记录的IP范围内时
				findIndex = q.firstOffset + mid*IndexLen
				break // 则表示找到结果，退出循环
			}
		}
	}
	return findIndex
}

//模拟文件读取Seek
func (q *IPDict) seekOffset(offset uint32) {
	q.offset = offset
}

//模拟文件读取Read
func (q *IPDict) readBuf(length uint32) []byte {
	q.offset = q.offset + length
	return q.fileData[q.offset-length : q.offset] // 标志字节
}

//返回读取的长整型数
func (q *IPDict) getIPLong4() uint32 {
	buf := q.readBuf(4)
	return binary.LittleEndian.Uint32(buf)
}

//返回读取的3个字节的长整型数
func (q *IPDict) getRedirectOffset() uint32 {
	buf := q.readBuf(3)
	return binary.LittleEndian.Uint32([]byte{buf[0], buf[1], buf[2], 0})
}

// readString 获取字符
func (q *IPDict) readMode() byte {
	return q.readBuf(1)[0] // 标志字节
}

// readString 获取字符串
func (q *IPDict) readString(char byte) []byte {
	data := make([]byte, 0, 30)
	if char != 0 {
		data = append(data, char)
	}
	buf := q.readBuf(1)
	for buf[0] != 0 {
		data = append(data, buf[0])
		buf = q.readBuf(1)
	}
	return data
}

// readArea 获取地区字符串
func (q *IPDict) readArea() []byte {
	mode := q.readMode()
	switch mode { // 标志字节
	case 0: // 结束标识
		return []byte{}
	case RedirectMode1:
	case RedirectMode2:                     // 标志字节为1或2，表示区域信息被重定向
		q.seekOffset(q.getRedirectOffset()) // 重定向地址
		return q.readString(0)
	}
	return q.readString(mode)
}

//数值IP转换字符串IP
func long2ip(ipInt uint32) string {
	// need to do two bit shifting and “0xff” masking
	ipInt64 := int64(ipInt)
	b0 := strconv.FormatInt((ipInt64>>24)&0xff, 10)
	b1 := strconv.FormatInt((ipInt64>>16)&0xff, 10)
	b2 := strconv.FormatInt((ipInt64>>8)&0xff, 10)
	b3 := strconv.FormatInt(ipInt64&0xff, 10)
	return b0 + "." + b1 + "." + b2 + "." + b3
}

//字符串IP转换数值IP
func ip2long(ip string) uint32 {
	bIP := net.ParseIP(ip).To4()
	if nil == bIP {
		return 0
	}
	return binary.BigEndian.Uint32(bIP)
}

//返回字典绝对路径
func dictPath(dictFileName string) (string, error) {
	if filepath.IsAbs(dictFileName) {
		return dictFileName, nil
	}
	var dictFilePath string
	cwd, err := os.Getwd()
	if err != nil {
		return dictFilePath, err
	}
	dictFilePath = filepath.Clean(filepath.Join(cwd, dictFileName))
	return dictFilePath, nil
}

//检查ip地址
func checkIPv4(IP string) bool {
	// 字符串这样切割
	strList := strings.Split(IP, ".")
	if len(strList) != 4 {
		return false
	}
	for _, s := range strList {
		if len(s) == 0 || (len(s) > 1 && s[0] == '0') {
			return false
		}
		// 直接访问字符串的值
		if s[0] < '0' || s[0] > '9' {
			return false
		}
		// 字符串转数字
		n, err := strconv.Atoi(s)
		if err != nil {
			return false
		}
		if n < 0 || n > 255 {
			return false
		}
	}
	return true
}
