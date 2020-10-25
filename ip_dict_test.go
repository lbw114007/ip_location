package ip_location

import (
	"strconv"
	"strings"
	"testing"
)

type testData struct {
	IP      string
	Index   uint32
	BeginIP string
	EndIP   string
	Country string
	Area    string
}

var ipDataList = []testData{
	{
		IP:      "116.0.0.0",
		BeginIP: "116.0.0.0",
		EndIP:   "116.0.7.255",
		Country: "印度尼西亚",
		Area:    "", // CZ88.NET
		Index:   7907656,
	},
	{
		IP:      "116.230.60.1",
		BeginIP: "116.230.36.0",
		EndIP:   "116.230.84.255",
		Country: "上海市",
		Area:    "电信",
		Index:   7927305,
	},
	{
		IP:      "116.0.64.64",
		BeginIP: "116.0.64.64",
		EndIP:   "116.0.64.79",
		Country: "新加坡",
		Area:    "塔塔通信(Tata Communications)美洲公司(加拿大)BGP网络",
		Index:   7907698,
	},
	{
		IP:      "127.0.0.1",
		BeginIP: "127.0.0.1",
		EndIP:   "127.0.0.1",
		Country: "本机地址",
		Area:    "", // CZ88.NET
		Index:   8148099,
	},
	{
		IP:      "192.168.3.1",
		BeginIP: "192.168.0.0",
		EndIP:   "192.168.255.255",
		Country: "局域网",
		Area:    "对方和您在同一内部网",
		Index:   8548177,
	},
}

func TestFindIndex(t *testing.T) {
	IPDict := NewIPDict()
	for _, IPItem := range ipDataList {
		index := IPDict.findIndex(IPItem.IP)
		if index == IPItem.Index {
			t.Log(index)
		} else {
			t.Fail()
		}
	}
}

func TestFindIP(t *testing.T) {
	IPDict := NewIPDict()
	for _, IPItem := range ipDataList {
		res, err := IPDict.FindIP(IPItem.IP)
		if err != nil {
			t.Fatal(err)
		}
		if res.IP == IPItem.IP && res.BeginIP == IPItem.BeginIP && res.EndIP == IPItem.EndIP && res.Country == IPItem.Country && res.Area == IPItem.Area {
			t.Log(res)
		} else {
			t.Fatal(res)
		}
	}
}

func TestAll(t *testing.T) {
	IPDict := NewIPDict()
	for i := 1; i <= 255; i++ {
		for j := 1; j <= 255; j++ {
			for m := 1; m <= 255; m++ {
				for n := 1; n <= 255; n++ {
					ip := strconv.Itoa(i) + "." + strconv.Itoa(j) + "." + strconv.Itoa(m) + "." + strconv.Itoa(n)
					res, err := IPDict.FindIP(ip)
					if err != nil {
						t.Fatal(err)
					}
					t.Log(res)
					if res.EndIP != res.BeginIP && res.EndIP != ip {
						ips := strings.Split(res.EndIP, ".")
						i, _ = strconv.Atoi(ips[0])
						j, _ = strconv.Atoi(ips[1])
						m, _ = strconv.Atoi(ips[2])
						n, _ = strconv.Atoi(ips[3])
					}
				}
			}
		}
	}
}
