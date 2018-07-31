

## 数据类型

```
type Response struct {
	Id       int64
	Event    string
	Data     []byte
}

type Client struct {
}

func New(addr string) *Client   // addr为lainlet地址, 如"192.168.77.21:9001"

func (c *Client) Get(uri string, timeout time.Duration) ([]byte, error)  // get请求

func (c *Client) Watch(uri string, ctx context.Context) (<-chan *Response, error) // watch请求

func (c *Client) Do(uri string, timeout time.Duration, watch bool) (io.ReadCloser, error) // rawrequest

```


## Demo

```golang
package main

import (
	"fmt"
	"golang.org/x/net/context"
	api "github.com/laincloud/lainlet/api/v2"
	"github.com/laincloud/lainlet/client"
	"log"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("lainlet addr required")
		return
	}
	info := new(api.GeneralCoreInfo)

	// get request
	c := client.New(os.Args[1])
	data, err := c.Get("/coreinfowatcher/?appname=registry", 0)
	if err != nil {
		panic(err)
	}
	info.Decode(data)
	fmt.Println(info.Data)

	// watch request
	ctx, _ := context.WithTimeout(context.Background(), time.Second*30) // 30 seconds timeout
	ch, err := c.Watch("/coreinfowatcher/?appname=registry&heartbeat=5", ctx)
	if err != nil {
		panic(err)
	}
	for event := range ch {
		fmt.Println("Get a event:")
		fmt.Println("    ", event.Id)
		fmt.Println("    ", event.Event)
		if event.Id != 0 { // id == 0 means error-event or heartbeat
			if err := info.Decode(event.Data); err != nil {
				log.Println(err.Error())
			} else {
				fmt.Println("    ", info.Data)
			}
		}
	}
}
```
