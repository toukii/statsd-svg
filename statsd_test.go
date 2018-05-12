package statsd

import (
	"fmt"
	"testing"
	"text/template"

	"github.com/toukii/bezier"
	svg "github.com/toukii/bezier/bezier_in_svg"
	"github.com/toukii/goutils"
	"github.com/toukii/icat"
)

func TestDisplay(t *testing.T) {
	svgFmt := `<svg width="1800px" height="240px" viewBox="-10 -10 1500 230" style="background:white" version="1.1" xmlns="http://www.w3.org/2000/svg">
	<font>


	  <font-face font-family="Super Sans"/>
	</font>
	<rect x="-30" y="0" height="230" width="1500" style="stroke: #70d5dd; fill: #DCDCDC" />

	{{.Polyline}}
	<path d="{{.Path}}" stroke="{{.Color}}" stroke-width="{{.StrokeWidth}}" fill="none"></path>

	<g fill="none" stroke="black" stroke-width="1">
	{{.LineMin}}
	{{.LineAvg}}
	{{.LineAvgDbl}}
</g>
	<text x="0" y="25" font-family="Super Sans" style="font-size: 18pt;">{{.Remark}}</text>
</svg>`

	lineFmt := `<line x1="{{.From.X}}" y1="{{.From.Y}}" x2="{{.To.X}}" y2="{{.To.Y}}" style="stroke:rgb(0,0,0);stroke-width:1"></line>
<text x="0" y="{{.From.Y}}" font-family="Super Sans" style="font-size: 18pt;">{{.Remark}}</text>	
	`
	pathLineFmt := `<path stroke-dasharray="{{.Dash}}" d="M{{.From.X}} {{.From.Y}} L{{.To.X}} {{.To.Y}}" />
<text x="0" y="{{.From.Y}}" font-family="Super Sans" style="font-size: 18pt;">{{.Remark}}</text>`

	svgTpl, err := template.New("Path").Parse(svgFmt)
	if err != nil {
		panic(err)
	}

	lineTpl, err := template.New("Line").Parse(lineFmt)
	if err != nil {
		panic(err)
	}

	pathLineTpl, err := template.New("pathLine").Parse(pathLineFmt)
	if err != nil {
		panic(err)
	}

	// ps, stats := StatsdFromNet(1450, 200, "stats.sample.Count")
	ps, stats := StatsdSample(1450, 200)
	if ps == nil {
		return
	}

	th := float32(200) / float32(stats.Max)

	minHshow := 200 - int(th*float32(stats.Min))
	min := map[string]interface{}{
		"From":   bezier.NewPoint(0, minHshow),
		"To":     bezier.NewPoint(1500, minHshow),
		"Remark": fmt.Sprintf("min: %d", stats.Min),
		"Dash":   "5,5",
	}

	avgH := int(th * float32(stats.Avg))
	avgHshow := 200 - avgH
	avg := map[string]interface{}{
		"From":   bezier.NewPoint(0, avgHshow),
		"To":     bezier.NewPoint(1500, avgHshow),
		"Remark": fmt.Sprintf("avg: %d", stats.Avg),
		"Dash":   "5,10",
	}

	avgDblHshow := 200 - avgH<<1
	avgDbl := map[string]interface{}{
		"From":   bezier.NewPoint(0, avgDblHshow),
		"To":     bezier.NewPoint(1500, avgDblHshow),
		"Remark": fmt.Sprintf("%d", stats.Avg<<1),
		"Dash":   "5,20",
	}

	data := map[string]string{
		"Path":        goutils.ToString(bezier.Trhs(2, ps...)),
		"Color":       "#dd524b",
		"StrokeWidth": "1",

		"Polyline":   goutils.ToString(svg.MultiExcute(svg.PolylineTpl, ps...)),
		"LineMin":    goutils.ToString(svg.Excute(lineTpl, min)),
		"LineAvg":    goutils.ToString(svg.Excute(pathLineTpl, avg)),
		"LineAvgDbl": goutils.ToString(svg.Excute(pathLineTpl, avgDbl)),
		"Remark":     fmt.Sprintf("max: %d, min:%d, avg:%d", stats.Max, stats.Min, stats.Avg),
	}

	bs := svg.Excute(svgTpl, data)
	icat.DisplaySVG(bs)
}
