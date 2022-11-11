package biz

import (
	"context"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"log"
	"strconv"
	"strings"
	"time"
)

func PlayVideo(mainCtxt context.Context, childTargetId target.ID) {

	// 创建新的上下文 进行操作 ，对上层页面不影响
	ctxt, cancel := chromedp.NewContext(mainCtxt, chromedp.WithTargetID(childTargetId))
	defer cancel()

	var nodess []*cdp.Node
	if err := chromedp.Run(ctxt,
		chromedp.Sleep(5*time.Second),
		chromedp.Nodes(`//*[@class="chapter-course js-chapter-course"]/li`, &nodess, chromedp.BySearch),
		chromedp.Sleep(10*time.Second),
	); err != nil {
		log.Fatalf("Failed getting body of duckduckgo.com: %v", err)
	}

aa:
	for _, node := range nodess {

		lessonId := node.Attributes[7]
		lessonIdStr := "lessonId-" + node.Attributes[7]

		spanFinishLenId := `//*[@id="spanFinishLen` + lessonId + `"]`
		spanTotalLenId := `//*[@id="spanTotalLen` + lessonId + `"]`
		var spanFinishLenNodes []*cdp.Node
		var spanTotalLenNodes []*cdp.Node
		if err := chromedp.Run(ctxt,
			chromedp.Nodes(spanFinishLenId, &spanFinishLenNodes, chromedp.BySearch),
			chromedp.Nodes(spanTotalLenId, &spanTotalLenNodes, chromedp.BySearch),
			chromedp.Sleep(2*time.Second),
		); err != nil {
			log.Fatalf("Failed getting body of duckduckgo.com: %v", err)
		}

		value := spanFinishLenNodes[0].Children[0].NodeValue
		value2 := spanTotalLenNodes[0].Children[0].NodeValue
		if strings.Contains(value2, value) {
			log.Printf("跳过准备播放下一个视频," + value2 + "  " + value)
			continue aa
		}

		log.Printf("点击下一个视频")
		if err := chromedp.Run(ctxt,
			chromedp.Click(`//*[@id="`+lessonIdStr+`"]/li`, chromedp.BySearch),
		); err != nil {
			log.Fatalf("Failed getting body of duckduckgo.com: %v", err)
		}
		log.Printf("点击下一个视频2")

		second := -1

	bb:
		for true {
			// run task list
			sel := `//*[@id="` + lessonIdStr + `"]/li`

			var nodes []*cdp.Node
			if err := chromedp.Run(ctxt,
				chromedp.Sleep(10*time.Second),
				chromedp.Nodes(sel, &nodes, chromedp.BySearch),
			); err != nil {
				log.Fatalf("Failed getting body of duckduckgo.com: %v", err)
			}

			for _, node := range nodes {
				s := node.Attributes[len(node.Attributes)-1]
				_int, _ := strconv.Atoi(s)

				log.Printf("上一次秒数:" + strconv.Itoa(second) + ";本次秒数:" + s)

				if second == _int {
					break bb
				} else {
					second = _int
				}

			}
		}
	}

}
