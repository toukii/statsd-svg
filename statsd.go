package statsd

import (
	"fmt"
	"os"

	"github.com/everfore/exc"
	"github.com/toukii/bezier"
	"github.com/toukii/goutils"
	"github.com/toukii/jsnm"
)

var (
	cmd = `curl 'http://stats.sample.me/api/datasources/proxy/1/render' -H 'Origin: https://stats.sample.me' -H 'Accept-Encoding: gzip, deflate, br' -H 'Accept-Language: zh-CN,zh;q=0.9,en;q=0.8' -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.139 Safari/537.36' -H 'Content-Type: application/x-www-form-urlencoded' -H 'Accept: application/json, text/plain, */*' -H 'Referer: https://stats.sample.me/dashboard/db/nadesico-customer?panelId=2&fullscreen&edit' -H 'Connection: keep-alive' --data 'target=%s&from=-5h&until=now&format=json&maxDataPoints=1432' --compressed --insecure -s`
)

type Stats struct {
	Max, Min, Avg    int
	SMax, SMin, SAvg int
}

func (s *Stats) Scale(th float32) {
	s.SMax, s.SMin, s.SAvg = int(th*float32(s.Max)), int(th*float32(s.Min)), int(th*float32(s.Avg))
}

func StatsdFromNet(width, height int, mesh string) ([]*bezier.Point, *Stats) {
	bs, err := exc.Bash(fmt.Sprintf(cmd, mesh)).Do()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
		return nil, nil
	}
	return StatsdFromBytes(width, height, bs)
}

func StatsdSample(width, height int) ([]*bezier.Point, *Stats) {
	return StatsdFromBytes(width, height, goutils.ReadFile("sample.data"))
}

func StatsdFromBytes(width, height int, bs []byte) ([]*bezier.Point, *Stats) {
	js := jsnm.BytesFmt(bs)
	if js == nil {
		os.Exit(0)
		return nil, nil
	}

	ps, _ := getPoints(js)
	return ps, Statsd(width, height, ps)
}

func Statsd(width, height int, ps []*bezier.Point) *Stats {
	stats, size := getStats(ps)
	if size <= 0 {
		return nil
	}

	th := float32(height) / float32(stats.Max)
	ScalePoints(th, width, height, size, ps)
	return stats
}

func ScalePoints(th float32, width, height, size int, ps []*bezier.Point) {
	padding := float32(width) / float32(size)
	for i, p := range ps {
		p.X = int(float32(i) * padding)
		p.Y = height - int(float32(p.Y)*th)
	}
}

func getPoints(js *jsnm.Jsnm) ([]*bezier.Point, []int64) {
	arr := js.ArrLoc(0).Get("datapoints").Arr()
	size := len(arr)
	if size <= 0 {
		return nil, nil
	}
	ps := make([]*bezier.Point, 0, size)
	tms := make([]int64, 0, size)

	var iv int
	var tm int64
	for i, it := range arr {
		cell := it.Arr()
		if len(cell) >= 2 {
			v := cell[0].MustFloat64()
			iv = int(v)
			tm = cell[1].MustInt64()
		}

		ps = append(ps, &bezier.Point{
			X: i,
			Y: iv,
		})
		tms = append(tms, tm)
	}

	return ps, tms
}

func getStats(ps []*bezier.Point) (*Stats, int) {
	stats := new(Stats)
	stats.Min = 1 << 31
	size := 0
	sum := 0
	for _, it := range ps {
		if it.Y <= 0 {
			continue
		}
		size++
		sum += it.Y
		if it.Y > stats.Max {
			stats.Max = it.Y
		} else if it.Y < stats.Min {
			stats.Min = it.Y
		}
	}

	if size <= 0 {
		size = 1
	}
	stats.Avg = sum / size

	return stats, size
}
