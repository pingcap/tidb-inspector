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
	"github.com/pborman/uuid"
	"github.com/pingcap/tidb-inspect-tools/metrics/grafana_collector/grafana"
	"github.com/signintech/gopdf"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
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
}

const (
	imgDir    = "images"
	reportPdf = "report.pdf"
)

// New creates a new Report.
func New(g grafana.Client, dashName string, time grafana.TimeRange) Report {
	return new(g, dashName, time)
}

func new(g grafana.Client, dashName string, time grafana.TimeRange) *report {
	tmpDir := filepath.Join("tmp", uuid.New())
	return &report{g, time, dashName, tmpDir}
}

// Generate returns the report.pdf file.  After reading this file it should be Closed()
// After closing the file, call report.Clean() to delete the file as well the temporary build files
func (rep *report) Generate() (pdf io.ReadCloser, err error) {
	dash, err := rep.gClient.GetDashboard(rep.dashName)
	if err != nil {
		err = fmt.Errorf("error fetching dashboard %v: %v", rep.dashName, err)
		return
	}
	err = rep.renderPNGsParallel(dash)
	if err != nil {
		err = fmt.Errorf("error rendering PNGs in parralel for dash %+v: %v", dash, err)
		return
	}
	pdf, err = rep.renderPDF(dash)
	return
}

// Clean deletes the temporary directory used during report generation
func (rep *report) Clean() {
	err := os.RemoveAll(rep.tmpDir)
	if err != nil {
		log.Println("Error cleaning up tmp dir:", err)
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
					log.Printf("Error creating image for panel: %v", err)
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
		return fmt.Errorf("error getting panel %+v: %v", p, err)
	}
	defer body.Close()

	err = os.MkdirAll(rep.imgDirPath(), 0777)
	if err != nil {
		return fmt.Errorf("error creating img directory:%v", err)
	}
	imgFileName := fmt.Sprintf("image%d.png", p.ID)
	file, err := os.Create(filepath.Join(rep.imgDirPath(), imgFileName))
	if err != nil {
		return fmt.Errorf("error creating image file:%v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, body)
	if err != nil {
		return fmt.Errorf("error copying body to file:%v", err)
	}
	return nil
}

func (rep *report) renderPDF(dash grafana.Dashboard) (outputPDF *os.File, err error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 595.28, H: 841.89}})
	pdf.AddPage()
	err = pdf.AddTTFFont("opensans", "ttf/OpenSans-Regular.ttf")
	if err != nil {
		log.Print(err.Error())
		return
	}
	err = pdf.SetFont("opensans", "", 14)
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	rectGraph := &gopdf.Rect{W: 480, H: 240}
	rectSinglestat := &gopdf.Rect{W: 300, H: 150}
	rect := &gopdf.Rect{}

	pdf.SetX(50)
	pdf.Cell(nil, "Dashboard: "+dash.Title)
	pdf.Br(20)
	pdf.SetX(50)
	pdf.Cell(nil, rep.time.FromFormatted()+" to "+rep.time.ToFormatted())

	panels := make(chan grafana.Panel, len(dash.Panels))
	for _, p := range dash.Panels {
		panels <- p
	}
	close(panels)

	var wg sync.WaitGroup
	var count int
	wg.Add(1)
	errs := make(chan error, len(dash.Panels))
	go func(panels <-chan grafana.Panel, errs chan<- error) {
		defer wg.Done()
		for p := range panels {
			imgFileName := fmt.Sprintf("image%d.png", p.ID)
			imgFilePath := filepath.Join(rep.imgDirPath(), imgFileName)

			if p.IsSingleStat() {
				rect = rectSinglestat
			} else {
				rect = rectGraph
			}

			if count%2 == 0 {
				err = pdf.Image(imgFilePath, 50, 80, rect)
			} else {
				err = pdf.Image(imgFilePath, 50, 350, rect)
				pdf.AddPage()
			}
			if err != nil {
				log.Printf("Error rendering image to PDF: %v", err)
				errs <- err
			} else {
				log.Printf("Rendering image to PDF: %s", imgFileName)
			}
			count++
		}
	}(panels, errs)

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			return nil, err
		}
	}

	pdf.WritePdf(rep.pdfPath())
	outputPDF, err = os.Open(rep.pdfPath())
	return outputPDF, err
}
