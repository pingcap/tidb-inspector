// Copyright 2018 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

/*
   Copyright 2016 Vastech SA (PTY) LTD

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package report

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/pborman/uuid"
	"github.com/pingcap/tidb-inspect-tools/grafana_collector/grafana"
	"github.com/signintech/gopdf"
)

// Report groups functions related to genrating the report.
// After reading and closing the pdf returned by Generate(), call Clean() to delete the pdf file as well the temporary build files
type Report interface {
	Generate() (pdf io.ReadCloser, err error)
	Clean()
}

type report struct {
	gClient  grafana.Client
	time     grafana.TimeRange
	dashName string
	tmpDir   string
	config   *tomlConfig
}

const (
	imgDir    = "images"
	reportPdf = "report.pdf"
)

// New ... creates a new Report
func New(g grafana.Client, dashName string, time grafana.TimeRange) Report {
	return new(g, dashName, time)
}

func new(g grafana.Client, dashName string, time grafana.TimeRange) *report {
	tmpDir := filepath.Join("tmp", uuid.New())
	return &report{g, time, dashName, tmpDir, ReportConfig}
}

// Generate returns the report.pdf file. After reading this file it should be Closed()
// After closing the file, call report.Clean() to delete the file
func (rep *report) Generate() (pdf io.ReadCloser, err error) {
	dash, err := rep.gClient.GetDashboard(rep.dashName)
	if err != nil {
		return nil, errors.Errorf("error fetching dashboard %v: %v", rep.dashName, err)
	}

	err = os.MkdirAll(rep.imgDirPath(), 0777)
	if err != nil {
		return nil, errors.Errorf("error creating image directory: %v", err)
	}

	err = rep.renderPNGsParallel(dash)
	if err != nil {
		return nil, errors.Errorf("error rendering PNGs in parralel for dash %+v: %v", dash, err)
	}

	pdf, err = rep.renderPDF(dash)
	if err != nil {
		return nil, errors.Errorf("error rendering pdf for dash %+v: %v", dash, err)
	}
	return pdf, nil
}

// Clean deletes the temporary directory used during report generation
func (rep *report) Clean() {
	err := os.RemoveAll(rep.tmpDir)
	if err != nil {
		log.Errorf("Error cleaning up tmp dir: %v", err)
	}
}

func (rep *report) imgDirPath() string {
	return filepath.Join(rep.tmpDir, imgDir)
}

func (rep *report) pdfPath() string {
	return filepath.Join(rep.tmpDir, reportPdf)
}

func (rep *report) renderPNGsParallel(dash grafana.Dashboard) error {
	//buffer all panels on a channel
	panels := make(chan grafana.Panel, len(dash.Panels))
	for _, p := range dash.Panels {
		panels <- p
	}
	close(panels)

	//fetch images in parrallel form Grafana sever.
	//limit concurrency using a worker pool to avoid overwhelming grafana
	//for dashboards with many panels.
	var wg sync.WaitGroup
	workers := 5
	wg.Add(workers)
	errs := make(chan error, len(dash.Panels)) //routines can return errors on a channel
	for i := 0; i < workers; i++ {
		go func(panels <-chan grafana.Panel, errs chan<- error) {
			defer wg.Done()
			for p := range panels {
				err := rep.renderPNG(p)
				if err != nil {
					log.Errorf("Error creating image for panel: %v", err)
					errs <- err
				}
			}
		}(panels, errs)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func (rep *report) renderPNG(p grafana.Panel) error {
	body, err := rep.gClient.GetPanelPng(p, rep.dashName, rep.time)
	if err != nil {
		return errors.Errorf("error getting panel %+v: %v", p, err)
	}
	defer body.Close()

	imgFileName := fmt.Sprintf("image%d.png", p.ID)
	file, err := os.Create(filepath.Join(rep.imgDirPath(), imgFileName))
	if err != nil {
		return errors.Errorf("error creating image file:%v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, body)
	if err != nil {
		return errors.Errorf("error copying body to file:%v", err)
	}
	return nil
}

func (rep *report) imgFilePath(p grafana.Panel) string {
	imgFileName := fmt.Sprintf("image%d.png", p.ID)
	imgFilePath := filepath.Join(rep.imgDirPath(), imgFileName)
	return imgFilePath
}

// NewPDF ... creates a new PDF and sets font
func (rep *report) NewPDF() (*gopdf.GoPdf, error) {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: rep.config.Rect["page"].Width, H: rep.config.Rect["page"].Height}})

	ttfPath := FontDir + rep.config.Font.Ttf
	err := pdf.AddTTFFont(rep.config.Font.Family, ttfPath)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	err = pdf.SetFont(rep.config.Font.Family, "", rep.config.Font.Size)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return pdf, nil
}

// createHomePage ... add Home Page for PDF
func (rep *report) createHomePage(pdf *gopdf.GoPdf, dash grafana.Dashboard) {
	pdf.AddPage()
	pdf.SetX(rep.config.Position.X)
	pdf.Cell(nil, "Dashboard: "+dash.Title)
	pdf.Br(rep.config.Position.Br)
	pdf.SetX(rep.config.Position.X)
	pdf.Cell(nil, rep.time.FromFormatted()+" to "+rep.time.ToFormatted())
}

func (rep *report) renderPDF(dash grafana.Dashboard) (outputPDF *os.File, err error) {
	log.Infof("PDF templates config: %+v\n", rep.config)

	pdf, err := rep.NewPDF()
	rep.createHomePage(pdf, dash)

	// setting rectangle size for grafana panel type: Graph/Singlestat
	rectGraph := &gopdf.Rect{W: rep.config.Rect["graph"].Width, H: rep.config.Rect["graph"].Height}
	rectSinglestat := &gopdf.Rect{W: rep.config.Rect["singlestat"].Width, H: rep.config.Rect["singlestat"].Height}
	rect := &gopdf.Rect{}

	var count int
	for _, p := range dash.Panels {
		imgPath := rep.imgFilePath(p)

		if p.IsSingleStat() {
			rect = rectSinglestat
		} else {
			rect = rectGraph
		}

		// Add two images on every page
		if count%2 == 0 {
			err = pdf.Image(imgPath, rep.config.Position.X, rep.config.Position.Y1, rect)
		} else {
			err = pdf.Image(imgPath, rep.config.Position.X, rep.config.Position.Y2, rect)
			pdf.AddPage()
		}
		if err != nil {
			log.Errorf("Error rendering image to PDF: %v", err)
		} else {
			log.Infof("Rendering image to PDF: %s", imgPath)
		}
		count++
	}

	pdf.WritePdf(rep.pdfPath())
	outputPDF, err = os.Open(rep.pdfPath())
	return outputPDF, err
}
