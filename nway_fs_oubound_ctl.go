package nway_fs_oubound_ctl

import (
	"errors"
	"fmt"
	"github.com/fiorix/go-eventsocket/eventsocket"
 
	"time"
)

//bool结果为是否为终止符，检查一个输入按键
func CheckADtmfEvent(c *eventsocket.Connection, t *time.Timer, EndDtmf string, TwoDtmfTimer int) (string, error, bool) {
	timer1 := time.NewTimer(time.Second * time.Duration(TwoDtmfTimer))
	for {
		ev, err := c.ReadEvent()
		if err != nil {
			 
			return "", err, false
		}
		if ev.Get("Event-Name") == "DTMF" {
			//有按键
			t.Stop()
			timer1.Stop()
			dtmf := ev.Get("Dtmf-Digit")
			if dtmf == EndDtmf {
				return "", nil, true
			} else {
				return dtmf, nil, false
			}
		}
	}
	<-timer1.C
	return "", errors.New("Wait key expired"), false

}

//按一次完整检测dtmf去检查
func CheckDtmfEvent(c *eventsocket.Connection, t *time.Timer, EndDtmf string, MaxDtmf, TwoDtmfTimer int) (string, error) {
	var dtmf string
	for i := 0; i < MaxDtmf; i++ {

		a_dtmf, err, res := CheckADtmfEvent(c, t, EndDtmf, TwoDtmfTimer)
		if err == nil && res == false {
			dtmf += a_dtmf
		} else {
			if err == nil && res {
				break
			} else {

				return "", err
			}
		}
	}
	return dtmf, nil
}

//按MaxFailure去get dtmf
func CheckDtmfEventMaxFailure(c *eventsocket.Connection, t *time.Timer, EndDtmf string, MaxDtmf, MaxTimer, TwoDtmfTimer, MaxFailure int) (string, error) {

	for i := 0; i < MaxFailure; i++ {
		dtmf, err := CheckDtmfEvent(c, t, EndDtmf, MaxDtmf, TwoDtmfTimer)
		//fmt.Println("The Dtmf:", dtmf)
		if err == nil {

			return dtmf, nil
		} else {
			t.Reset(time.Second * time.Duration(MaxTimer))
		}
	}
	c.Execute("exit", "", false)
	return "", errors.New("Mare than Max Failure")
}

//检测最长时间如果没有按键，则raise Max timer错
func CheckDtmfEventMaxTimer(c *eventsocket.Connection, EndDtmf string, MaxDtmf, MaxTimer, TwoDtmfTimer, MaxFailure int) (string, error) {

	timer2 := time.NewTimer(time.Second * time.Duration(MaxTimer))
	dtmf, err := CheckDtmfEventMaxFailure(c, timer2, EndDtmf, MaxDtmf, MaxTimer, TwoDtmfTimer, MaxFailure)
	if err == nil {
		return dtmf, err
	} else {
		return "", err
	}
	<-timer2.C
	return "", errors.New("More than Max timer")
	//整个按键超时，提示按键出错

}

/*
//播音取按键
//返回:
//参数:handle:会话handle
//filename:语音文件名称，多个文件以分号";"隔开,文件名称可以带.wav,扩展名,假如没有,默认是.pcm扩展名.可指定路径,假如没有,默认是语音程序的./data/system目录下
//EndDtmf:按键结束条件,比如"#"表示按#号结束输入,""表示没有结束按键条件//支持最大3个结束按键 比如 EndDtmf="*0#" 表示按 0，* 或者#都可以结束
//MaxDtmf:最大按键个数结束条件,0表示没有按键个数结束条件
//MaxTimer:按键最大时间结束条件,单位秒,0表示没有最大时间结束条件
//TwoDtmfTimer:两个按键间的间隔时间结束条件,单位秒,0表示没有两个按键间的间隔时间结束条件
//dtmf:收到的用户的按键(输出参数),包括结束按键的条件,比如"#"
//说明:假如只有播音,不收取按键 设置：MaxDtmf=0

*/
func PlayGetDigits(c *eventsocket.Connection, filename, invalidfile, EndDtmf string, MaxDtmf, MaxTimer, TwoDtmfTimer, MaxFailure int) (string, error) {
	var params, dtmf string
	params = fmt.Sprintf("1 %d %d %d %s %s %s", MaxDtmf, MaxFailure, MaxTimer*1000, EndDtmf, filename, invalidfile)
	//fmt.Println("play_and_get_digits ", params)
	_, err := c.Execute("play_and_get_digits", params, false)
	if err != nil {
		return "", err
	} else {

		for {
			_, err = c.ReadEvent()
			if err != nil {
				logger.Error("Read dtmf failed")
				return "", err
			}
			dtmf, err = CheckDtmfEventMaxTimer(c, EndDtmf, MaxDtmf, MaxTimer, TwoDtmfTimer, MaxFailure)
			if err != nil {
				return "", err
			} else {
				break
			}
		}

	}
	return dtmf, nil
}
