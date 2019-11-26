package main

import (
	"flag"
	"fmt"
	"log"
	"time"
	"os"
	"encoding/hex"
	"strings"
	"strconv"

	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/currantlabs/ble"
	"github.com/currantlabs/ble/examples/lib/dev"
)

var (
	device = flag.String("device", "default", "implementation of ble")
	du     = flag.Duration("du", 100*time.Millisecond, "advertising duration, 0 for indefinitely")
	str    = flag.String("str", "hello world, and ble advertising demo message", "advertising message")
	doFlag = flag.Bool("do", false, "false -> advertising(default), true -> scan")
	dup    = flag.Bool("dup", true, "allow duplicate reported")
)

type Tweets struct {
    Mac,Count string
}

var tweet map[Tweets]string

func main() {
	tweet = make(map[Tweets]string)

	flag.Parse()

	if *doFlag == true {
		d, err := dev.NewDevice(*device)
		if err != nil {
		log.Fatalf("can't new device : %s", err)
		}
		ble.SetDefaultDevice(d)

		fmt.Printf("Scanning for %s...\n", *du)
		ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), *du))
		chkErr(ble.Scan(ctx, *dup, advHandler, nil))

		os.Exit(0)
	} 

	counter := 3
	mesc := 0
	strs := string(*str)

	src := []byte(strs)

	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)

	mess := fmt.Sprintf("%s", dst) + "0000"
	var decStr []string

	fmt.Println("src: ", strs)
	fmt.Println("enc: ", mess)

	charParse := 30
	
	if len(mess) > charParse {
		for i := 0; i * charParse < len(mess); i++ {
			if ((i + 1) * charParse) > len(mess) {
				splitMess := fmt.Sprintf("%d%d%s", mesc, i, mess[(i * charParse):])
				if len(splitMess) < charParse {
					for r := 0; len(splitMess) < charParse + 2; r++ {
						splitMess += "0"
					}
				}
				decStr = append(decStr, splitMess[0:8] + "-" + splitMess[8:13] + "-" + splitMess[13:18] + "-" + splitMess[18:23] + "-" + splitMess[23:])
			} else {
				splitMess := fmt.Sprintf("%d%d%s", mesc, i, mess[(i * charParse):((i + 1) * charParse)])
				decStr = append(decStr, splitMess[0:8] + "-" + splitMess[8:13] + "-" + splitMess[13:18] + "-" + splitMess[18:23] + "-" + splitMess[23:])
			}
		}
	} else {
		if len(mess) < charParse {
			for r := 0; len(mess) < charParse; r++ {
				mess += "0"
			}
		}
		decStr = append(decStr, "00" + mess[0:8] + "-" + mess[8:13] + "-" + mess[13:18] + "-" + mess[18:23] + "-" + mess[23:])
	}

	fmt.Println(decStr)

	d, err := dev.NewDevice("default")
	if err != nil {
		log.Fatalf("can't new device : %s", err)
	}
	ble.SetDefaultDevice(d)
	fmt.Printf("Advertising for %s...\n", *du)

	for r := 0; r < counter; r++ {
		for i := 0; i < len(decStr); i++ {
			TestSvcUUID := ble.MustParse(decStr[i])
			testSvc := ble.NewService(TestSvcUUID)
            
			ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), *du))
			chkErr(ble.AdvertiseNameAndServices(ctx, "SNNS", testSvc.UUID))
		}
	}

	os.Exit(0)
}

func chkErr(err error) {
	switch errors.Cause(err) {
	case nil:
	case context.DeadlineExceeded:
		fmt.Printf("done\n")
	case context.Canceled:
		fmt.Printf("canceled\n")
	default:
		log.Fatalf(err.Error())
	}
}

func advHandler(a ble.Advertisement) {
	if len(a.Services()) > 0 {
		if string(a.LocalName()) == "SNNS" {
			macs := fmt.Sprintf("%s", a.Address())
			svcs := fmt.Sprintf("%s", a.Services())
			svcs = strings.Replace(svcs, "[", "", -1)
			svcs = strings.Replace(svcs, "]", "", -1)

			if len(tweet[Tweets{macs, string(svcs[0])}]) == 0 {
				tweet[Tweets{macs, string(svcs[0])}] = "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
			}
			strCnt, _ := strconv.Atoi(string(svcs[1]))
			tweet[Tweets{macs, string(svcs[0])}] = tweet[Tweets{macs, string(svcs[0])}][0:strCnt * 30] + string(svcs[2:]) + tweet[Tweets{macs, string(svcs[0])}][(strCnt + 1) * 30:]
			if strings.LastIndex(string(svcs), "0000") == 28 {
				decStrs := strings.Replace(tweet[Tweets{macs, string(svcs[0])}], "0000", "", -1)
				rets, _ := hex.DecodeString(decStrs)
				fmt.Println(string(rets))
			}
		}
	}
}
