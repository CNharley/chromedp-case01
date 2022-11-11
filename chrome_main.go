package main

import (
	"chromedp-case01/internal/biz"
	"context"
	"github.com/chromedp/cdproto/target"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

func main() {

	//增加打开浏览器对应的操作
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),  // 设置true 表示开启无头浏览器
		chromedp.Flag("mute-audio", true), // 关闭声音
	)

	allocatorContext, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// 创建上下文
	ctxt, cancel := chromedp.NewContext(allocatorContext, chromedp.WithErrorf(log.Printf))
	defer cancel()

	//业务步骤
	//自动登录
	//获取目录
	//根据目录找出哪些是没看的
	//根据找出的课程点击进入下一页
	//根据展示内容点击第一节课视频，跳转到视频播放页
	//循环获取当前页面视频时间，根据时间判断是否可以点击下一个视频
	//播放下一个课程
	//校验视频播放完成 关闭当前两个页面
	//开始播放下一个课程

	//SendKeys 根据ID输入对应value值到指定标签
	//Click 点击指定ID按钮登录
	//WaitVisible 等待页面加载指定ID元素 出现 roll_course
	//Sleep 等待，休眠
	//Nodes 根据Xpath获取对应元素集合
	var tbodyNodes []*cdp.Node
	if err := chromedp.Run(ctxt,
		chromedp.Navigate("https://jlufe.mhtall.com/jlufe//login/"),
		chromedp.Sleep(1*time.Second),
		chromedp.SendKeys(`txtLoginName`, "", chromedp.ByID), // 账号
		chromedp.SendKeys(`txtPassword`, "", chromedp.ByID),  //密码
		chromedp.Click(`rdoUserType3`, chromedp.ByID),
		chromedp.Click(`login_button`, chromedp.ByID),
		chromedp.WaitVisible(`roll_course`, chromedp.ByID),
		chromedp.Sleep(2*time.Second),
		chromedp.Click(`roll_course`, chromedp.ByID),
		chromedp.Sleep(4*time.Second),
		chromedp.Nodes(`//*[@id="app"]/div[2]/div[3]/table/tbody`, &tbodyNodes, chromedp.BySearch),
	); err != nil {
		log.Fatalf("Failed getting body of duckduckgo.com: %v", err)
	}

	//获取当前页面得TargetID,可以认为是唯一标识
	var mainTargetId string
	var aa, _ = chromedp.Targets(ctxt)
	for _, a := range aa {
		if strings.EqualFold(a.Title, "学生平台") {
			mainTargetId = a.TargetID.String()
		}
		log.Printf(a.URL)
	}

	//循环遍历课程元素集合
	if nil != tbodyNodes {
		for _, tbody := range tbodyNodes {
			var childNodeCount = int(tbody.ChildNodeCount)
		childNode:
			for q := 1; q <= childNodeCount; q++ {
				var lessonNodes []*cdp.Node
				//创建新得Context 上下文,并设置超时时间60秒
				ctxt1, cancel1 := context.WithTimeout(ctxt, time.Duration(60)*time.Second)
				defer cancel1()

				//根据新得上下文来操作获取元素 Tips:因为chromedp无法感知 页面是否真的完全加载完成，所以当获取一个不存在的元素时 可能存在卡死
				//这个时候就需要通过cancel1() 来关闭当前上下文Context，不会导致程序死循环 无法结束
				if err := chromedp.Run(ctxt1,
					chromedp.Nodes(`//*[@id="app"]/div[2]/div[3]/table/tbody/tr[`+strconv.Itoa(q)+`]/td[5]/div/el-div/div[2]`, &lessonNodes, chromedp.BySearch),
				); err != nil {
					log.Fatalf("1 Failed getting body of duckduckgo.com: %v", err)
					cancel1()
				}
				if nil == lessonNodes {
					continue
				}
				lessonNode := lessonNodes[0]
				if nil == lessonNode.Children {
					continue
				}

				//根据获取的课程元素虚幻
				for _, node := range lessonNode.Children {
					//校验当前循环的课程是否已经播放完
					if strings.Contains(node.NodeValue, "点播:100") {
						log.Printf("点播:100,跳过准备播放下一个课程")
						continue childNode
					}
					//根据没有播放完的课程 操作点击进入新页面
					if err := chromedp.Run(ctxt,
						chromedp.Sleep(3*time.Second),
						chromedp.Click(`//*[@id="app"]/div[2]/div[3]/table/tbody/tr[`+strconv.Itoa(q)+`]/td[4]/div/button`, chromedp.BySearch),
						chromedp.Sleep(4*time.Second),
					); err != nil {
						log.Fatalf("2 Failed getting body of duckduckgo.com: %v", err)
					}

					//通过最早获取到的父targetId(即上一个页面的id) 拿到它的子页面id，就是我们上面点击进入的新页面
					var childTargetId target.ID
					var ee, _ = chromedp.Targets(ctxt)
					for _, a := range ee {
						if strings.EqualFold(a.OpenerID.String(), mainTargetId) {
							childTargetId = a.TargetID
						}
						log.Printf(a.URL)
					}

					//获取到子id后可以通过子id创建新的Context 上下文,这样通过新的上下文操作 即便出问题了 操作cancel2() 也只是关闭当前页面
					//而不会影响我们上一页 课程列表页面，可以继续执行去看下一个课程
					ctxt2, cancel2 := chromedp.NewContext(ctxt, chromedp.WithTargetID(childTargetId))

					var taskPartList []*cdp.Node
					if err := chromedp.Run(ctxt2,
						chromedp.Sleep(4*time.Second),
						chromedp.Nodes(`//*[@id="task-part-list"]`, &taskPartList, chromedp.BySearch),
					); err != nil {
						log.Fatalf("3 Failed getting body of duckduckgo.com: %v", err)
					}

					//根据获取到的视频列表 点击第一个视频进入视频播放页面看视频
					if err := chromedp.Run(ctxt2,
						chromedp.Sleep(1*time.Second),
						chromedp.Click(`//*[@id="task-part-list"]/div[1]/div[2]/a/p/span[1]`, chromedp.BySearch),
					); err != nil {
						log.Fatalf("4 Failed getting body of duckduckgo.com: %v", err)
					}

					//同上获取新的子id 进行下一个页面的操作
					var childTargetId2 target.ID
					var bb, _ = chromedp.Targets(ctxt2)
					for _, a := range bb {
						if strings.EqualFold(a.OpenerID.String(), childTargetId.String()) {
							childTargetId2 = a.TargetID
						}
						log.Printf(a.URL)
					}

					//视频播放
					biz.PlayVideo(ctxt2, childTargetId2)

					//刷新上上一个页面
					if err := chromedp.Run(ctxt2,
						chromedp.Sleep(5*time.Second),
					); err != nil {
						log.Fatalf("5 Failed getting body of duckduckgo.com: %v", err)
					}
					//播放后操作关闭页面
					cancel2()

					//跳出循环
				}
			}
		}
	}

}
