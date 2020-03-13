package wxpay

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"weqi_service/api"
	"weqi_service/serializer"
)

const (
	APPID = "wx9ac41132c5635499"
	APPSECRET = "a32eb302c156b45c1d856ec56ee963c4"
	WXOrderURL  = "https://api.mch.weixin.qq.com/pay/unifiedorder"
)

//WXOrderParam	微信请求参数
type WXOrderParam struct {
	APPID          string `xml:"appid"`            //公众账号ID
	MchID          string `xml:"mch_id"`           //商户号
	NonceStr       string `xml:"nonce_str"`        //随机字符串
	Sign           string `xml:"sign"`             //签名
	Body           string `xml:"body"`             //商品描述
	OutTradeNo     string `xml:"out_trade_no"`     //商户订单号
	TotalFee       string `xml:"total_fee"`        //总金额
	SpbillCreateIP string `xml:"spbill_create_ip"` //终端IP
	NotifyUrl      string `xml:"notify_url"`       //通知地址
	TradeType      string `xml:"trade_type"`       //交易类型
	SceneInfo      string `xml:"scene_info"`       //场景信息
}

//WXOrderReply	微信请求返回结果
type WXOrderReply struct {
	ReturnCode string `xml:"return_code"` //返回状态码
	ReturnMsg  string `xml:"return_msg"`  //返回信息

	APPID      string `xml:"appid"`        //公众账号ID
	MchID      string `xml:"mch_id"`       //商户号
	DeviceInfo string `xml:"device_info"`  //设备号
	NonceStr   string `xml:"nonce_str"`    //随机字符串
	Sign       string `xml:"sign"`         //签名
	ResultCode string `xml:"result_code"`  //业务结果
	ErrCode    string `xml:"err_code"`     //错误代码
	ErrCodeDes string `xml:"err_code_des"` //错误代码描述

	TradeType string `xml:"trade_type"` //交易类型
	PrepayID  string `xml:"prepay_id"`  //预支付交易会话标识
	MwebURL   string `xml:"mweb_url"`   //支付跳转链接
}



func Callback(c *gin.Context) {
	code := c.Query("code")

	fmt.Println("get open id: ", code)

	if code == "" {
		fmt.Println("can't get code")
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "无法得到code",
		})
		return
	}

	GetOpenID(code)

	//otp-O1Vt3-fFPRSleTx2u9pUlnG0
}

type WXPayUserOpenID struct {
	OpenID string `json:"openid"`
}

// 每个User都有一个唯一的openID 需要把他存到数据库中
func GetOpenID(code string) {
	req_openid := "https://api.weixin.qq.com/sns/oauth2/access_token?appid=" + APPID + "&secret=" + APPSECRET + "&code=" + code + "&grant_type=authorization_code"

	//new request
	req, err := http.NewRequest(http.MethodGet, req_openid, nil)
	if err != nil {
		log.Println(err)
		return
	}

	//http client
	client := &http.Client{}
	log.Printf("Go %s URL : %s \n",http.MethodGet, req.URL.String())

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Can't set get request: ", err)
		return
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ioutil.ReadAll %v", err)
		return
	}
	fmt.Println(string(respBytes))

	openID := WXPayUserOpenID{}
	err = json.Unmarshal(respBytes, &openID)
	if err != nil {
		fmt.Println("Can't get openid: ", err)
		return
	}
	api.SmartPrint(openID)
}



func UnifiedOrder() {
	//url := "https://api.mch.weixin.qq.com/pay/unifiedorder"

}


/*
<xml>
   <appid>wx2421b1c4370ec43b</appid>
   <attach>支付测试</attach>
   <body>JSAPI支付测试</body>
   <mch_id>10000100</mch_id>
   <detail><![CDATA[{ "goods_detail":[ { "goods_id":"iphone6s_16G", "wxpay_goods_id":"1001", "goods_name":"iPhone6s 16G", "quantity":1, "price":528800, "goods_category":"123456", "body":"苹果手机" }, { "goods_id":"iphone6s_32G", "wxpay_goods_id":"1002", "goods_name":"iPhone6s 32G", "quantity":1, "price":608800, "goods_category":"123789", "body":"苹果手机" } ] }]]></detail>
   <nonce_str>1add1a30ac87aa2db72f57a2375d8fec</nonce_str>
   <notify_url>http://wxpay.wxutil.com/pub_v2/pay/notify.v2.php</notify_url>
   <openid>oUpF8uMuAJO_M2pxb1Q9zNjWeS6o</openid>
   <out_trade_no>1415659990</out_trade_no>
   <spbill_create_ip>14.23.150.211</spbill_create_ip>
   <total_fee>1</total_fee>
   <trade_type>JSAPI</trade_type>
   <sign>0CB01533B8C1EF103065174F50BCA001</sign>
</xml>


*/

/*
*	submitWXOrder	提交微信订单
*	param	data	WXOrderParam
*	reply	prepay_id	预支付交易会话标识
*	reply	mweb_url	支付跳转链接
 */
func submitWXOrder() (prepay_id string, mweb_url string) {

	order := WXOrderParam{
		APPID: "",
		MchID: "",
		Attach:"",
		Body : "",                      //商品描述
		Detail: "",
		NonceStr: "",        //随机字符串
		NotifyUrl : "",     //通知地址
		OpenID:"",
		OutTradeNo : "",       //商户订单号
		SpbillCreateIP : "", //终端IP
		TotalFee   : "",           //总金额
		Sign: "",                       //签名
		TradeType  : "",         //交易类型
		//SceneInfo  : "",        //场景信息
	}


	xdata, err := xml.Marshal(order)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(xdata))
	xmldata := strings.Replace(string(xdata), "WXOrderParam", "xml", -1)
	fmt.Println("request: ", xmldata)

	body := bytes.NewBufferString(xmldata)
	resp, err := http.Post(WXOrderURL, "content-type:text/xml; charset=utf-8", body)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	var reply WXOrderReply
	err = xml.Unmarshal(result, &reply)
	if err != nil {
		fmt.Println(err)
		return "", ""
	}
	fmt.Println(reply)

	if reply.ReturnCode == "SUCCESS" && reply.ResultCode == "SUCCESS" {
		return reply.PrepayID, reply.MwebURL
	}
	return "", ""
}
