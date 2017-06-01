package collectors

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	autoctx "github.com/vladimirvivien/automi/api/context"
)

type WriterCollector struct {
	wrtParam io.Writer
	writer   *bufio.Writer
	input    <-chan interface{}
	log      *log.Logger
}

func Writer(writer io.Writer) *WriterCollector {
	return &WriterCollector{
		wrtParam: writer,
	}
}

func (c *WriterCollector) SetInput(in <-chan interface{}) {
	c.input = in
}

func (c *WriterCollector) Open(ctx context.Context) <-chan error {
	c.log = autoctx.GetLogger(ctx)
	c.log.Print("opening io.Writer collector")
	result := make(chan error)

	if err := c.setupWriter(); err != nil {
		result <- err
		return result
	}

	go func() {
		defer func() {
			c.writer.Flush() //TODO handle error
			close(result)
			c.log.Print("closing io.Writer collector")
		}()

		for val := range c.input {
			switch data := val.(type) {
			case string:
				fmt.Fprint(c.writer, data)
			case []byte:
				if _, err := c.writer.Write(data); err != nil {
					c.log.Println(err)
					//TODO runtime error handling
					continue
				}
			default:
				c.log.Printf("unexpected type %T, needs []byte", data)
			}
		}
	}()

	return result
}

func (c *WriterCollector) setupWriter() error {
	if c.wrtParam == nil {
		return errors.New("missing io.Writer parameter")
	}
	c.writer = bufio.NewWriter(c.wrtParam)

	return nil
}