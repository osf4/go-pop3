# go-pop3
 A small Golang library for POP3 clients

## Installation
`go get -u github.com/osf4/go-pop3`

## Usage

```go
import (
	"fmt"
	"log"
	"time"

	"github.com/osf4/go-pop3"
)

func main() {
	opt := &pop3.Opt{
		DialTimeout: time.Second * 3,
		TLSEnabled:  true,
	}

	client, err := pop3.Dial("pop.mail.ru:995", opt)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	err = client.Auth("username", "password")
	if err != nil {
		log.Fatal(err)
	}

	count, size, err := client.Stat()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("count = %v\nsize = %v\n", count, size)
}
```
## Testing
You can use [InBucket](https://github.com/inbucket/inbucket) to run POP3 + SMTP server and test all the commands in the library