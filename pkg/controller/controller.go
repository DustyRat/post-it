
package controller

import (
	"context"
	"errors"
	"text/template"
	"time"

	"github.com/DustyRat/post-it/pkg/options"

	"github.com/DustyRat/post-it/pkg/client"
	"github.com/DustyRat/post-it/pkg/file"
	"github.com/DustyRat/post-it/pkg/file/csv"
	"github.com/DustyRat/post-it/pkg/stats"
	"github.com/DustyRat/post-it/pkg/worker"

	"github.com/goinggo/work"
	"github.com/vbauerster/mpb"
)

// Controller ...
type Controller struct {
	Options  *options.Options
	Client   *client.Client
	Routines int
	Stats    *stats.Stats
	Writer   *csv.Writer
	template *template.Template
}

// Run ...
func (c *Controller) Run(headers []string, requests []*file.Data) error {
	c.Stats = stats.NewStats()
	//c.template = template.Must(template.New("text").Parse(textTemplate))
	wp, err := work.New(c.Routines, time.Hour*24, func(message string) {})
	if err != nil {
		return errors.New("error creating worker pools")
	}

	progress := mpb.New(mpb.WithContext(context.Background()))
	pool := worker.NewPool(c.Options, wp, c.Client, c.Stats, progress, int64(len(requests)), c.Writer)

	if c.Writer != nil {
		headers = append(headers, "status")
		if c.Options.Flags.Headers {
			headers = append(headers, "headers")
		}
		if c.Options.Flags.Body {
			headers = append(headers, "response_body")
		}
		headers = append(headers, "error")
		c.Writer.Write(headers)
	}

	for i := range requests {
		pool.NewWorker(requests[i])
	}
	pool.Run()
	//c.template.Execute(os.Stdout, c.Stats.Latencies)
	//c.Stats.Latencies.Print()
	return nil
}
